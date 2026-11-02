package obialign

// BandedLCSBuffer is a working memory that can be reused from one alignment to
// the next to avoid any allocation. It must not be shared between goroutines.
type BandedLCSBuffer struct {
	ws bandedLCSWorkspace
}

// AtLeast is BandedLCSAtLeast on the buffer.
func (b *BandedLCSBuffer) AtLeast(seqA, seqB []byte, minLCS int) (lcs, alilen int, err error) {
	return atLeast(seqA, seqB, minLCS, &b.ws)
}

func atLeast(seqA, seqB []byte, minLCS int, ws *bandedLCSWorkspace) (lcs, alilen int, err error) {
	if len(seqA) > MaxBandedLCSLength || len(seqB) > MaxBandedLCSLength {
		return -1, -1, ErrSequenceTooLong
	}
	if len(seqA) < len(seqB) {
		seqA, seqB = seqB, seqA
	}
	lcs, alilen = bandedLCSAtLeast(seqA, seqB, minLCS, ws, haveSIMD)
	return lcs, alilen, nil
}

// BandedLCSAtLeast computes the LCS score of two nucleotide sequences and the
// length of the shortest alignment achieving it, as BandedLCS does, but with
// the bound expressed on the score: the result is the exact optimum when its
// score is at least minLCS, and (-1, -1) otherwise.
//
// The DP only visits the diagonals an alignment of score minLCS can use (one
// whose path makes at most lB-minLCS gaps on one side and lA-minLCS on the
// other), so the cost decreases as minLCS grows. This is exact: every alignment
// of score >= minLCS lies in that band, hence so does the optimum when it
// reaches minLCS, and when the best path of the band is below minLCS so is
// the global optimum.
func BandedLCSAtLeast(seqA, seqB []byte, minLCS int) (lcs, alilen int, err error) {
	ws := bandedLCSPool.Get().(*bandedLCSWorkspace)
	lcs, alilen, err = atLeast(seqA, seqB, minLCS, ws)
	bandedLCSPool.Put(ws)
	return lcs, alilen, err
}

// bandedLCSAtLeast expects len(a) >= len(b), both within MaxBandedLCSLength.
func bandedLCSAtLeast(a, b []byte, minLCS int, ws *bandedLCSWorkspace, useSIMD bool) (int, int) {
	lA, lB := len(a), len(b)
	minLCS = max(minLCS, 0)
	if minLCS > lB {
		return -1, -1
	}
	if lB == 0 {
		return 0, lA
	}
	simdOK := useSIMD && allRegularIUPAC(a) && allRegularIUPAC(b)
	kmin, kmax := -(lB - minLCS), lA-minLCS

	if !simdOK {
		lcs, alilen := ws.align(a, b, kmin, kmax, false)
		if lcs < minLCS {
			return -1, -1
		}
		return lcs, alilen
	}

	// The score alone is much cheaper than the score with the length of the
	// alignment: the latter is only computed when the pair reaches minLCS. With
	// no minimum there is nothing to reject, and the pass would be wasted.
	if minLCS > 0 && lcsScoreSIMD(a, b, kmin, kmax, minLCS, ws) < minLCS {
		return -1, -1
	}
	encodeBases(ws, a, b)
	return ws.align(a, b, kmin, kmax, true)
}

// ScoreAtLeast computes the LCS score of two nucleotide sequences when it is
// at least minLCS, and returns -1 otherwise. It is much cheaper than AtLeast,
// which also computes the length of the shortest alignment.
func (b *BandedLCSBuffer) ScoreAtLeast(seqA, seqB []byte, minLCS int) (int, error) {
	if len(seqA) > MaxBandedLCSLength || len(seqB) > MaxBandedLCSLength {
		return -1, ErrSequenceTooLong
	}
	if len(seqA) < len(seqB) {
		seqA, seqB = seqB, seqA
	}
	lA, lB := len(seqA), len(seqB)
	minLCS = max(minLCS, 0)
	if minLCS > lB {
		return -1, nil
	}
	if lB == 0 {
		return 0, nil
	}
	kmin, kmax := -(lB - minLCS), lA-minLCS
	var lcs int
	if haveSIMD && allRegularIUPAC(seqA) && allRegularIUPAC(seqB) {
		lcs = lcsScoreSIMD(seqA, seqB, kmin, kmax, minLCS, &b.ws)
	} else {
		lcs, _ = b.ws.align(seqA, seqB, kmin, kmax, false)
	}
	if lcs < minLCS {
		return -1, nil
	}
	return lcs, nil
}
