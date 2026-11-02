package obiclust

// table4 is the 4-mer occurrence table of a sequence, as sumaclust builds it:
// one saturating counter per word, and the number of occurrences that did
// not fit in the counters.
type table4 struct {
	count    [256]uint8
	overflow int
}

// buildTable4 counts the 4-mers of a sequence made of the letters a, c, g
// and t. The 2-bit code of a base is (letter>>1)&3, i.e. a=0, c=1, t=2, g=3,
// which is the encoding of sumaclust.
func buildTable4(seq []byte) *table4 {
	t := new(table4)
	t.fill(seq)
	return t
}

// fill counts the 4-mers of seq in a zeroed table.
func (t *table4) fill(seq []byte) {
	if len(seq) < 4 {
		return
	}
	w := uint(0)
	for i, c := range seq {
		w = (w<<2 | uint(c>>1)&3) & 0xFF
		if i < 3 {
			continue
		}
		if t.count[w] < 255 {
			t.count[w]++
		} else {
			t.overflow++
		}
	}
}

// minCommonWords is the minimum number of 4-mers that two sequences must
// share when the longest has length reflen and they have a common
// subsequence of length lcs (integer arithmetic of sumaclust, thresholdLCS4).
func minCommonWords(reflen, lcs int) int {
	nbfrag := (reflen-lcs)*2 + 1
	smin := lcs / nbfrag
	r := lcs - smin*nbfrag
	return max(smin-2, 0)*r + max(smin-3, 0)*(nbfrag-r)
}
