package obialign

func __backtracking__(path_matrix []int, lseqA, lseqB int, path *[]int) []int {

	needed := (lseqA + lseqB) * 2

	if needed > cap(*path) {
		*path = make([]int, 0, needed)
	}

	*path = (*path)[:cap(*path)]
	p := cap(*path)

	i := lseqA - 1
	j := lseqB - 1

	ldiag := 0
	lup := 0
	lleft := 0

	for i > -1 || j > -1 {
		step := __get_matrix__(&path_matrix, lseqA, i, j)
		// log.Printf("I: %d J:%d -> %d\n", i, j, step)

		switch {
		case step == 0:
			if lleft != 0 {
				p--
				(*path)[p] = ldiag
				p--
				(*path)[p] = lleft
				lleft = 0
				ldiag = 0
			}
			if lup != 0 {
				p--
				(*path)[p] = ldiag
				p--
				(*path)[p] = lup
				lup = 0
				ldiag = 0
			}
			ldiag++
			i--
			j--
		case step > 0:
			if lup != 0 {
				p--
				(*path)[p] = ldiag
				p--
				(*path)[p] = lup
				lup = 0
				ldiag = 0
			}
			lleft += step
			j -= step
		case step < 0:
			if lleft != 0 {
				p--
				(*path)[p] = ldiag
				p--
				(*path)[p] = lleft
				lleft = 0
				ldiag = 0
			}
			lup += step
			i += step
		}
	}

	if lleft != 0 {
		p--
		(*path)[p] = ldiag
		p--
		(*path)[p] = lleft
		ldiag = 0
	}
	if lup != 0 {
		p--
		(*path)[p] = ldiag
		p--
		(*path)[p] = lup
		ldiag = 0
	}
	if ldiag != 0 {
		p--
		(*path)[p] = ldiag
		p--
		(*path)[p] = 0
	}

	*path = (*path)[p:cap((*path))]

	return *path
}
