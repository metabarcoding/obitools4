package obialign

func buffIndex(i, j, width int) int {
	return (i+1)*width + (j + 1)
}
func LocatePattern(pattern, sequence []byte) (int, int, int) {
	width := len(pattern) + 1
	buffsize := (len(pattern) + 1) * (len(sequence) + 1)
	buffer := make([]int, buffsize)
	path := make([]int, buffsize)

	for j := 0; j < len(pattern); j++ {
		idx := buffIndex(-1, j, width)
		buffer[idx] = -j - 1
		path[idx] = -1
	}

	for i := -1; i < len(sequence); i++ {
		idx := buffIndex(i, -1, width)
		buffer[idx] = 0
		path[idx] = +1
	}

	path[0] = 0
	jmax := len(pattern) - 1
	for i := 0; i < len(sequence); i++ {
		for j := 0; j < jmax; j++ {
			match := -1
			if _samenuc(pattern[j], sequence[i]) {
				match = 0
			}

			idx := buffIndex(i, j, width)

			diag := buffer[buffIndex(i-1, j-1, width)] + match
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

	for i := 0; i < len(sequence); i++ {
		idx := buffIndex(i, jmax, width)

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

	i := len(sequence) - 1
	j := jmax
	end := -1
	lali := 0
	for i > -1 && j > 0 {
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
	return i, end + 1, -buffer[buffIndex(len(sequence)-1, len(pattern)-1, width)]
}
