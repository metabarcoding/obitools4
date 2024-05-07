package obialign

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"golang.org/x/exp/constraints"
)

// abs computes the absolute value of a given float or integer.
//
// x: the input value of type k (float or integer).
// k: the return type, which is the absolute value of x.
func abs[k constraints.Float | constraints.Integer](x k) k {
	if x < 0 {
		return -x
	}

	return x
}

// D1Or0 checks if two sequences are identical or differ by one position.
//
// Parameters:
// - seq1: a pointer to the first sequence
// - seq2: a pointer to the second sequence
//
// Returns:
// - int: 0 if the sequences are identical or 0 if they differ by one position, -1 otherwise
// - int: the position where the sequences differ, -1 if they are identical
// - byte: the character in the first sequence at the differing position, '-' if it's a deletion
// - byte: the character in the second sequence at the differing position, '-' if it's a deletion
func D1Or0(seq1, seq2 *obiseq.BioSequence) (int, int, byte, byte) {

	pos := -1

	l1 := seq1.Len()
	l2 := seq2.Len()

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
