package obialign

import (
	"log"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

type FullLCSMatrix struct {
	matrix []int16 // Score matrix
	lenght []int16 // Alignment length matrix
	ll     int     // Length of the longest sequence
	l      int     // Length of the shortest sequence
}

func NewFullLCSMatrix(matrix *FullLCSMatrix, L, l int) *FullLCSMatrix {
	if matrix == nil {
		matrix = &FullLCSMatrix{}
	}

	if l > L {
		log.Panicf("L (%d) must be greater or equal to l (%d)", L, l)
	}

	needed := (L) * (l)

	if needed > matrix.Cap() {
		matrix.matrix = make([]int16, needed)
		matrix.lenght = make([]int16, needed)
	}

	matrix.matrix = matrix.matrix[:needed]
	matrix.lenght = matrix.lenght[:needed]

	matrix.ll = L
	matrix.l = l

	return matrix
}

func (matrix *FullLCSMatrix) Cap() int {
	return cap(matrix.matrix)
}

func (matrix *FullLCSMatrix) Length() int {
	return len(matrix.matrix)
}

func (matrix *FullLCSMatrix) Get(i, j int) (int16, int16) {
	if i == 0 {
		return 0, int16(j)
	}
	if j == 0 {
		return 0, int16(i)
	}

	pos := (i-1)*matrix.ll + j - 1
	return matrix.matrix[pos], matrix.lenght[pos]
}

func (matrix *FullLCSMatrix) Set(i, j int, score, length int16) {
	if i > 0 && j > 0 {
		pos := (i-1)*matrix.ll + j - 1
		matrix.matrix[pos] = score
		matrix.lenght[pos] = length
	}
}

// Computes the LCS between two DNA sequences and the length of the
// corresponding alignment
func FullLCSScore(seqA, seqB *obiseq.BioSequence,
	matrix *FullLCSMatrix) (int, int) {

	if seqA.Length() == 0 {
		log.Fatal("Sequence A has a length of 0")
	}

	if seqB.Length() == 0 {
		log.Fatal("Sequence B has a length of 0")
	}
	// swapped := false

	if seqA.Length() < seqB.Length() {
		seqA, seqB = seqB, seqA
		// swapped = true
	}

	la := seqA.Length()
	lb := seqB.Length()
	sa := seqA.Sequence()
	sb := seqB.Sequence()

	matrix = NewFullLCSMatrix(matrix, la, lb)

	for i := 1; i <= matrix.l; i++ {
		for j := 1; j <= matrix.ll; j++ {
			sd, ld := matrix.Get(i-1, j-1)

			if sb[i-1] == sa[j-1] {
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

	s, l := matrix.Get(lb, la)

	return int(s), int(l)
}
