package obialign

import log "github.com/sirupsen/logrus"

// buffIndex converts a pair of coordinates (i, j) into a linear index in a matrix
// of size width x width. The coordinates are (-1)-indexed, and the linear index
// is 0-indexed as well. The function first adds 1 to both coordinates to make
// sure the (-1,-1) coordinate is at position 0 in the matrix, and then computes
// the linear index by multiplying the first coordinate by the width and adding
// the second coordinate.
func buffIndex(i, j, width int) int {
	return (i+1)*width + (j + 1)
}

// LocatePattern is a function to locate a pattern in a sequence.
//
// It uses a dynamic programming approach to build a matrix of scores.
// The score at each cell is the maximum of the score of the cell
// above it (representing a deletion), the score of the cell to its
// left (representing an insertion), and the score of the cell
// diagonally above it (representing a match).
//
// The score of a match is 0 if the two characters are the same,
// and -1 if they are different.
//
// The function returns the start and end positions of the best
// match, as well as the number of errors in the best match.
func LocatePattern(id string, pattern, sequence []byte) (int, int, int) {

	if len(pattern) >= len(sequence) {
		log.Panicf("Sequence %s:Pattern %s must be shorter than sequence %s", id, pattern, sequence)
	}

	// Pattern spreads over the columns
	// Sequence spreads over the rows
	width := len(pattern) + 1
	buffsize := (len(pattern) + 1) * (len(sequence) + 1)
	buffer := make([]int, buffsize)

	// The path matrix keeps track of the best path through the matrix
	//  0 : indicate the diagonal path
	//  1 : indicate the up path
	// -1 : indicate the left path
	path := make([]int, buffsize)

	// Initialize the first row of the matrix
	for j := 0; j < len(pattern); j++ {
		idx := buffIndex(-1, j, width)
		buffer[idx] = -j - 1
		path[idx] = -1
	}

	// Initialize the first column of the matrix
	// Alignment is endgap free so first column = 0
	// to allow primer to shift freely along the sequence
	for i := -1; i < len(sequence); i++ {
		idx := buffIndex(i, -1, width)
		buffer[idx] = 0
		path[idx] = +1
	}

	// Fills the matrix except the last column
	// where gaps must be free too.
	path[0] = 0
	jmax := len(pattern) - 1
	for i := 0; i < len(sequence); i++ {
		for j := 0; j < jmax; j++ {

			// Mismatch score = -1
			// Match score = 0
			match := -1
			if _samenuc(pattern[j], sequence[i]) {
				match = 0
			}

			idx := buffIndex(i, j, width)

			diag := buffer[buffIndex(i-1, j-1, width)] + match

			// Each gap cost -1
			left := buffer[buffIndex(i, j-1, width)] - 1
			up := buffer[buffIndex(i-1, j, width)] - 1

			score := max(diag, up, left)

			buffer[idx] = score

			switch {
			case score == left:
				path[idx] = -1
			case score == diag:
				path[idx] = 0
			case score == up:
				path[idx] = +1
			}
		}
	}

	// Fills the last column considering the free up gap
	for i := 0; i < len(sequence); i++ {
		idx := buffIndex(i, jmax, width)

		// Mismatch score = -1
		// Match score = 0
		match := -1
		if _samenuc(pattern[jmax], sequence[i]) {
			match = 0
		}

		diag := buffer[buffIndex(i-1, jmax-1, width)] + match
		left := buffer[buffIndex(i, jmax-1, width)] - 1
		up := buffer[buffIndex(i-1, jmax, width)]

		score := max(diag, up, left)
		buffer[idx] = score

		switch {
		case score == left:
			path[idx] = -1
		case score == diag:
			path[idx] = 0
		case score == up:
			path[idx] = +1
		}
	}

	// Bactracking of the aligment

	i := len(sequence) - 1
	j := jmax
	end := -1
	lali := 0
	for j > 0 { // C'était i > -1 && j > 0
		lali++
		switch path[buffIndex(i, j, width)] {
		case 0:
			j--
			if end == -1 {
				end = i
				lali = 1
			}
			i--
		case 1:
			i--
		case -1:
			j--
			if end == -1 {
				end = i
				lali = 1
			}
		}

	}

	// log.Warnf("from : %d to: %d error: %d match: %v",
	// 	i, end+1, -buffer[buffIndex(len(sequence)-1, len(pattern)-1, width)],
	// 	string(sequence[i:(end+1)]))
	return i, end + 1, -buffer[buffIndex(len(sequence)-1, len(pattern)-1, width)]
}
