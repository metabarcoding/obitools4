package obiprimer

import (
	"cmp"
	"fmt"
	"slices"
	"sort"
	"strings"
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
)

// PairParams holds the parameters of the primer pair search, named after the
// ecoPrimers options.
type PairParams struct {
	PrimerLength int // -O, word size
	MinLength    int // -l: the amplicon (primers excluded) must be longer than this
	MaxLength    int // -L: and shorter than this
	// MinLengthNeg and MaxLengthNeg bound the amplicons looked for in the
	// counterexamples, where any amplification is unwanted, whatever its size:
	// MinLengthNeg 0 means no lower bound (the primer sites cannot overlap, so
	// an amplicon is at least 1 long), and a MaxLengthNeg lower than MaxLength
	// (0 included) is raised to MaxLength.
	MinLengthNeg   int
	MaxLengthNeg   int
	ThreePrime     int     // -3: number of 3' bases that must match exactly (0 = no constraint)
	Sensitivity    float64 // -s: a primer must hit more than this fraction of the examples (0.9)
	FalsePositive  float64 // -x: and less than this fraction of the counterexamples (0.1)
	MinSpecificity float64 // -T: pairs with a lower Bs are dropped (0.6)
	Circular       bool    // -c
	Reference      int     // -R: corpus index of the reference sequence (used when HasReference)
	HasReference   bool
	AmpBudget      int    // amplicon records held at once in phase 2 (0 = DefaultAmpBudget)
	Release        func() // see SearchParams.Release
}

// window returns the exclusive length bounds of the amplicons looked for in a
// sequence of the given role.
func (p PairParams) window(role Role) (minLen, maxLen int) {
	if role == CounterExample {
		return p.MinLengthNeg, max(p.MaxLengthNeg, p.MaxLength)
	}
	return p.MinLength, p.MaxLength
}

func (p PairParams) validate() error {
	switch {
	case p.PrimerLength < 2 || p.PrimerLength > 31:
		return fmt.Errorf("obiprimer: primer length must be between 2 and 31, got %d", p.PrimerLength)
	case p.MaxLength <= p.MinLength+1:
		return fmt.Errorf("obiprimer: empty amplicon length range ]%d,%d[", p.MinLength, p.MaxLength)
	case p.ThreePrime < 0 || p.ThreePrime > p.PrimerLength:
		return fmt.Errorf("obiprimer: 3' strict bases must be in [0,%d], got %d", p.PrimerLength, p.ThreePrime)
	case p.Sensitivity <= 0 || p.Sensitivity > 1 || p.FalsePositive < 0 || p.FalsePositive > 1:
		return fmt.Errorf("obiprimer: quorums must be in ]0,1] and [0,1]")
	}
	return nil
}

// Primer is a conserved word: one found, with at most the allowed errors, in
// enough example sequences to be a candidate primer.
type Primer struct {
	Word            uint64 // canonical word
	Examples        int    // example sequences where it is found (either strand)
	CounterExamples int
	// Good means conserved in the examples and rare enough in the
	// counterexamples; a pair needs at least one good primer.
	Good bool
}

// seqMatch is one occurrence of a primer in one sequence, on 8 bytes: there
// are millions of them.
type seqMatch struct {
	primer int32
	place  uint32 // position << 1, plus 1 when the primer is found reversed
}

func newSeqMatch(primer int32, pos uint32, reverse bool) seqMatch {
	return seqMatch{primer, pos<<1 | uint32(b2i(reverse))}
}

func (m seqMatch) pos() uint32   { return m.place >> 1 }
func (m seqMatch) reverse() bool { return m.place&1 == 1 }

// primerSet is the conserved words and, for each corpus sequence, their
// occurrences sorted by position.
type primerSet struct {
	primers []Primer
	perSeq  [][]seqMatch
}

// quorumCount is floor(n*fraction), robust to the binary form of the fraction.
func quorumCount(n int, fraction float64) int { return MinSequences(n, fraction) }

// selectPrimers keeps the words found in at least floor(examples*Sensitivity)
// example sequences and marks as good those found in at most
// floor(counterexamples*FalsePositive) counterexamples.
func selectPrimers(c *Corpus, wm []WordMatches, p PairParams) *primerSet {
	counts := c.Counts()
	inQ := quorumCount(counts.Examples, p.Sensitivity)
	outQ := quorumCount(counts.CounterExamples, p.FalsePositive)

	ex, co := make([]int, len(wm)), make([]int, len(wm))
	parallelFor(len(wm), func(_, i int) { ex[i], co[i] = wm[i].SequencesMatched(c) })

	// the occurrences of the words kept, by sequence, each in a slice of the
	// exact size (and not in one big array: a few hundred bytes to some kB each
	// find room in the free memory, which a single block of tens of MB does not)
	ps := &primerSet{perSeq: make([][]seqMatch, c.Len())}
	size := make([]int, c.Len())
	for i, w := range wm {
		if ex[i] < inQ {
			continue
		}
		ps.primers = append(ps.primers, Primer{w.Word, ex[i], co[i], co[i] <= outQ})
		for _, m := range w.Matches {
			size[m.Seq]++
		}
	}
	for s, n := range size {
		if n > 0 {
			ps.perSeq[s] = make([]seqMatch, 0, n)
		}
	}
	id := int32(0)
	for i, w := range wm {
		if ex[i] < inQ {
			continue
		}
		for _, m := range w.Matches {
			ps.perSeq[m.Seq] = append(ps.perSeq[m.Seq], newSeqMatch(id, m.Pos, m.Reverse))
		}
		id++
	}
	parallelFor(c.Len(), func(_, s int) {
		slices.SortStableFunc(ps.perSeq[s], func(a, b seqMatch) int { return cmp.Compare(a.pos(), b.pos()) })
	})
	return ps
}

// pairKey identifies a pair: the two primers, ordered by canonical word, and
// for each whether the primer is the canonical word as is (true) or its reverse
// complement (false). It is the key of ecoPrimers' pair tree.
type pairKey struct {
	p1, p2 int32
	a1, a2 bool
}

// occurrence is one amplicon: a pair of matches in a sequence.
type occurrence struct {
	key      pairKey
	swapped  bool // true when the upstream match is the second primer of the key
	up, down seqMatch
	length   int // amplicon length, primers excluded
}

// forEachOccurrence calls fn for every amplicon of sequence s, applying the
// length, good-primer and 3' constraints of ecoPrimers. buf is the buffer in
// which the sequence is unpacked, if the corpus is packed.
func (ps *primerSet) forEachOccurrence(c *Corpus, s int, buf []byte, p PairParams, fn func(occurrence)) []byte {
	minLen, maxLen := p.window(c.Roles[s])
	ms := ps.perSeq[s]
	var raw []byte
	if p.ThreePrime > 0 && len(ms) > 1 {
		raw, buf = c.bases(s, buf)
	}
	ps.forEachOccurrenceIn(raw, ms, p, minLen, maxLen, fn)
	return buf
}

// forEachOccurrenceIn is forEachOccurrence over ms, the matches of a sequence
// sorted by position (all of them, or only those of the primers of interest,
// which yields the amplicons between these primers and none other), and raw its
// bases, needed when the 3' bases must match exactly. Only the amplicons whose
// length is strictly between minLen and maxLen are reported.
func (ps *primerSet) forEachOccurrenceIn(raw []byte, ms []seqMatch, p PairParams, minLen, maxLen int, fn func(occurrence)) {
	lp := p.PrimerLength

	for i := range ms {
		for j := i + 1; j < len(ms); j++ {
			dpos := int(ms[j].pos()) - int(ms[i].pos())
			if dpos <= lp {
				continue
			}
			length := dpos - lp
			if length >= maxLen {
				break
			}
			up, down := ms[i], ms[j]
			if length <= minLen || !(ps.primers[up.primer].Good || ps.primers[down.primer].Good) {
				continue
			}

			// The upstream match is read as is (primer 1 of the pair when it
			// comes first), the downstream one as the reverse primer. The same
			// amplicon read on the other strand is the pair (R, F) instead of
			// (F, R): the primers are ordered by (word, direct flag), false
			// before true, so that both give the same key. Ordering by the word
			// alone left the two views of a pair of one word, one primer being
			// the reverse complement of the other, as two different keys.
			key := pairKey{up.primer, down.primer, !up.reverse(), down.reverse()}
			w1, w2 := ps.primers[key.p1].Word, ps.primers[key.p2].Word
			swapped := w2 < w1 || (w2 == w1 && key.a1 && !key.a2)
			if swapped {
				key = pairKey{key.p2, key.p1, key.a2, key.a1}
			}

			if p.ThreePrime > 0 && !ps.threePrimeMatches(raw, up, down, p) {
				continue
			}
			fn(occurrence{key, swapped, up, down, length})
		}
	}
}

// wordBase returns base t (0 = 5') of a 2-bit packed word of size lp.
func wordBase(word uint64, lp, t int) byte { return byte(word>>(2*uint(lp-1-t))) & 3 }

// threePrimeMatches reports whether the last p.ThreePrime bases of the forward
// primer (as found at the upstream match) and of the reverse primer (reverse
// complement of the downstream window) equal the sequence, i.e. have no
// mismatch at their 3' end.
func (ps *primerSet) threePrimeMatches(raw []byte, up, down seqMatch, p PairParams) bool {
	lp, n := p.PrimerLength, len(raw)
	at := func(pos int) byte {
		if p.Circular {
			pos %= n
		}
		return baseCode[raw[pos]]
	}

	fw := ps.primers[up.primer].Word
	rv := ps.primers[down.primer].Word
	for t := lp - p.ThreePrime; t < lp; t++ {
		// forward primer: the word when found direct, its reverse complement otherwise
		var f byte
		if !up.reverse() {
			f = wordBase(fw, lp, t)
		} else {
			f = 3 - wordBase(fw, lp, lp-1-t)
		}
		if s := at(int(up.pos()) + t); s > 3 || s != f {
			return false
		}

		// reverse primer: the word when found reverse, its reverse complement otherwise
		var r byte
		if down.reverse() {
			r = wordBase(rv, lp, t)
		} else {
			r = 3 - wordBase(rv, lp, lp-1-t)
		}
		if s := at(int(down.pos()) + lp - 1 - t); s > 3 || 3-s != r {
			return false
		}
	}
	return true
}

// pairAcc accumulates, over the sequences, the counts ecoPrimers keeps while it
// builds the pairs.
type pairAcc struct {
	examples, counters uint32
	last               int32 // 1 + last sequence counted, to count each sequence once
	amplicons          uint32
	sumLen             uint64
	minLen, maxLen     uint32
}

// sensitivityNeeded is the smallest number of examples, out of e, that a pair
// must amplify to pass the sensitivity test q > sens (the test of FindPairs,
// in float32 as there). It is e+1 when no number does.
func sensitivityNeeded(e int, sens float32) int {
	return sort.Search(e+1, func(n int) bool { return float32(n)/float32(e) > sens })
}

// countPairs enumerates the amplicons of the sequences, in parallel, and
// returns for each pair the sequences it amplifies and the length statistics
// of its example amplicons.
//
// Only the pairs that can pass the sensitivity test are returned, which is what
// FindPairs keeps: a pair passing it misses at most M = E - needed examples,
// so it is found in at least one of any M+1 examples. The pairs are therefore
// registered from M+1 examples only (those with the fewest primer sites, to
// keep the set small), and the other examples and the counterexamples are only
// used to count the amplicons of the pairs already registered. Registering the
// pairs of every sequence would build millions of pairs, nearly all of them
// found in a counterexample alone, to drop them at once. The counts of a
// retained pair are those of an exhaustive enumeration.
func (ps *primerSet) countPairs(c *Corpus, p PairParams) map[pairKey]*pairAcc {
	counts := c.Counts()
	if counts.Examples == 0 {
		return map[pairKey]*pairAcc{}
	}
	needed := sensitivityNeeded(counts.Examples, float32(p.Sensitivity))
	nFirst := counts.Examples - needed + 1 // examples a pair may miss, plus one
	if nFirst <= 0 {
		return map[pairKey]*pairAcc{} // no pair can amplify enough examples
	}

	var examples, counters []int
	for s, r := range c.Roles {
		switch r {
		case Example:
			examples = append(examples, s)
		case CounterExample:
			counters = append(counters, s)
		}
	}
	slices.SortStableFunc(examples, func(a, b int) int { return cmp.Compare(len(ps.perSeq[a]), len(ps.perSeq[b])) })
	nFirst = min(nFirst, len(examples))

	cand := ps.scanPairs(c, p, examples[:nFirst], nil, "Registering candidate pairs")
	rest := append(slices.Clone(examples[nFirst:]), counters...)
	more := ps.scanPairs(c, p, rest, cand, "Counting candidate pairs")
	for k, a := range more {
		mergeAcc(cand[k], a)
	}
	return cand
}

// scanPairs enumerates the amplicons of the sequences seqs and accumulates them
// by pair. With a non-nil known, only the amplicons of the pairs of known are
// counted (and only their primers are looked at); the result then holds a
// subset of the keys of known, to be merged into it.
func (ps *primerSet) scanPairs(c *Corpus, p PairParams, seqs []int, known map[pairKey]*pairAcc, label string) map[pairKey]*pairAcc {
	workers := obidefault.ParallelWorkers()
	partial := make([]map[pairKey]*pairAcc, workers)

	var wanted []bool
	if known != nil {
		wanted = make([]bool, len(ps.primers))
		for k := range known {
			wanted[k.p1], wanted[k.p2] = true, true
		}
	}

	bar := newStageBar(label, len(seqs))
	defer bar.Done()
	jobs := make(chan int)
	var wg sync.WaitGroup
	for w := range workers {
		wg.Add(1)
		partial[w] = map[pairKey]*pairAcc{}
		go func() {
			defer wg.Done()
			acc := partial[w]
			var buf []byte
			var sites []seqMatch
			for s := range jobs {
				bar.Add(1)
				isExample := c.Roles[s] == Example
				minLen, maxLen := p.window(c.Roles[s])
				ms := ps.perSeq[s]
				if wanted != nil {
					sites = sites[:0]
					for _, m := range ms {
						if wanted[m.primer] {
							sites = append(sites, m)
						}
					}
					ms = sites
				}
				if len(ms) < 2 {
					continue // an amplicon needs two primer sites
				}
				var raw []byte
				if p.ThreePrime > 0 {
					raw, buf = c.bases(s, buf)
				}
				ps.forEachOccurrenceIn(raw, ms, p, minLen, maxLen, func(o occurrence) {
					if known != nil {
						if _, ok := known[o.key]; !ok {
							return
						}
					}
					a := acc[o.key]
					if a == nil {
						a = &pairAcc{minLen: ^uint32(0)}
						acc[o.key] = a
					}
					if a.last != int32(s)+1 {
						a.last = int32(s) + 1
						if isExample {
							a.examples++
						} else {
							a.counters++
						}
					}
					if isExample {
						l := uint32(o.length)
						a.amplicons++
						a.sumLen += uint64(l)
						a.minLen = min(a.minLen, l)
						a.maxLen = max(a.maxLen, l)
					}
				})
			}
		}()
	}
	for _, s := range seqs {
		jobs <- s
	}
	close(jobs)
	wg.Wait()

	// A sequence is handled by one worker only, so the sequence counts add up.
	total := partial[0]
	for _, m := range partial[1:] {
		for k, a := range m {
			if t := total[k]; t == nil {
				total[k] = a
			} else {
				mergeAcc(t, a)
			}
		}
	}
	return total
}

// mergeAcc adds the counts of src, found in other sequences, to dst.
func mergeAcc(dst, src *pairAcc) {
	dst.examples += src.examples
	dst.counters += src.counters
	dst.amplicons += src.amplicons
	dst.sumLen += src.sumLen
	dst.minLen = min(dst.minLen, src.minLen)
	dst.maxLen = max(dst.maxLen, src.maxLen)
}

// orientedWord returns the upper-case primer string 5'->3': the canonical word
// when direct, its reverse complement otherwise.
func orientedWord(word uint64, lp int, direct bool) string {
	if !direct {
		word = obikmer.ReverseComplement(word, lp)
	}
	return strings.ToUpper(string(obikmer.DecodeKmer(word, lp, nil)))
}
