package obialign

import "git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"

func abs(x int) int {
	if x < 0 {
		return -x
	}

	return x
}

func D1Or0(seq1, seq2 *obiseq.BioSequence) (int, int, byte, byte) {

	pos := -1

	l1 := seq1.Length()
	l2 := seq2.Length()

	if abs(l1-l2) > 1 {
		return -1, pos, 0, 0
	}

	s1 := seq1.Sequence()
	s2 := seq2.Sequence()

	b1 := 0
	b2 := 0

	// Scans the sequences from their beginings as long as they are identical

	for b1 < l1 && b2 < l2 && s1[b1] == s2[b2] {
		b1++
		b2++
	}

	if b1 == l1 && b2 == l2 {
		return 0, pos, 0, 0
	}

	// Scans the sequences from their ends as long as they are identical

	e1 := l1 - 1
	e2 := l2 - 1

	for (e1 > b1 || e2 > b2) && s1[e1] == s2[e2] {
		e1--
		e2--
	}

	if (l1 == l2 && (e1 > b1 || e2 > b2)) ||
		(l1 > l2 && e1 > b1) ||
		(l1 < l2 && e2 > b2) {
		return -1, pos, 0, 0
	}

	if b1 >= e1 {
		if e1 > e2 {
			pos = e1
		} else {
			pos = e2
		}
	}

	a1 := byte('-')
	a2 := byte('-')

	if e2 >= e1 {
		a2 = s2[e2]
	}

	if e2 <= e1 {
		a1 = s1[e1]
	}

	return 1, pos, a1, a2

}
