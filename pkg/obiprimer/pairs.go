package obiprimer

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
)

// PairParams holds the parameters of the primer pair search, named after the
// ecoPrimers options.
type PairParams struct {
	PrimerLength   int     // -O, word size
	MinLength      int     // -l: the amplicon (primers excluded) must be longer than this
	MaxLength      int     // -L: and shorter than this
	ThreePrime     int     // -3: number of 3' bases that must match exactly (0 = no constraint)
	Sensitivity    float64 // -s: a primer must hit more than this fraction of the examples (0.9)
	FalsePositive  float64 // -x: and less than this fraction of the counterexamples (0.1)
	MinSpecificity float64 // -T: pairs with a lower Bs are dropped (0.6)
	Circular       bool    // -c
	Reference      int     // -R: corpus index of the reference sequence (used when HasReference)
	HasReference   bool
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

// seqMatch is one occurrence of a primer in one sequence.
type seqMatch struct {
	primer  int32
	pos     uint32
	reverse bool
}

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

	ps := &primerSet{perSeq: make([][]seqMatch, len(c.Seqs))}
	for _, w := range wm {
		ex, co := w.SequencesMatched(c)
		if ex < inQ {
			continue
		}
		id := int32(len(ps.primers))
		ps.primers = append(ps.primers, Primer{w.Word, ex, co, co <= outQ})
		for _, m := range w.Matches {
			ps.perSeq[m.Seq] = append(ps.perSeq[m.Seq], seqMatch{id, m.Pos, m.Reverse})
		}
	}
	for _, ms := range ps.perSeq {
		slices.SortStableFunc(ms, func(a, b seqMatch) int { return cmp.Compare(a.pos, b.pos) })
	}
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
// length, good-primer and 3' constraints of ecoPrimers.
func (ps *primerSet) forEachOccurrence(c *Corpus, s int, p PairParams, fn func(occurrence)) {
	ms := ps.perSeq[s]
	lp := p.PrimerLength
	var raw []byte
	if p.ThreePrime > 0 {
		raw = c.Seqs[s].Sequence()
	}

	for i := range ms {
		for j := i + 1; j < len(ms); j++ {
			dpos := int(ms[j].pos) - int(ms[i].pos)
			if dpos <= lp {
				continue
			}
			length := dpos - lp
			if length >= p.MaxLength {
				break
			}
			up, down := ms[i], ms[j]
			if length <= p.MinLength || !(ps.primers[up.primer].Good || ps.primers[down.primer].Good) {
				continue
			}

			// The upstream match is read as is (primer 1 of the pair when its word
			// is smaller), the downstream one as the reverse primer.
			key := pairKey{up.primer, down.primer, !up.reverse, down.reverse}
			swapped := !(ps.primers[key.p2].Word > ps.primers[key.p1].Word)
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
		if !up.reverse {
			f = wordBase(fw, lp, t)
		} else {
			f = 3 - wordBase(fw, lp, lp-1-t)
		}
		if s := at(int(up.pos) + t); s > 3 || s != f {
			return false
		}

		// reverse primer: the word when found reverse, its reverse complement otherwise
		var r byte
		if down.reverse {
			r = wordBase(rv, lp, t)
		} else {
			r = 3 - wordBase(rv, lp, lp-1-t)
		}
		if s := at(int(down.pos) + lp - 1 - t); s > 3 || 3-s != r {
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

// countPairs enumerates the amplicons of every sequence, in parallel, and
// returns for each pair the sequences it amplifies and the length statistics
// of its example amplicons.
func (ps *primerSet) countPairs(c *Corpus, p PairParams) map[pairKey]*pairAcc {
	workers := obidefault.ParallelWorkers()
	partial := make([]map[pairKey]*pairAcc, workers)

	jobs := make(chan int)
	var wg sync.WaitGroup
	for w := range workers {
		wg.Add(1)
		partial[w] = map[pairKey]*pairAcc{}
		go func() {
			defer wg.Done()
			acc := partial[w]
			for s := range jobs {
				if c.Roles[s] == Other {
					continue // only the quorums are counted here
				}
				isExample := c.Roles[s] == Example
				ps.forEachOccurrence(c, s, p, func(o occurrence) {
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
	for s := range c.Seqs {
		jobs <- s
	}
	close(jobs)
	wg.Wait()

	// A sequence is handled by one worker only, so the sequence counts add up.
	total := partial[0]
	for _, m := range partial[1:] {
		for k, a := range m {
			t := total[k]
			if t == nil {
				total[k] = a
				continue
			}
			t.examples += a.examples
			t.counters += a.counters
			t.amplicons += a.amplicons
			t.sumLen += a.sumLen
			t.minLen = min(t.minLen, a.minLen)
			t.maxLen = max(t.maxLen, a.maxLen)
		}
	}
	return total
}

// orientedWord returns the upper-case primer string 5'->3': the canonical word
// when direct, its reverse complement otherwise.
func orientedWord(word uint64, lp int, direct bool) string {
	if !direct {
		word = obikmer.ReverseComplement(word, lp)
	}
	return strings.ToUpper(string(obikmer.DecodeKmer(word, lp, nil)))
}
