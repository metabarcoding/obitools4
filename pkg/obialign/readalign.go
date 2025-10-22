package obialign

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

func ReadAlign(seqA, seqB *obiseq.BioSequence,
	gap, scale float64, delta int, fastScoreRel bool,
	arena PEAlignArena, shift_buff *map[int]int) (int, []int, int, int, float64, bool) {
	var score, shift int
	var startA, startB int
	var partLen, over int
	var rawSeqA, qualSeqA []byte
	var rawSeqB, qualSeqB []byte
	var extra5, extra3 int

	var path []int

	if !_InitializedDnaScore {
		_InitDNAScoreMatrix()
	}

	fastCount := -1
	fastScore := -1.0

	directAlignment := true

	index := obikmer.Index4mer(seqA,
		&arena.pointer.fastIndex,
		&arena.pointer.fastBuffer)

	shift, fastCount, fastScore = obikmer.FastShiftFourMer(index, shift_buff, seqA.Len(), seqB, fastScoreRel, nil)

	seqBR := seqB.ReverseComplement(false)
	shiftR, fastCountR, fastScoreR := obikmer.FastShiftFourMer(index, shift_buff, seqA.Len(), seqBR, fastScoreRel, nil)

	if fastCount < fastCountR {
		shift = shiftR
		fastCount = fastCountR
		fastScore = fastScoreR
		seqB = seqBR
		directAlignment = false
	}

	// Compute the overlapping region length
	switch {
	case shift > 0:
		over = seqA.Len() - shift
	case shift < 0:
		over = seqB.Len() + shift
	default:
		over = min(seqA.Len(), seqB.Len())
	}

	// obilog.Warnf("fw/fw: %v shift=%d fastCount=%d/over=%d fastScore=%f",
	// 	directAlignment, shift, fastCount, over, fastScore)

	// obilog.Warnf(("seqA: %s\nseqB: %s\n"), seqA.String(), seqB.String())

	// At least one mismatch exists in the overlaping region
	if fastCount+3 < over {

		if shift > 0 || (shift == 0 && seqB.Len() >= seqA.Len()) {
			startA = shift - delta
			if startA < 0 {
				startA = 0
			}
			extra5 = -startA
			startB = 0

			rawSeqA = seqA.Sequence()[startA:]
			qualSeqA = seqA.Qualities()[startA:]
			partLen = len(rawSeqA)
			if partLen > seqB.Len() {
				partLen = seqB.Len()
			}
			rawSeqB = seqB.Sequence()[0:partLen]
			qualSeqB = seqB.Qualities()[0:partLen]
			extra3 = seqB.Len() - partLen
			score = _FillMatrixPeLeftAlign(
				rawSeqA, qualSeqA, rawSeqB, qualSeqB, gap, scale,
				&arena.pointer.scoreMatrix,
				&arena.pointer.pathMatrix)
		} else {

			startA = 0
			startB = -shift - delta
			if startB < 0 {
				startB = 0
			}
			extra5 = startB
			rawSeqB = seqB.Sequence()[startB:]
			qualSeqB = seqB.Qualities()[startB:]
			partLen = len(rawSeqB)
			if partLen > seqA.Len() {
				partLen = seqA.Len()
			}
			rawSeqA = seqA.Sequence()[:partLen]
			qualSeqA = seqA.Qualities()[:partLen]
			extra3 = partLen - seqA.Len()

			score = _FillMatrixPeRightAlign(
				rawSeqA, qualSeqA, rawSeqB, qualSeqB, gap, scale,
				&arena.pointer.scoreMatrix,
				&arena.pointer.pathMatrix)
		}

		path = _Backtracking(arena.pointer.pathMatrix,
			len(rawSeqA), len(rawSeqB),
			&arena.pointer.path)

	} else {

		// Both overlaping regions are identicals

		if shift > 0 || (shift == 0 && seqB.Len() >= seqA.Len()) {
			startA = shift
			startB = 0
			extra5 = -startA
			qualSeqA = seqA.Qualities()[startA:]
			partLen = len(qualSeqA)
			qualSeqB = seqB.Qualities()[0:partLen]
			extra3 = seqB.Len() - partLen
			score = 0
		} else {
			startA = 0
			startB = -shift
			extra5 = startB
			qualSeqB = seqB.Qualities()[startB:]
			partLen = len(qualSeqB)
			extra3 = partLen - seqA.Len()
			qualSeqA = seqA.Qualities()[:partLen]
		}

		score = 0
		for i, qualA := range qualSeqA {
			qualB := qualSeqB[i]
			score += _NucScorePartMatchMatch[qualA][qualB]
		}

		path = arena.pointer.path[:0]
		path = append(path, 0, partLen)
	}

	path[0] += extra5
	if path[len(path)-1] == 0 {
		path[len(path)-2] += extra3
	} else {
		path = append(path, extra3, 0)
	}

	return score, path, fastCount, over, fastScore, directAlignment
}
