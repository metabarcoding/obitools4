package obialign

import (
	"log"

	"git.metabarcoding.org/lecasofts/go/oa2/pkg/obikmer"
	"git.metabarcoding.org/lecasofts/go/oa2/pkg/obiseq"
)

type __pe_align_arena__ struct {
	score_matrix []int
	path_matrix  []int
	path         []int
	fast_index   [][]int
	fast_buffer  []byte
}

type PEAlignArena struct {
	pointer *__pe_align_arena__
}

var NilPEAlignArena = PEAlignArena{nil}

func MakePEAlignArena(lseqA, lseqB int) PEAlignArena {
	a := __pe_align_arena__{
		score_matrix: make([]int, 0, (lseqA+1)*(lseqB+1)),
		path_matrix:  make([]int, 0, (lseqA+1)*(lseqB+1)),
		path:         make([]int, 2*(lseqA+lseqB)),
		fast_index:   make([][]int, 256),
		fast_buffer:  make([]byte, 0, lseqA),
	}

	return PEAlignArena{&a}
}

func __set_matrices__(matrixA, matrixB *[]int, lenA, a, b, valueA, valueB int) {
	i := (b+1)*(lenA+1) + a + 1
	(*matrixA)[i] = valueA
	(*matrixB)[i] = valueB
}

func __get_matrix__(matrix *[]int, lenA, a, b int) int {
	return (*matrix)[(b+1)*(lenA+1)+a+1]
}

func __get_matrix_from__(matrix *[]int, lenA, a, b int) (int, int, int) {
	i := (b+1)*(lenA+1) + a
	j := i - lenA
	m := *matrix
	return m[j], m[j-1], m[i]
}

func __pairing_score_pe_align__(baseA, qualA, baseB, qualB byte) int {
	part_match := __nuc_part_match__[baseA&31][baseB&31]
	// log.Printf("id : %f A : %s %d B : %s %d\n", part_match, string(baseA), qualA, string(baseB), qualB)
	switch {
	case part_match == 1:
		// log.Printf("match\n")
		return __nuc_score_part_match_match__[qualA][qualB]
	case part_match == 0:
		return __nuc_score_part_match_mismatch__[qualA][qualB]
	default:
		return int(part_match*float64(__nuc_score_part_match_match__[qualA][qualB]) +
			(1-part_match)*float64(__nuc_score_part_match_mismatch__[qualA][qualB]) + 0.5)
	}
}

func __fill_matrix_pe_left_align__(seqA, qualA, seqB, qualB []byte, gap int,
	score_matrix, path_matrix *[]int) int {

	la := len(seqA)
	lb := len(seqB)

	// The actual gap score is the gap score times the mismatch between
	// two bases with a score of 40
	gap = gap * __nuc_score_part_match_mismatch__[40][40]

	needed := (la + 1) * (lb + 1)

	if needed > cap(*score_matrix) {
		*score_matrix = make([]int, needed)
	}

	if needed > cap(*path_matrix) {
		*path_matrix = make([]int, needed)
	}

	*score_matrix = (*score_matrix)[:needed]
	*path_matrix = (*path_matrix)[:needed]

	__set_matrices__(score_matrix, path_matrix, la, -1, -1, 0, 0)

	// Fills the first column with score 0
	for i := 0; i < la; i++ {
		__set_matrices__(score_matrix, path_matrix, la, i, -1, 0, -1)
	}

	la1 := la - 1

	for j := 0; j < lb; j++ {

		__set_matrices__(score_matrix, path_matrix, la, -1, j, (j+1)*gap, 1)

		for i := 0; i < la1; i++ {
			left, diag, top := __get_matrix_from__(score_matrix, la, i, j)
			diag += __pairing_score_pe_align__(seqA[i], qualA[i], seqB[j], qualB[j])
			left += gap
			top += gap
			switch {
			case diag > left && diag > top:
				__set_matrices__(score_matrix, path_matrix, la, i, j, diag, 0)
			case left > diag && left > top:
				__set_matrices__(score_matrix, path_matrix, la, i, j, left, +1)
			default:
				__set_matrices__(score_matrix, path_matrix, la, i, j, top, -1)
			}
		}

		// Special case for the last line Left gap are free

		left, diag, top := __get_matrix_from__(score_matrix, la, la1, j)
		diag += __pairing_score_pe_align__(seqA[la1], qualA[la1], seqB[j], qualB[j])
		top += gap

		switch {
		case diag > left && diag > top:
			__set_matrices__(score_matrix, path_matrix, la, la1, j, diag, 0)
		case left > diag && left > top:
			__set_matrices__(score_matrix, path_matrix, la, la1, j, left, +1)
		default:
			__set_matrices__(score_matrix, path_matrix, la, la1, j, top, -1)

		}
	}

	return __get_matrix__(score_matrix, la, la1, lb-1)
}

func __fill_matrix_pe_right_align__(seqA, qualA, seqB, qualB []byte, gap int,
	score_matrix, path_matrix *[]int) int {

	la := len(seqA)
	lb := len(seqB)

	// The actual gap score is the gap score times the mismatch between
	// two bases with a score of 40
	gap = gap * __nuc_score_part_match_mismatch__[40][40]

	needed := (la + 1) * (lb + 1)

	if needed > cap(*score_matrix) {
		*score_matrix = make([]int, needed)
	}

	if needed > cap(*path_matrix) {
		*path_matrix = make([]int, needed)
	}

	*score_matrix = (*score_matrix)[:needed]
	*path_matrix = (*path_matrix)[:needed]

	__set_matrices__(score_matrix, path_matrix, la, -1, -1, 0, 0)

	// Fills the first column with score 0
	for i := 0; i < la; i++ {
		__set_matrices__(score_matrix, path_matrix, la, i, -1, (i+1)*gap, -1)
	}

	lb1 := lb - 1

	for j := 0; j < lb1; j++ {

		__set_matrices__(score_matrix, path_matrix, la, -1, j, 0, 1)

		for i := 0; i < la; i++ {
			left, diag, top := __get_matrix_from__(score_matrix, la, i, j)

			diag += __pairing_score_pe_align__(seqA[i], qualA[i], seqB[j], qualB[j])
			left += gap
			top += gap

			switch {
			case diag > left && left > top:
				__set_matrices__(score_matrix, path_matrix, la, i, j, diag, 0)
			case left > diag && left > top:
				__set_matrices__(score_matrix, path_matrix, la, i, j, left, +1)
			default:
				__set_matrices__(score_matrix, path_matrix, la, i, j, top, -1)
			}

		}
	}

	// Special case for the last colump Up gap are free
	__set_matrices__(score_matrix, path_matrix, la, -1, lb1, 0, 1)

	for i := 0; i < la; i++ {

		left, diag, top := __get_matrix_from__(score_matrix, la, i, lb1)
		diag += __pairing_score_pe_align__(seqA[i], qualA[i], seqB[lb1], qualB[lb1])
		left += gap

		switch {
		case diag > left && diag > top:
			__set_matrices__(score_matrix, path_matrix, la, i, lb1, diag, 0)
		case left > diag && left > top:
			__set_matrices__(score_matrix, path_matrix, la, i, lb1, left, +1)
		default:
			__set_matrices__(score_matrix, path_matrix, la, i, lb1, top, -1)
		}
	}

	return __get_matrix__(score_matrix, la, la-1, lb1)
}

func PELeftAlign(seqA, seqB obiseq.BioSequence, gap int, arena PEAlignArena) (int, []int) {

	if !__initialized_dna_score__ {
		log.Println("Initializing the DNA Scoring matrix")
		InitDNAScoreMatrix()
	}

	if arena.pointer == nil {
		arena = MakePEAlignArena(seqA.Length(), seqB.Length())
	}

	score := __fill_matrix_pe_left_align__(seqA.Sequence(), seqA.Qualities(),
		seqB.Sequence(), seqB.Qualities(), gap,
		&arena.pointer.score_matrix,
		&arena.pointer.path_matrix)

	arena.pointer.path = __backtracking__(arena.pointer.path_matrix,
		seqA.Length(), seqB.Length(),
		&arena.pointer.path)

	return score, arena.pointer.path
}

func PERightAlign(seqA, seqB obiseq.BioSequence, gap int, arena PEAlignArena) (int, []int) {

	if !__initialized_dna_score__ {
		log.Println("Initializing the DNA Scoring matrix")
		InitDNAScoreMatrix()
	}

	if arena.pointer == nil {
		arena = MakePEAlignArena(seqA.Length(), seqB.Length())
	}

	score := __fill_matrix_pe_right_align__(seqA.Sequence(), seqA.Qualities(),
		seqB.Sequence(), seqB.Qualities(), gap,
		&arena.pointer.score_matrix,
		&arena.pointer.path_matrix)

	arena.pointer.path = __backtracking__(arena.pointer.path_matrix,
		seqA.Length(), seqB.Length(),
		&arena.pointer.path)

	return score, arena.pointer.path
}

func PEAlign(seqA, seqB obiseq.BioSequence,
	gap, delta int,
	arena PEAlignArena) (int, []int) {
	var score, shift int
	var startA, startB int
	var part_len, over int
	var raw_seqA, qual_seqA []byte
	var raw_seqB, qual_seqB []byte
	var extra5, extra3 int

	if !__initialized_dna_score__ {
		log.Println("Initializing the DNA Scoring matrix")
		InitDNAScoreMatrix()
	}

	index := obikmer.Index4mer(seqA,
		&arena.pointer.fast_index,
		&arena.pointer.fast_buffer)

	shift, fast_score := obikmer.FastShiftFourMer(index, seqB, nil)

	if shift > 0 {
		over = seqA.Length() - shift
	} else {
		over = seqB.Length() + shift
	}

	if fast_score+3 < over {
		if shift > 0 {
			startA = shift - delta
			if startA < 0 {
				startA = 0
			}
			extra5 = -startA
			startB = 0
			raw_seqA = seqA.Sequence()[startA:]
			qual_seqA = seqA.Qualities()[startA:]
			part_len = len(raw_seqA)
			raw_seqB = seqB.Sequence()[0:part_len]
			qual_seqB = seqB.Qualities()[0:part_len]
			extra3 = seqB.Length() - part_len
			score = __fill_matrix_pe_left_align__(
				raw_seqA, qual_seqA, raw_seqB, qual_seqB, gap,
				&arena.pointer.score_matrix,
				&arena.pointer.path_matrix)
		} else {
			startA = 0
			startB = -shift - delta
			if startB < 0 {
				startB = 0
			}
			extra5 = startB
			raw_seqB = seqB.Sequence()[startB:]
			qual_seqB = seqB.Qualities()[startB:]
			part_len = len(raw_seqB)
			raw_seqA = seqA.Sequence()[:part_len]
			qual_seqA = seqA.Qualities()[:part_len]
			extra3 = part_len - seqA.Length()
			score = __fill_matrix_pe_right_align__(
				raw_seqA, qual_seqA, raw_seqB, qual_seqB, gap,
				&arena.pointer.score_matrix,
				&arena.pointer.path_matrix)
		}

		arena.pointer.path = __backtracking__(arena.pointer.path_matrix,
			len(raw_seqA), len(raw_seqB),
			&arena.pointer.path)

	} else {
		if shift > 0 {
			startA = shift
			startB = 0
			extra5 = -startA
			qual_seqA = seqA.Qualities()[startA:]
			part_len = len(qual_seqA)
			qual_seqB = seqB.Qualities()[0:part_len]
			extra3 = seqB.Length() - part_len
			score = 0
		} else {
			startA = 0
			startB = -shift
			extra5 = startB
			qual_seqB = seqB.Qualities()[startB:]
			part_len = len(qual_seqB)
			extra3 = part_len - seqA.Length()
			qual_seqA = seqA.Qualities()[:part_len]
		}
		score = 0
		for i, qualA := range qual_seqA {
			qualB := qual_seqB[i]
			score += __nuc_score_part_match_match__[qualA][qualB]
		}
		arena.pointer.path = arena.pointer.path[:0]
		arena.pointer.path = append(arena.pointer.path, 0, part_len)
	}

	arena.pointer.path[0] += extra5
	if arena.pointer.path[len(arena.pointer.path)-1] == 0 {
		arena.pointer.path[len(arena.pointer.path)-2] += extra3
	} else {
		arena.pointer.path = append(arena.pointer.path, extra3, 0)
	}

	return score, arena.pointer.path
}
