package obialign

import (
	"log"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

type _PeAlignArena struct {
	scoreMatrix []int
	pathMatrix  []int
	path        []int
	fastIndex   [][]int
	fastBuffer  []byte
	aligneSeqA  *[]byte
	aligneSeqB  *[]byte
	aligneQualA *[]byte
	aligneQualB *[]byte
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
	aligneSeqA := make([]byte, 0, lseqA+lseqB)
	aligneSeqB := make([]byte, 0, lseqA+lseqB)
	aligneQualA := make([]byte, 0, lseqA+lseqB)
	aligneQualB := make([]byte, 0, lseqA+lseqB)

	a := _PeAlignArena{
		scoreMatrix: make([]int, 0, (lseqA+1)*(lseqB+1)),
		pathMatrix:  make([]int, 0, (lseqA+1)*(lseqB+1)),
		path:        make([]int, 2*(lseqA+lseqB)),
		fastIndex:   make([][]int, 256),
		fastBuffer:  make([]byte, 0, lseqA),
		aligneSeqA:  &aligneSeqA,
		aligneSeqB:  &aligneSeqB,
		aligneQualA: &aligneQualA,
		aligneQualB: &aligneQualB,
	}

	return PEAlignArena{&a}
}

// _SetMatrices updates the values in matrixA and matrixB at the specified indices with the given values.
//
// The data in the matrix is stored in column-major order.
// Positions in the matrix are numbered from -1
//
// Parameters:
// - matrixA: a pointer to the slice of integers representing matrixA.
// - matrixB: a pointer to the slice of integers representing matrixB.
// - lenA: the length of matrixA.
// - a: the row index a.
// - b: the column index b .
// - valueA: the value to be set in matrixA.
// - valueB: the value to be set in matrixB.
func _SetMatrices(matrixA, matrixB *[]int, lenA, a, b, valueA, valueB int) {
	i := (b+1)*(lenA+1) + a + 1
	(*matrixA)[i] = valueA
	(*matrixB)[i] = valueB
}

// _GetMatrix returns the value at the specified position in the matrix.
//
// The data in the matrix is stored in column-major order.
// Positions in the matrix are numbered from -1
//
// Parameters:
// - matrix: a pointer to the matrix
// - lenA: the length of the matrix
// - a: the row index a.
// - b: the column index b .
//
// Returns:
// - int: the value at the specified position in the matrix
func _GetMatrix(matrix *[]int, lenA, a, b int) int {
	return (*matrix)[(b+1)*(lenA+1)+a+1]
}

// Returns left, diag, top compare to the position (a, b)
//
// with a the row index and b the column index.
// Positions in the matrix are numbered from -1
//
// i = (b+1)*(lenA+1) + a + 1
//
// diag = M[a-1][b-1]  : i_diag = ((b-1)+1)*(lenA+1) + (a-1) + 1
//
//	: i_diag = b*(lenA+1) + a
//
// left = M[a][b-1]    : i_left = ((b-1)+1)*(lenA+1) + a + 1
//
//	: i_left = b*(lenA+1) + a + 1
//
// top = M[a-1][b]	   : i_top = (b+1)*(lenA+1) + (a-1) + 1
//
//	: i_top = (b+1)*(lenA+1) + a
//
// Parameters:
//
// - matrix: a pointer to the matrix
// - lenA: the number of row -1 in the matrix
// - a: the row index a.
// - b: the column index b .
func _GetMatrixFrom(matrix *[]int, lenA, a, b int) (int, int, int) {
	// Formula rechecked on 11/24/2023
	i_top := (b+1)*(lenA+1) + a
	i_left := i_top - lenA
	i_diag := i_left - 1
	m := *matrix
	return m[i_left], m[i_diag], m[i_top]
}

func _PairingScorePeAlign(baseA, qualA, baseB, qualB byte, scale float64) int {
	partMatch := _NucPartMatch[baseA&31][baseB&31]
	// log.Printf("id : %f A : %s %d B : %s %d\n", part_match, string(baseA), qualA, string(baseB), qualB)
	switch int(partMatch * 100) {
	case 100:
		return _NucScorePartMatchMatch[qualA][qualB]
	case 0:
		return int(float64(_NucScorePartMatchMismatch[qualA][qualB])*scale + 0.5)
	default:
		return int(partMatch*float64(_NucScorePartMatchMatch[qualA][qualB]) +
			(1-partMatch)*float64(_NucScorePartMatchMismatch[qualA][qualB])*scale +
			0.5)
	}
}

// Gaps at the beginning of seqB and at the end of seqA are free
// With seqA spanning over lines and seqB over columns
//   - First column gap = 0
//   - Last line gaps = 0
//
// Paths are encoded :
//   - 0  :  for diagonal
//   - -1 :  for top
//   - +1 :  for left
func _FillMatrixPeLeftAlign(seqA, qualA, seqB, qualB []byte, gap, scale float64,
	scoreMatrix, pathMatrix *[]int) int {

	la := len(seqA)
	lb := len(seqB)

	// The actual gap score is the gap score times the mismatch between
	// two bases with a score of 40
	gapPenalty := int(scale*gap*float64(_NucScorePartMatchMismatch[40][40]) + 0.5)

	needed := (la + 1) * (lb + 1)

	if needed > cap(*scoreMatrix) {
		*scoreMatrix = make([]int, needed)
	}

	if needed > cap(*pathMatrix) {
		*pathMatrix = make([]int, needed)
	}

	*scoreMatrix = (*scoreMatrix)[:needed]
	*pathMatrix = (*pathMatrix)[:needed]

	// Sets the first position of the matrix with 0 score
	_SetMatrices(scoreMatrix, pathMatrix, la, -1, -1, 0, 0)

	// Fills the first column with score 0
	for i := 0; i < la; i++ {
		_SetMatrices(scoreMatrix, pathMatrix, la, i, -1, 0, -1)
	}

	la1 := la - 1 // Except the last line (gaps are free on it)

	for j := 0; j < lb; j++ {

		// Fill the first line with scores corresponding to a set of gaps
		_SetMatrices(scoreMatrix, pathMatrix, la, -1, j, (j+1)*gapPenalty, 1)

		for i := 0; i < la1; i++ {
			left, diag, top := _GetMatrixFrom(scoreMatrix, la, i, j)
			// log.Infof("LA: i : %d j : %d left : %d diag : %d top : %d\n", i, j, left, diag, top)

			diag += _PairingScorePeAlign(seqA[i], qualA[i], seqB[j], qualB[j], scale)
			left += gapPenalty
			top += gapPenalty

			switch {
			case diag >= left && diag >= top:
				_SetMatrices(scoreMatrix, pathMatrix, la, i, j, diag, 0)
			case left >= diag && left >= top:
				_SetMatrices(scoreMatrix, pathMatrix, la, i, j, left, +1)
			default:
				_SetMatrices(scoreMatrix, pathMatrix, la, i, j, top, -1)
			}
			// log.Infof("LA: i : %d j : %d left : %d diag : %d top : %d [%d]\n", i, j, left, diag, top, _GetMatrix(scoreMatrix, la, i, j))
		}

		// Special case for the last line Left gap are free

		left, diag, top := _GetMatrixFrom(scoreMatrix, la, la1, j)
		diag += _PairingScorePeAlign(seqA[la1], qualA[la1], seqB[j], qualB[j], scale)
		top += gapPenalty

		switch {
		case diag >= left && diag >= top:
			_SetMatrices(scoreMatrix, pathMatrix, la, la1, j, diag, 0)
		case left >= diag && left >= top:
			_SetMatrices(scoreMatrix, pathMatrix, la, la1, j, left, +1)
		default:
			_SetMatrices(scoreMatrix, pathMatrix, la, la1, j, top, -1)

		}

		// log.Infof("LA: i : %d j : %d left : %d diag : %d top : %d [%d]\n", la1, j, left, diag, top, _GetMatrix(scoreMatrix, la, la1, j))
	}

	return _GetMatrix(scoreMatrix, la, la1, lb-1)
}

// Gaps at the beginning of A and at the end of B are free
// With A spanning over lines and B over columns
//   - First line gap = 0
//   - Last column gaps = 0
func _FillMatrixPeRightAlign(seqA, qualA, seqB, qualB []byte, gap, scale float64,
	scoreMatrix, pathMatrix *[]int) int {

	la := len(seqA)
	lb := len(seqB)

	// The actual gap score is the gap score times the mismatch between
	// two bases with a score of 40
	gapPenalty := int(scale*gap*float64(_NucScorePartMatchMismatch[40][40]) + 0.5)

	needed := (la + 1) * (lb + 1)

	if needed > cap(*scoreMatrix) {
		*scoreMatrix = make([]int, needed)
	}

	if needed > cap(*pathMatrix) {
		*pathMatrix = make([]int, needed)
	}

	*scoreMatrix = (*scoreMatrix)[:needed]
	*pathMatrix = (*pathMatrix)[:needed]

	// Sets the first position of the matrix with 0 score
	_SetMatrices(scoreMatrix, pathMatrix, la, -1, -1, 0, 0)

	// Fills the first column with scores corresponding to a set of gaps
	for i := 0; i < la; i++ {
		_SetMatrices(scoreMatrix, pathMatrix, la, i, -1, (i+1)*gapPenalty, -1)
	}

	lb1 := lb - 1 // Except the last column (gaps are free on it)

	for j := 0; j < lb1; j++ {

		// Fill the first line with zero score
		_SetMatrices(scoreMatrix, pathMatrix, la, -1, j, 0, 1)

		for i := 0; i < la; i++ {
			left, diag, top := _GetMatrixFrom(scoreMatrix, la, i, j)

			diag += _PairingScorePeAlign(seqA[i], qualA[i], seqB[j], qualB[j], scale)
			left += gapPenalty
			top += gapPenalty

			switch {
			case diag >= left && diag >= top:
				_SetMatrices(scoreMatrix, pathMatrix, la, i, j, diag, 0)
			case left >= diag && left >= top:
				_SetMatrices(scoreMatrix, pathMatrix, la, i, j, left, +1)
			default:
				_SetMatrices(scoreMatrix, pathMatrix, la, i, j, top, -1)
			}

			// log.Infof("LR: i : %d j : %d left : %d diag : %d top : %d [%d]\n", i, j, left, diag, top, _GetMatrix(scoreMatrix, la, i, j))

		}
	}

	// Special case for the last colump Up gap are free
	_SetMatrices(scoreMatrix, pathMatrix, la, -1, lb1, 0, 1)

	for i := 0; i < la; i++ {

		left, diag, top := _GetMatrixFrom(scoreMatrix, la, i, lb1)
		diag += _PairingScorePeAlign(seqA[i], qualA[i], seqB[lb1], qualB[lb1], scale)
		left += gapPenalty

		// log.Infof("LR: i : %d j : %d left : %d diag : %d top : %d [%d]\n", i, lb1, left, diag, top, _GetMatrix(scoreMatrix, la, i, lb1))

		switch {
		case diag >= left && diag >= top:
			_SetMatrices(scoreMatrix, pathMatrix, la, i, lb1, diag, 0)
		case left >= diag && left >= top:
			_SetMatrices(scoreMatrix, pathMatrix, la, i, lb1, left, +1)
		default:
			_SetMatrices(scoreMatrix, pathMatrix, la, i, lb1, top, -1)
		}
	}

	return _GetMatrix(scoreMatrix, la, la-1, lb1)

}

// Gaps at the beginning and at the end of seqA are free
// With seqA spanning over lines and seqB over columns
//
// SeqA must be the longer sequence. If that constraint is not
// respected, the function will panic.
//
// TO BE FINISHED
//   - First column gap = 0
//   - Last column gaps = 0
//
// Paths are encoded :
//   - 0  :  for diagonal
//   - -1 :  for top
//   - +1 :  for left
func _FillMatrixPeCenterAlign(seqA, qualA, seqB, qualB []byte, gap, scale float64,
	scoreMatrix, pathMatrix *[]int) int {

	la := len(seqA)
	lb := len(seqB)

	if len(seqA) < len(seqB) {
		log.Panicf("len(seqA) < len(seqB) : %d < %d", len(seqA), len(seqB))
	}

	// The actual gap score is the gap score times the mismatch between
	// two bases with a score of 40
	gapPenalty := int(scale*gap*float64(_NucScorePartMatchMismatch[40][40]) + 0.5)

	needed := (la + 1) * (lb + 1)

	if needed > cap(*scoreMatrix) {
		*scoreMatrix = make([]int, needed)
	}

	if needed > cap(*pathMatrix) {
		*pathMatrix = make([]int, needed)
	}

	*scoreMatrix = (*scoreMatrix)[:needed]
	*pathMatrix = (*pathMatrix)[:needed]

	// Sets the first position of the matrix with 0 score
	_SetMatrices(scoreMatrix, pathMatrix, la, -1, -1, 0, 0)

	// Fills the first column with score 0
	for i := 0; i < la; i++ {
		_SetMatrices(scoreMatrix, pathMatrix, la, i, -1, 0, -1)
	}

	// la1 := la - 1 // Except the last line (gaps are free on it)
	lb1 := lb - 1 // Except the last column (gaps are free on it)

	for j := 0; j < lb1; j++ {

		// Fill the first line with scores corresponding to a set of gaps
		_SetMatrices(scoreMatrix, pathMatrix, la, -1, j, (j+1)*gapPenalty, 1)

		for i := 0; i < la; i++ {
			left, diag, top := _GetMatrixFrom(scoreMatrix, la, i, j)
			// log.Infof("LA: i : %d j : %d left : %d diag : %d top : %d\n", i, j, left, diag, top)

			diag += _PairingScorePeAlign(seqA[i], qualA[i], seqB[j], qualB[j], scale)
			left += gapPenalty
			top += gapPenalty

			switch {
			case diag >= left && diag >= top:
				_SetMatrices(scoreMatrix, pathMatrix, la, i, j, diag, 0)
			case left >= diag && left >= top:
				_SetMatrices(scoreMatrix, pathMatrix, la, i, j, left, +1)
			default:
				_SetMatrices(scoreMatrix, pathMatrix, la, i, j, top, -1)
			}
			// log.Infof("LA: i : %d j : %d left : %d diag : %d top : %d [%d]\n", i, j, left, diag, top, _GetMatrix(scoreMatrix, la, i, j))
		}

	}

	for i := 0; i < la; i++ {
		left, diag, top := _GetMatrixFrom(scoreMatrix, la, i, lb1)
		// log.Infof("LA: i : %d j : %d left : %d diag : %d top : %d\n", i, j, left, diag, top)

		diag += _PairingScorePeAlign(seqA[i], qualA[i], seqB[lb1], qualB[lb1], scale)
		left += gapPenalty

		switch {
		case diag >= left && diag >= top:
			_SetMatrices(scoreMatrix, pathMatrix, la, i, lb1, diag, 0)
		case left >= diag && left >= top:
			_SetMatrices(scoreMatrix, pathMatrix, la, i, lb1, left, +1)
		default:
			_SetMatrices(scoreMatrix, pathMatrix, la, i, lb1, top, -1)
		}
		// log.Infof("LA: i : %d j : %d left : %d diag : %d top : %d [%d]\n", i, j, left, diag, top, _GetMatrix(scoreMatrix, la, i, j))
	}

	return _GetMatrix(scoreMatrix, la, la-1, lb1)
}

func PELeftAlign(seqA, seqB *obiseq.BioSequence, gap, scale float64,
	arena PEAlignArena) (int, []int) {

	if !_InitializedDnaScore {
		_InitDNAScoreMatrix()
	}

	if arena.pointer == nil {
		arena = MakePEAlignArena(seqA.Len(), seqB.Len())
	}

	score := _FillMatrixPeLeftAlign(seqA.Sequence(), seqA.Qualities(),
		seqB.Sequence(), seqB.Qualities(), gap, scale,
		&arena.pointer.scoreMatrix,
		&arena.pointer.pathMatrix)

	path := _Backtracking(arena.pointer.pathMatrix,
		seqA.Len(), seqB.Len(),
		&arena.pointer.path)

	return score, path
}

func PERightAlign(seqA, seqB *obiseq.BioSequence, gap, scale float64,
	arena PEAlignArena) (int, []int) {

	if !_InitializedDnaScore {
		_InitDNAScoreMatrix()
	}

	if arena.pointer == nil {
		arena = MakePEAlignArena(seqA.Len(), seqB.Len())
	}

	score := _FillMatrixPeRightAlign(seqA.Sequence(), seqA.Qualities(),
		seqB.Sequence(), seqB.Qualities(), gap, scale,
		&arena.pointer.scoreMatrix,
		&arena.pointer.pathMatrix)

	path := _Backtracking(arena.pointer.pathMatrix,
		seqA.Len(), seqB.Len(),
		&arena.pointer.path)

	return score, path
}

func PECenterAlign(seqA, seqB *obiseq.BioSequence, gap, scale float64,
	arena PEAlignArena) (int, []int) {

	if !_InitializedDnaScore {
		_InitDNAScoreMatrix()
	}

	if arena.pointer == nil {
		arena = MakePEAlignArena(seqA.Len(), seqB.Len())
	}

	score := _FillMatrixPeCenterAlign(seqA.Sequence(), seqA.Qualities(),
		seqB.Sequence(), seqB.Qualities(), gap, scale,
		&arena.pointer.scoreMatrix,
		&arena.pointer.pathMatrix)

	path := _Backtracking(arena.pointer.pathMatrix,
		seqA.Len(), seqB.Len(),
		&arena.pointer.path)

	return score, path
}

func PEAlign(seqA, seqB *obiseq.BioSequence,
	gap, scale float64, fastAlign bool, delta int, fastScoreRel bool,
	arena PEAlignArena, shift_buff *map[int]int) (bool, int, []int, int, int, float64) {
	var isLeftAlign bool
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

	if fastAlign {

		index := obikmer.Index4mer(seqA,
			&arena.pointer.fastIndex,
			&arena.pointer.fastBuffer)

		shift, fastCount, fastScore = obikmer.FastShiftFourMer(index, shift_buff, seqA.Len(), seqB, fastScoreRel, nil)

		if shift > 0 {
			over = seqA.Len() - shift
		} else {
			over = seqB.Len() + shift
		}

		// At least one mismatch exists in the overlaping region
		if fastCount+3 < over {

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
				if partLen > seqB.Len() {
					partLen = seqB.Len()
				}
				rawSeqB = seqB.Sequence()[0:partLen]
				qualSeqB = seqB.Qualities()[0:partLen]
				extra3 = seqB.Len() - partLen
				isLeftAlign = true
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
				isLeftAlign = false
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

			if shift > 0 {
				startA = shift
				startB = 0
				extra5 = -startA
				qualSeqA = seqA.Qualities()[startA:]
				partLen = len(qualSeqA)
				qualSeqB = seqB.Qualities()[0:partLen]
				extra3 = seqB.Len() - partLen
				score = 0
				isLeftAlign = true
			} else {
				startA = 0
				startB = -shift
				extra5 = startB
				qualSeqB = seqB.Qualities()[startB:]
				partLen = len(qualSeqB)
				extra3 = partLen - seqA.Len()
				qualSeqA = seqA.Qualities()[:partLen]
				isLeftAlign = false
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
	} else {
		//
		// No Fast Heuristic
		//

		rawSeqA = seqA.Sequence()
		qualSeqA = seqA.Qualities()
		rawSeqB = seqB.Sequence()
		qualSeqB = seqB.Qualities()

		scoreR := _FillMatrixPeRightAlign(
			rawSeqA, qualSeqA, rawSeqB, qualSeqB, gap, scale,
			&arena.pointer.scoreMatrix,
			&arena.pointer.pathMatrix)

		path = _Backtracking(arena.pointer.pathMatrix,
			len(rawSeqA), len(rawSeqB),
			&(arena.pointer.path))

		isLeftAlign = false

		scoreL := _FillMatrixPeLeftAlign(
			rawSeqA, qualSeqA, rawSeqB, qualSeqB, gap, scale,
			&arena.pointer.scoreMatrix,
			&arena.pointer.pathMatrix)

		if scoreL > scoreR {
			path = _Backtracking(arena.pointer.pathMatrix,
				len(rawSeqA), len(rawSeqB),
				&(arena.pointer.path))
			isLeftAlign = true
		}

	}

	return isLeftAlign, score, path, fastCount, over, fastScore
}
