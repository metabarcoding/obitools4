package obialign

import (
	"log"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type LCSMatrix struct {
	matrix []int16 // Score matrix
	lenght []int16 // Alignment length matrix
	ll     int     // Length of the longest sequence
	l      int     // Length of the shortest sequence
	delta  int     // ll - l
	extra  int
	w      int
}

func NewLCSMatrix(matrix *LCSMatrix, L, l int, maxError int) *LCSMatrix {
	if matrix == nil {
		matrix = &LCSMatrix{}
	}

	if l > L {
		log.Panicf("L (%d) must be greater or equal to l (%d)", L, l)
	}

	delta := L - l
	extra := ((maxError - delta) / 2) + 1

	needed := L * (1 + delta + 2*extra)

	if needed > matrix.Cap() {
		matrix.matrix = make([]int16, needed)
		matrix.lenght = make([]int16, needed)
	}

	matrix.matrix = matrix.matrix[:needed]
	matrix.lenght = matrix.lenght[:needed]

	matrix.ll = L
	matrix.l = l
	matrix.delta = delta
	matrix.extra = extra
	matrix.w = delta + 1 + 2*extra

	return matrix
}

func (matrix *LCSMatrix) Cap() int {
	return cap(matrix.matrix)
}

func (matrix *LCSMatrix) Length() int {
	return len(matrix.matrix)
}

func (matrix *LCSMatrix) Get(i, j int) (int16, int16) {
	ij := max(0, i-matrix.extra)
	sj := min(i+matrix.delta+matrix.extra, matrix.ll)

	switch {
	case i == 0:
		return int16(0), int16(j)
	case j == 0:
		return int16(0), int16(i)
	case j < ij || j > sj:
		return -1, 30000
	default:
		return matrix.matrix[matrix.extra+matrix.w*(i-1)+(matrix.w-1)*(j-i)],
			matrix.lenght[matrix.extra+matrix.w*(i-1)+(matrix.w-1)*(j-i)]
	}
}

func (matrix *LCSMatrix) Set(i, j int, score, length int16) {
	ij := max(0, i-matrix.extra)
	sj := min(i+matrix.delta+matrix.extra, matrix.ll)

	if i > 0 && j > 0 && j >= ij && j <= sj {
		matrix.matrix[matrix.extra+matrix.w*(i-1)+(matrix.w-1)*(j-i)] = score
		matrix.lenght[matrix.extra+matrix.w*(i-1)+(matrix.w-1)*(j-i)] = length
	}
}

func LCSScore(seqA, seqB *obiseq.BioSequence, maxError int,
	matrix *LCSMatrix) (int, int) {
	// swapped := false

	if seqA.Length() < seqB.Length() {
		seqA, seqB = seqB, seqA
		// swapped = true
	}

	if (seqA.Length() - seqB.Length()) > maxError {
		return -1, -1
	}

	matrix = NewLCSMatrix(matrix, seqA.Length(), seqB.Length(), maxError)

	for i := 1; i <= matrix.l; i++ {
		ij := max(1, i-matrix.extra)
		sj := min(i+matrix.delta+matrix.extra, matrix.ll)
		for j := ij; j <= sj; j++ {
			sd, ld := matrix.Get(i-1, j-1)
			if seqB.Sequence()[i-1] == seqA.Sequence()[j-1] {
				sd++
			}
			sup, lup := matrix.Get(i-1, j)
			sleft, lleft := matrix.Get(i, j-1)
			switch {
			case sd >= sup && sd >= sleft:
				matrix.Set(i, j, sd, ld+1)
			case sup >= sleft && sup >= sd:
				matrix.Set(i, j, sup, lup+1)
			default:
				matrix.Set(i, j, sleft, lleft+1)
			}
		}
	}

	s, l := matrix.Get(seqB.Length(), seqA.Length())

	if (l - s) > int16(maxError) {
		return -1, -1
	}

	return int(s), int(l)
}
