package obikmer

import (
	"math"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

type Table4mer [256]uint16

func Count4Mer(seq *obiseq.BioSequence, buffer *[]byte, counts *Table4mer) *Table4mer {
	iternal_buffer := Encode4mer(seq, buffer) // The slice of 4-mer codes

	if counts == nil {
		var w Table4mer
		counts = &w
	}

	// Every cells of the counter is set to zero
	for i := 0; i < 256; i++ { // 256 is the number of possible 4-mer codes
		(*counts)[i] = 0
	}

	for _, code := range iternal_buffer {
		(*counts)[code]++
	}
	return counts
}

func Common4Mer(count1, count2 *Table4mer) int {
	sum := 0
	for i := 0; i < 256; i++ {
		sum += int(min((*count1)[i], (*count2)[i]))
	}
	return sum
}

func Sum4Mer(count *Table4mer) int {
	sum := 0
	for i := 0; i < 256; i++ {
		sum += int((*count)[i])
	}
	return sum
}

func LCS4MerBounds(count1, count2 *Table4mer) (int, int) {
	s1 := Sum4Mer(count1)
	s2 := Sum4Mer(count2)
	smin := min(s1, s2)

	cw := Common4Mer(count1, count2)

	lcsMax := smin + 3 - int(math.Ceil(float64(smin-cw)/4.0))
	lcsMin := cw

	if cw > 0 {
		lcsMin += 3
	}

	return lcsMin, lcsMax
}

func Error4MerBounds(count1, count2 *Table4mer) (int, int) {
	s1 := Sum4Mer(count1)
	s2 := Sum4Mer(count2)
	smax := max(s1, s2)

	cw := Common4Mer(count1, count2)

	errorMax := smax - cw + 2*int(math.Floor(float64(cw+5)/8.0))
	errorMin := int(math.Ceil(float64(errorMax) / 4.0))

	return errorMin, errorMax
}
