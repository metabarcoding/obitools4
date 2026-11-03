package obiprimer

import (
	"math/bits"
	"slices"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// seedIndex maps the value of one half of a window to the patterns that are
// candidates for it: those whose same half is within d substitutions of it.
//
// Pigeonhole: a window with at most e = 2d+1 errors against a pattern has at
// most d errors in at least one of the two halves of the pattern, so the
// windows to verify are exactly those found through one of the two indexes.
// Indexing the half of every pattern together with its neighbours at distance
// d (1 + 3L values for d = 1, 1 + 3L + 9L(L-1)/2 for d = 2, L = 9 bases) costs
// a few tens of entries per pattern, and a window of a random text finds about
// patterns x neighbours / 4^L candidates, far fewer than the patterns.
type seedIndex struct {
	shift   uint     // slot = seed >> shift when hashed, else seed
	hashed  bool     // the seed space is too large for a direct table
	offsets []uint32 // CSR: ids[offsets[s]:offsets[s+1]] are the candidates of slot s
	ids     []int32  // pattern ids (2*word + strand)
}

const maxDirectSeedBits = 22

func (x *seedIndex) slot(seed uint64) uint32 {
	if !x.hashed {
		return uint32(seed)
	}
	return uint32((seed * 0x9E3779B97F4A7C15) >> x.shift)
}

func (x *seedIndex) candidates(seed uint64) []int32 {
	s := x.slot(seed)
	return x.ids[x.offsets[s]:x.offsets[s+1]]
}

// neighbours calls fn for v and every value within d substitutions of it (each
// once), v being L bases packed on 2 bits.
func neighbours(v uint64, L, d int, fn func(uint64)) {
	var rec func(v uint64, from, left int)
	rec = func(v uint64, from, left int) {
		fn(v)
		if left == 0 {
			return
		}
		for pos := from; pos < L; pos++ {
			for alt := uint64(1); alt <= 3; alt++ {
				rec(v^(alt<<(2*uint(pos))), pos+1, left-1)
			}
		}
	}
	rec(v, 0, d)
}

// newSeedIndex indexes, for every pattern, the half of the given width (in
// bases) selected by (code >> shiftBases*2) & mask, with its neighbours at
// distance d.
func newSeedIndex(patterns []uint64, shiftBases, width, d int) *seedIndex {
	x := &seedIndex{}
	bitsNeeded := 2 * width
	if bitsNeeded > maxDirectSeedBits {
		x.hashed, bitsNeeded = true, maxDirectSeedBits
		x.shift = uint(64 - maxDirectSeedBits)
	}
	nslots := 1 << uint(bitsNeeded)
	halfMask := uint64(1)<<(2*uint(width)) - 1

	type entry struct {
		slot uint32
		id   int32
	}
	var entries []entry
	var scratch []uint32
	for id, pat := range patterns {
		half := (pat >> (2 * uint(shiftBases))) & halfMask
		scratch = scratch[:0]
		neighbours(half, width, d, func(v uint64) { scratch = append(scratch, x.slot(v)) })
		slices.Sort(scratch)
		scratch = slices.Compact(scratch) // hashing may merge neighbours
		for _, s := range scratch {
			entries = append(entries, entry{s, int32(id)})
		}
	}

	x.offsets = make([]uint32, nslots+1)
	for _, e := range entries {
		x.offsets[e.slot+1]++
	}
	for i := 1; i <= nslots; i++ {
		x.offsets[i] += x.offsets[i-1]
	}
	x.ids = make([]int32, len(entries))
	fill := slices.Clone(x.offsets[:nslots])
	for _, e := range entries {
		x.ids[fill[e.slot]] = e.id
		fill[e.slot]++
	}
	return x
}

// MatchWords finds every approximate occurrence, on both strands, of the given
// canonical words in the sequences of the corpus, with the rules of MatchParams
// (substitutions only, ambiguous bases count as errors, circular junction).
//
// It uses the pigeonhole seed indexes above. It returns exactly what the
// reference matchNaive returns, and what ecoPrimers' fuzzy search finds, in a
// small fraction of the time: on the Teleostei tutorial data (255 words, 3909
// sequences, 65 Mb, e = 3) 0.42 s on one core, against 113 s for the
// obiapat (agrep) search it replaces.
func MatchWords(c *Corpus, words []uint64, p MatchParams) ([]WordMatches, error) {
	if err := p.validate(); err != nil {
		return nil, err
	}
	k := p.K
	d := p.MaxErrors / 2

	// pattern 2w is the word, 2w+1 its reverse complement
	patterns := make([]uint64, 2*len(words))
	for i, w := range words {
		patterns[2*i] = codeOf(wordCodes(w, k, false), k)
		patterns[2*i+1] = codeOf(wordCodes(w, k, true), k)
	}

	right := k / 2 // the right half holds the low bits
	left := k - right
	leftIdx := newSeedIndex(patterns, right, left, d)
	rightIdx := newSeedIndex(patterns, 0, right, d)

	rightMask := uint64(1)<<(2*uint(right)) - 1
	windowMask := uint64(1)<<(2*uint(k)) - 1
	lowBits := uint64(0x5555555555555555) & windowMask
	maxErr := p.MaxErrors

	mk := func() scanFunc {
		return func(seq *obiseq.BioSequence, emit func(word, pos, errors int, reverse bool)) {
			raw := seq.Sequence()
			n := len(raw)
			if n < k {
				return
			}
			total := n // number of bases fed to the rolling window
			windows := n - k + 1
			if p.Circular {
				total, windows = n+k-1, n
			}

			var code, bad uint64 // window, and its non-ACGT positions (bit 2j)
			verify := func(pos int, id int32) {
				pat := patterns[id]
				x := code ^ pat
				if errs := bits.OnesCount64(((x | x>>1) & lowBits) | bad); errs <= maxErr {
					emit(int(id)>>1, pos, errs, id&1 == 1)
				}
			}

			for i := 0; i < total; i++ {
				j := i
				if j >= n {
					j -= n
				}
				v := baseCode[raw[j]]
				isBad := uint64(0)
				if v > 3 {
					v, isBad = 0, 1
				}
				code = (code<<2 | uint64(v)) & windowMask
				bad = (bad<<2 | isBad) & windowMask

				if i < k-1 {
					continue
				}
				pos := i - k + 1
				if pos >= windows {
					break
				}
				leftCands := leftIdx.candidates(code >> (2 * uint(right)))
				for _, id := range leftCands {
					verify(pos, id)
				}
				for _, id := range rightIdx.candidates(code & rightMask) {
					// ids are sorted: skip the ones the left index has just verified
					if _, done := slices.BinarySearch(leftCands, id); !done {
						verify(pos, id)
					}
				}
			}
		}
	}
	return matchCorpus(c, words, mk), nil
}

// wordCodes returns the 2-bit codes (A=0..T=3) of the word or of its reverse
// complement, 5' to 3'.
func wordCodes(word uint64, k int, reverse bool) []byte {
	fwd, rev := wordStrings([]uint64{word}, k)
	s := fwd[0]
	if reverse {
		s = rev[0]
	}
	return codesOf(s)
}

// codeOf packs codes (each 0..3) on 2 bits, first base in the high bits.
func codeOf(codes []byte, k int) uint64 {
	var v uint64
	for _, c := range codes[:k] {
		v = v<<2 | uint64(c)
	}
	return v
}
