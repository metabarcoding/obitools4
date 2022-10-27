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
	//lenght []int16 // Alignment length matrix
	ll    int // Length of the longest sequence
	l     int // Length of the shortest sequence
	delta int // ll - l
	extra int
	w     int
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

	needed := (L * (1 + delta + 2*extra)) * 2

	if needed > matrix.Cap() {
		matrix.matrix = make([]int16, needed)
		//matrix.lenght = make([]int16, needed)
	}

	matrix.matrix = matrix.matrix[0:needed]
	// matrix.lenght = matrix.lenght[:needed]

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

func (matrix *LCSMatrix) Get(i, j int) (int, int) {
	ij := max(0, i-matrix.extra)
	sj := min(i+matrix.delta+matrix.extra, matrix.ll)

	switch {
	case i == 0:
		return 0,j
	case j == 0:
		return 0,i
	case j < ij || j > sj:
		return -1, 30000
	default:
		offset := (matrix.extra + matrix.w*(i-1) + (matrix.w-1)*(j-i)) * 2
		return int(matrix.matrix[offset]), int(matrix.matrix[offset+1])
	}
}

func (matrix *LCSMatrix) GetNeibourgh(i, j int) (
	sd int, ld int,
	sup int, lup int,
	sleft int, lleft int) {
	offset := matrix.extra + matrix.w*(j-2) + i - j

	it := i - 1
	jt := j - 1
	
	ij0 := max(0, i-matrix.extra)
	sj0 := min(i+matrix.delta+matrix.extra, matrix.ll)
	ij1 := max(0, it-matrix.extra)
	sj1 := min(it+matrix.delta+matrix.extra, matrix.ll)

	switch {
	case i == 1:
		sd = 0
		ld = jt
	case j == 1:
		sd = 0
		ld = it
	case jt < ij1 || jt > sj1:
		sd = -1
		ld = 30000
	default:
		sd = int(matrix.matrix[offset*2])
		ld = int(matrix.matrix[offset*2+1])
	}

	offset++


	switch {
	case i == 0:
		sleft = 0
		lleft = jt
	case j == 1:
		sleft = 0
		lleft = i
	case jt < ij0 || jt > sj0:
		sleft = -1
		lleft = 30000
	default:
		sleft = int(matrix.matrix[offset*2])
		lleft = int(matrix.matrix[offset*2+1])
	}

	offset += matrix.w - 2

	switch {
	case i == 1:
		sup = 0
		lup = j
	case j == 0:
		sup = 0
		lup = it
	case j < ij1 || j > sj1:
		sup = -1
		lup = 30000
	default:
		sup = int(matrix.matrix[offset*2])
		lup = int(matrix.matrix[offset*2+1])
	}

	return
}

func (matrix *LCSMatrix) Set(i, j int, score, length int) {
	ij := max(0, i-matrix.extra)
	sj := min(i+matrix.delta+matrix.extra, matrix.ll)

	if i > 0 && j > 0 && j >= ij && j <= sj {
		offset := (matrix.extra + matrix.w*(i-1) + (matrix.w-1)*(j-i)) * 2
		matrix.matrix[offset] = int16(score)
		matrix.matrix[offset+1] = int16(length)
	}
}

// Computes the LCS between two DNA sequences and the length of the
// corresponding alignment
func LCSScore(seqA, seqB *obiseq.BioSequence, maxError int,
	matrix *LCSMatrix) (int, int) {

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

	if (seqA.Length() - seqB.Length()) > maxError {
		return -1, -1
	}

	la := seqA.Length()
	lb := seqB.Length()
	sa := seqA.Sequence()
	sb := seqB.Sequence()

	matrix = NewLCSMatrix(matrix, la, lb, maxError)

	for i := 1; i <= matrix.l; i++ {
		ij := max(1, i-matrix.extra)
		sj := min(i+matrix.delta+matrix.extra, matrix.ll)
		for j := ij; j <= sj; j++ {
			sd, ld, sup, lup, sleft, lleft := matrix.GetNeibourgh(i, j)
			if i > lb {
				log.Println("Error on seq B ", 1, matrix.l)
				log.Println(i)
				log.Println(seqB.Length(), "/", lb)
				log.Println(string(sa))
				log.Fatalln(string(sb))
			}
			if j > la {
				log.Println("Error on seq A ", ij, sj)
				log.Println(j)
				log.Println(seqA.Length(), "/", la)
				log.Println(string(sa))
				log.Fatalln(string(sb))
			}
			if sb[i-1] == sa[j-1] {
				sd++
			}
			// sup, lup := matrix.Get(i-1, j)
			// sleft, lleft := matrix.Get(i, j-1)
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

	if (l - s) > maxError {
		// log.Println(l,s,l-s,maxError)
		return -1, -1
	}

	return int(s), int(l)
}
