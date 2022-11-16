package obialign

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/goutils"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

const wsize = 16
const dwsize = wsize * 2

// Out values are always the smallest
// Among in values, they rank according to their score
// For equal score the shortest path is the best
func encodeValues(score, length int, out bool) uint64 {
	const mask = uint64(1<<wsize) - 1
	us := uint64(score)
	fo := (us << wsize) | (uint64((^length)-1) & mask)
	if !out {
		fo |= (uint64(1) << dwsize)
	}
	return fo
}

func _isout(value uint64) bool {
	const outmask = uint64(1) << dwsize
	return (value & outmask) == 0
}

func _lpath(value uint64) int {
	const mask = uint64(1<<wsize) - 1
	return int(((value + 1) ^ mask) & mask)
}

func decodeValues(value uint64) (int, int, bool) {
	const mask = uint64(1<<wsize) - 1
	const outmask = uint64(1) << dwsize
	score := int((value >> wsize) & mask)
	length := int(((value + 1) ^ mask) & mask)
	out := (value & outmask) == 0
	return score, length, out
}

func _incpath(value uint64) uint64 {
	return value - 1
}

func _incscore(value uint64) uint64 {
	const incr = uint64(1) << wsize
	return value + incr
}

func _setout(value uint64) uint64 {
	const outmask = ^(uint64(1) << dwsize)
	return value & outmask
}

var _empty = encodeValues(0, 0, false)
var _out = encodeValues(0, 30000, true)
var _notavail = encodeValues(0, 30000, false)

func FastLCSScore(seqA, seqB *obiseq.BioSequence, maxError int, buffer *[]uint64) (int, int) {

	lA := seqA.Length()
	lB := seqB.Length()

	// Ensure that A is the longest
	if lA < lB {
		seqA, seqB = seqB, seqA
		lA, lB = lB, lA
	}

	if maxError == -1 {
		maxError = lA*2
	}

	delta := lA - lB

	// The difference of length is larger the maximum allowed errors
	if delta > maxError {
		return -1, -1
	}

	// Doit-on vraiment diviser par deux ??? pas certain
	extra := (maxError - delta) + 1

	even := 1 + delta + 2*extra
	width := 2*even - 1

	if buffer == nil {
		var local []uint64
		buffer = &local
	}

	if cap(*buffer) < 2*width {
		*buffer = make([]uint64, 3*width)
	}

	previous := (*buffer)[0:width]
	current := (*buffer)[width:(2 * width)]

	previous[extra] = _empty
	previous[extra+even] = encodeValues(0, 1, false)
	previous[extra+even-1] = encodeValues(0, 1, false)

	N := lB + ((delta) >> 1)

	bA := seqA.Sequence()
	bB := seqB.Sequence()

	// log.Println("N = ", N)

	for y := 1; y <= N; y++ {
		// in_matrix := false
		x1 := y - lB + extra
		x2 := extra - y
		xs := goutils.MaxInt(goutils.MaxInt(x1, x2), 0)

		x1 = y + extra
		x2 = lA + extra - y
		xf := goutils.MinInt(goutils.MinInt(x1, x2), even-1) + 1

		for x := xs; x < xf; x++ {

			i := y - x + extra
			j := y + x - extra

			var Sdiag, Sleft, Sup uint64

			switch {

			case i == 0:
				Sup = _notavail
				Sdiag = _notavail
				Sleft = encodeValues(0, j-1, false)
			case j == 0:
				Sup = encodeValues(0, i-1, false)
				Sdiag = _notavail
				Sleft = _notavail
			default:
				Sdiag = previous[x]

				if bA[j-1] == bB[i-1] {
					Sdiag = _incscore(Sdiag)
				}

				if x < (even - 1) {
					Sup = previous[x+even]
				} else {
					Sup = _out
				}
				if x > 0 {
					Sleft = previous[x+even-1]
				} else {
					Sleft = _out
				}
			}

			var score uint64
			switch {
			case Sdiag >= Sup && Sdiag >= Sleft:
				score = Sdiag
			case Sup >= Sleft:
				score = Sup
			default:
				score = Sleft
			}

			if _isout(Sdiag) || _isout(Sup) || _isout(Sleft) {
				score = _setout(score)
			}

			current[x] = _incpath(score)
		}
		// . 9   10 + 2 - 1
		x1 = y - lB + extra + even
		x2 = extra - y + even - 1
		xs = goutils.MaxInt(goutils.MaxInt(x1, x2), even)

		x1 = y + extra + even
		x2 = lA + extra - y + even - 1
		xf = goutils.MinInt(goutils.MinInt(x1, x2), width-1) + 1

		for x := xs; x < xf; x++ {

			i := y - x + extra + even
			j := y + x - extra - even + 1

			var Sdiag, Sleft, Sup uint64

			switch {

			case i == 0:
				Sup = _notavail
				Sdiag = _notavail
				Sleft = encodeValues(0, j-1, false)
			case j == 0:
				Sup = encodeValues(0, i-1, false)
				Sdiag = _notavail
				Sleft = _notavail
			default:
				Sdiag = previous[x]

				if bA[j-1] == bB[i-1] {
					Sdiag = _incscore(Sdiag)
				}

				Sleft = current[x-even]
				Sup = current[x-even+1]

			}

			var score uint64
			switch {
			case Sdiag >= Sup && Sdiag >= Sleft:
				score = Sdiag
			case Sup >= Sleft:
				score = Sup
			default:
				score = Sleft
			}

			if _isout(Sdiag) || _isout(Sup) || _isout(Sleft) {
				score = _setout(score)
			}

			current[x] = _incpath(score)
		}

		previous, current = current, previous

	}

	s, l, o := decodeValues(previous[(delta%2)*even+extra+(delta>>1)])

	if o {
		return -1, -1
	}

	return s, l
}
