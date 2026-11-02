// Package obisimilarity holds what the tools that compare nucleotide sequences
// by the length of their longest common subsequence (obiclust, obialign) have
// in common: the scoring of a pair against a threshold, with the lossless
// 4-mer filter that avoids most of the alignments, and the index that finds
// the sequences a given one can be similar to.
//
// The scoring follows the arithmetic of sumaclust and sumatra (integer
// truncations, floating point comparisons) so that the same pairs are
// accepted with the same scores.
package obisimilarity

// Table4 is the 4-mer occurrence table of a sequence: one saturating counter
// per word, and the number of occurrences that did not fit in the counters.
type Table4 struct {
	Count    [256]uint8
	Overflow int
}

// NewTable4 counts the 4-mers of a sequence made of the letters a, c, g
// and t. The 2-bit code of a base is (letter>>1)&3, i.e. a=0, c=1, t=2, g=3.
func NewTable4(seq []byte) *Table4 {
	t := new(Table4)
	t.Fill(seq)
	return t
}

// Fill counts the 4-mers of seq in a zeroed table.
func (t *Table4) Fill(seq []byte) {
	if len(seq) < 4 {
		return
	}
	w := uint(0)
	for i, c := range seq {
		w = (w<<2 | uint(c>>1)&3) & 0xFF
		if i < 3 {
			continue
		}
		if t.Count[w] < 255 {
			t.Count[w]++
		} else {
			t.Overflow++
		}
	}
}

// minCommonWords is the minimum number of 4-mers that two sequences must
// share when the longest has length reflen and they have a common
// subsequence of length lcs.
func minCommonWords(reflen, lcs int) int {
	nbfrag := (reflen-lcs)*2 + 1
	smin := lcs / nbfrag
	r := lcs - smin*nbfrag
	return max(smin-2, 0)*r + max(smin-3, 0)*(nbfrag-r)
}
