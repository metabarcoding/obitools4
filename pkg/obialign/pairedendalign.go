package obialign

import (
	"log"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obikmer"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

type _PeAlignArena struct {
	scoreMatrix []int
	pathMatrix  []int
	path        []int
	fastIndex   [][]int
	fastBuffer  []byte
}

// PEAlignArena defines memory arena usable by the
// Paired-End alignment related functions. The same arena can be reused
// from alignment to alignment to limit memory allocation
// and desallocation process.
type PEAlignArena struct {
	pointer *_PeAlignArena
}

// NilPEAlignArena is the nil instance of the PEAlignArena
// type.
var NilPEAlignArena = PEAlignArena{nil}

// MakePEAlignArena makes a new arena for the alignment of two paired sequences
// of maximum length indicated by lseqA and lseqB.
func MakePEAlignArena(lseqA, lseqB int) PEAlignArena {
	a := _PeAlignArena{
		scoreMatrix: make([]int, 0, (lseqA+1)*(lseqB+1)),
		pathMatrix:  make([]int, 0, (lseqA+1)*(lseqB+1)),
		path:        make([]int, 2*(lseqA+lseqB)),
		fastIndex:   make([][]int, 256),
		fastBuffer:  make([]byte, 0, lseqA),
	}

	return PEAlignArena{&a}
}

func _SetMatrices(matrixA, matrixB *[]int, lenA, a, b, valueA, valueB int) {
	i := (b+1)*(lenA+1) + a + 1
	(*matrixA)[i] = valueA
	(*matrixB)[i] = valueB
}

func _GetMatrix(matrix *[]int, lenA, a, b int) int {
	return (*matrix)[(b+1)*(lenA+1)+a+1]
}

func _GetMatrixFrom(matrix *[]int, lenA, a, b int) (int, int, int) {
	i := (b+1)*(lenA+1) + a
	j := i - lenA
	m := *matrix
	return m[j], m[j-1], m[i]
}

func _PairingScorePeAlign(baseA, qualA, baseB, qualB byte) int {
	partMatch := _NucPartMatch[baseA&31][baseB&31]
	// log.Printf("id : %f A : %s %d B : %s %d\n", part_match, string(baseA), qualA, string(baseB), qualB)
	switch int(partMatch * 100) {
	case 100:
		return _NucScorePartMatchMatch[qualA][qualB]
	case 0:
		return _NucScorePartMatchMismatch[qualA][qualB]
	default:
		return int(partMatch*float64(_NucScorePartMatchMatch[qualA][qualB]) +
			(1-partMatch)*float64(_NucScorePartMatchMismatch[qualA][qualB]) + 0.5)
	}
}

func _FillMatrixPeLeftAlign(seqA, qualA, seqB, qualB []byte, gap int,
	scoreMatrix, pathMatrix *[]int) int {

	la := len(seqA)
	lb := len(seqB)

	// The actual gap score is the gap score times the mismatch between
	// two bases with a score of 40
	gap = gap * _NucScorePartMatchMismatch[40][40]

	needed := (la + 1) * (lb + 1)

	if needed > cap(*scoreMatrix) {
		*scoreMatrix = make([]int, needed)
	}

	if needed > cap(*pathMatrix) {
		*pathMatrix = make([]int, needed)
	}

	*scoreMatrix = (*scoreMatrix)[:needed]
	*pathMatrix = (*pathMatrix)[:needed]

	_SetMatrices(scoreMatrix, pathMatrix, la, -1, -1, 0, 0)

	// Fills the first column with score 0
	for i := 0; i < la; i++ {
		_SetMatrices(scoreMatrix, pathMatrix, la, i, -1, 0, -1)
	}

	la1 := la - 1

	for j := 0; j < lb; j++ {

		_SetMatrices(scoreMatrix, pathMatrix, la, -1, j, (j+1)*gap, 1)

		for i := 0; i < la1; i++ {
			left, diag, top := _GetMatrixFrom(scoreMatrix, la, i, j)
			diag += _PairingScorePeAlign(seqA[i], qualA[i], seqB[j], qualB[j])
			left += gap
			top += gap
			switch {
			case diag > left && diag > top:
				_SetMatrices(scoreMatrix, pathMatrix, la, i, j, diag, 0)
			case left > diag && left > top:
				_SetMatrices(scoreMatrix, pathMatrix, la, i, j, left, +1)
			default:
				_SetMatrices(scoreMatrix, pathMatrix, la, i, j, top, -1)
			}
		}

		// Special case for the last line Left gap are free

		left, diag, top := _GetMatrixFrom(scoreMatrix, la, la1, j)
		diag += _PairingScorePeAlign(seqA[la1], qualA[la1], seqB[j], qualB[j])
		top += gap

		switch {
		case diag > left && diag > top:
			_SetMatrices(scoreMatrix, pathMatrix, la, la1, j, diag, 0)
		case left > diag && left > top:
			_SetMatrices(scoreMatrix, pathMatrix, la, la1, j, left, +1)
		default:
			_SetMatrices(scoreMatrix, pathMatrix, la, la1, j, top, -1)

		}
	}

	return _GetMatrix(scoreMatrix, la, la1, lb-1)
}

func _FillMatrixPeRightAlign(seqA, qualA, seqB, qualB []byte, gap int,
	scoreMatrix, pathMatrix *[]int) int {

	la := len(seqA)
	lb := len(seqB)

	// The actual gap score is the gap score times the mismatch between
	// two bases with a score of 40
	gap = gap * _NucScorePartMatchMismatch[40][40]

	needed := (la + 1) * (lb + 1)

	if needed > cap(*scoreMatrix) {
		*scoreMatrix = make([]int, needed)
	}

	if needed > cap(*pathMatrix) {
		*pathMatrix = make([]int, needed)
	}

	*scoreMatrix = (*scoreMatrix)[:needed]
	*pathMatrix = (*pathMatrix)[:needed]

	_SetMatrices(scoreMatrix, pathMatrix, la, -1, -1, 0, 0)

	// Fills the first column with score 0
	for i := 0; i < la; i++ {
		_SetMatrices(scoreMatrix, pathMatrix, la, i, -1, (i+1)*gap, -1)
	}

	lb1 := lb - 1

	for j := 0; j < lb1; j++ {

		_SetMatrices(scoreMatrix, pathMatrix, la, -1, j, 0, 1)

		for i := 0; i < la; i++ {
			left, diag, top := _GetMatrixFrom(scoreMatrix, la, i, j)

			diag += _PairingScorePeAlign(seqA[i], qualA[i], seqB[j], qualB[j])
			left += gap
			top += gap

			switch {
			case diag > left && left > top:
				_SetMatrices(scoreMatrix, pathMatrix, la, i, j, diag, 0)
			case left > diag && left > top:
				_SetMatrices(scoreMatrix, pathMatrix, la, i, j, left, +1)
			default:
				_SetMatrices(scoreMatrix, pathMatrix, la, i, j, top, -1)
			}

		}
	}

	// Special case for the last colump Up gap are free
	_SetMatrices(scoreMatrix, pathMatrix, la, -1, lb1, 0, 1)

	for i := 0; i < la; i++ {

		left, diag, top := _GetMatrixFrom(scoreMatrix, la, i, lb1)
		diag += _PairingScorePeAlign(seqA[i], qualA[i], seqB[lb1], qualB[lb1])
		left += gap

		switch {
		case diag > left && diag > top:
			_SetMatrices(scoreMatrix, pathMatrix, la, i, lb1, diag, 0)
		case left > diag && left > top:
			_SetMatrices(scoreMatrix, pathMatrix, la, i, lb1, left, +1)
		default:
			_SetMatrices(scoreMatrix, pathMatrix, la, i, lb1, top, -1)
		}
	}

	return _GetMatrix(scoreMatrix, la, la-1, lb1)
}

func PELeftAlign(seqA, seqB obiseq.BioSequence, gap int, arena PEAlignArena) (int, []int) {

	if !_InitializedDnaScore {
		log.Println("Initializing the DNA Scoring matrix")
		_InitDNAScoreMatrix()
	}

	if arena.pointer == nil {
		arena = MakePEAlignArena(seqA.Length(), seqB.Length())
	}

	score := _FillMatrixPeLeftAlign(seqA.Sequence(), seqA.Qualities(),
		seqB.Sequence(), seqB.Qualities(), gap,
		&arena.pointer.scoreMatrix,
		&arena.pointer.pathMatrix)

	arena.pointer.path = _Backtracking(arena.pointer.pathMatrix,
		seqA.Length(), seqB.Length(),
		&arena.pointer.path)

	return score, arena.pointer.path
}

func PERightAlign(seqA, seqB obiseq.BioSequence, gap int, arena PEAlignArena) (int, []int) {

	if !_InitializedDnaScore {
		log.Println("Initializing the DNA Scoring matrix")
		_InitDNAScoreMatrix()
	}

	if arena.pointer == nil {
		arena = MakePEAlignArena(seqA.Length(), seqB.Length())
	}

	score := _FillMatrixPeRightAlign(seqA.Sequence(), seqA.Qualities(),
		seqB.Sequence(), seqB.Qualities(), gap,
		&arena.pointer.scoreMatrix,
		&arena.pointer.pathMatrix)

	arena.pointer.path = _Backtracking(arena.pointer.pathMatrix,
		seqA.Length(), seqB.Length(),
		&arena.pointer.path)

	return score, arena.pointer.path
}

func PEAlign(seqA, seqB obiseq.BioSequence,
	gap, delta int,
	arena PEAlignArena) (int, []int) {
	var score, shift int
	var startA, startB int
	var partLen, over int
	var rawSeqA, qualSeqA []byte
	var rawSeqB, qualSeqB []byte
	var extra5, extra3 int

	if !_InitializedDnaScore {
		log.Println("Initializing the DNA Scoring matrix")
		_InitDNAScoreMatrix()
	}

	index := obikmer.Index4mer(seqA,
		&arena.pointer.fastIndex,
		&arena.pointer.fastBuffer)

	shift, fastScore := obikmer.FastShiftFourMer(index, seqB, nil)

	if shift > 0 {
		over = seqA.Length() - shift
	} else {
		over = seqB.Length() + shift
	}

	// log.Println(seqA.String())
	// log.Println(seqB.String())
	// log.Printf("Shift : %d Score : %d Over : %d La : %d:%d Lb: %d:%d\n", shift, fastScore, over, seqA.Length(), len(seqA.Qualities()), seqB.Length(), len(seqB.Qualities()))

	if fastScore+3 < over {

		// At least one mismatch exists in the overlaping region

		if shift > 0 {
			startA = shift - delta
			if startA < 0 {
				startA = 0
			}
			extra5 = -startA
			startB = 0
			rawSeqA = seqA.Sequence()[startA:]
			qualSeqA = seqA.Qualities()[startA:]
			partLen = len(rawSeqA)
			rawSeqB = seqB.Sequence()[0:partLen]
			qualSeqB = seqB.Qualities()[0:partLen]
			extra3 = seqB.Length() - partLen
			score = _FillMatrixPeLeftAlign(
				rawSeqA, qualSeqA, rawSeqB, qualSeqB, gap,
				&arena.pointer.scoreMatrix,
				&arena.pointer.pathMatrix)
		} else {

			// Both overlaping regions are identicals

			startA = 0
			startB = -shift - delta
			if startB < 0 {
				startB = 0
			}
			extra5 = startB
			rawSeqB = seqB.Sequence()[startB:]
			qualSeqB = seqB.Qualities()[startB:]
			partLen = len(rawSeqB)
			rawSeqA = seqA.Sequence()[:partLen]
			qualSeqA = seqA.Qualities()[:partLen]
			extra3 = partLen - seqA.Length()
			score = _FillMatrixPeRightAlign(
				rawSeqA, qualSeqA, rawSeqB, qualSeqB, gap,
				&arena.pointer.scoreMatrix,
				&arena.pointer.pathMatrix)
		}

		arena.pointer.path = _Backtracking(arena.pointer.pathMatrix,
			len(rawSeqA), len(rawSeqB),
			&arena.pointer.path)

	} else {
		if shift > 0 {
			startA = shift
			startB = 0
			extra5 = -startA
			qualSeqA = seqA.Qualities()[startA:]
			partLen = len(qualSeqA)
			qualSeqB = seqB.Qualities()[0:partLen]
			extra3 = seqB.Length() - partLen
			score = 0
		} else {
			startA = 0
			startB = -shift
			extra5 = startB
			qualSeqB = seqB.Qualities()[startB:]
			partLen = len(qualSeqB)
			extra3 = partLen - seqA.Length()
			qualSeqA = seqA.Qualities()[:partLen]
		}
		score = 0
		for i, qualA := range qualSeqA {
			qualB := qualSeqB[i]
			score += _NucScorePartMatchMatch[qualA][qualB]
		}
		arena.pointer.path = arena.pointer.path[:0]
		arena.pointer.path = append(arena.pointer.path, 0, partLen)
	}

	arena.pointer.path[0] += extra5
	if arena.pointer.path[len(arena.pointer.path)-1] == 0 {
		arena.pointer.path[len(arena.pointer.path)-2] += extra3
	} else {
		arena.pointer.path = append(arena.pointer.path, extra3, 0)
	}

	return score, arena.pointer.path
}
