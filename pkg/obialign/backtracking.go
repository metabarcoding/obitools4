package obialign

import (
	log "github.com/sirupsen/logrus"
	"slices"
)

func _Backtracking(pathMatrix []int, lseqA, lseqB int, path *[]int) []int {

	needed := (lseqA + lseqB) * 2
	(*path) = (*path)[:0]
	cp := cap(*path)
	(*path) = slices.Grow((*path), needed)
	if cp < cap(*path) {
		log.Infof("Resized path from %d to %d\n", cp, cap(*path))
	}
	p := cap(*path)
	*path = (*path)[:p]

	i := lseqA - 1
	j := lseqB - 1

	ldiag := 0
	lup := 0
	lleft := 0

	for i > -1 || j > -1 {
		step := _GetMatrix(&pathMatrix, lseqA, i, j)
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

	return (*path)[p:cap((*path))]
}
