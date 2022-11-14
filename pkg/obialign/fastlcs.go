package obialign

import (
	"log"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

// Out values are always the smallest
// Among in values, they rank according to their score
// For equal score the shortest path is the best
func _encodeValues(score, length int, out bool) uint64 {
	const mask = uint64(1<<16) - 1
	us := uint64(score)
	fo := (us << 16) | (uint64((^length)-1) & mask)
	if !out {
		fo |= (uint64(1) << 32)
	}
	return fo
}

func _decodeValues(value uint64) (int, int, bool) {
	const mask = uint64(1<<16) - 1
	score := int((value >> 16) & mask)
	length := int(((value + 1) ^ mask) & mask)
	out := (value & (1 << 32)) == 0
	return score, length, out
}

func _incpath(value uint64) uint64 {
	return value - 1
}

func _incscore(value uint64) uint64 {
	const incr = uint64(1) << 16
	return value + incr
}

func _setout(value uint64) uint64 {
	const outmask = uint64(1) << 32
	return value | outmask
}

func _isout(value uint64) bool {
	const outmask = uint64(1) << 32
	return (value & outmask) == 0
}

var _empty = _encodeValues(0, 0, false)
var _out = _encodeValues(0, 3000, true)

func FastLCSScore(seqA, seqB *obiseq.BioSequence, maxError int) (int, int) {

	// x:=_encodeValues(12,0,false)
	// xs,xl,xo := _decodeValues(x)
	// log.Println(x,xs,xl,xo)
	// x=_encodeValues(12,1,false)
	// xs,xl,xo = _decodeValues(x)
	// log.Println(x,xs,xl,xo)
	// x=_encodeValues(12,2,false)
	// xs,xl,xo = _decodeValues(x)
	// log.Println(x,xs,xl,xo)

	lA := seqA.Length()
	lB := seqB.Length()

	// Ensure that A is the longest
	if lA < lB {
		seqA, seqB = seqB, seqA
		lA, lB = lB, lA
	}

	delta := lA - lB

	// The difference of length is larger the maximum allowed errors
	if delta > maxError {
		//	log.Println("Too large difference of length")
		return -1, -1
	}

	// Doit-on vraiment diviser par deux ??? pas certain
	extra := ((maxError - delta) / 2) + 1

	even := 1 + delta + 2*extra
	width := 2*even - 1

	previous := make([]uint64, width)
	current := make([]uint64, width)

	// Initialise the first line

	for j := 0; j < width; j++ {
		if (j == extra+even) || (j == extra+even-1) {
			previous[j] = _encodeValues(0, 1, false)
		} else {
			previous[j] = _empty
		}
	}

	N := lB + ((delta) >> 1)
	X := width

	// log.Println("N = ", N)

	for y := 1; y <= N; y++ {
		// in_matrix := false
		for x := 0; x < width; x++ {
			X = y * width + x
			modulo := X % width
			begin := X - modulo
			quotien := begin / width

			i := quotien + extra - modulo
			j := quotien - extra + modulo

			if x >= even {
				i += even
				j += 1 - even
			}

			//log.Println(X, i, j, i < 0 || j < 0 || i > lB || j > lA)

			// We are out of tha alignement matrix
			if i < 0 || j < 0 || i > lB || j > lA {
				X++
				continue
			}

			var Sdiag, Sleft, Sup uint64

			switch {

			case i == 0:
				Sup = _encodeValues(0, 30000, false)
				Sdiag = _encodeValues(0, 30000, false)
				Sleft = _encodeValues(0, j-1, false)
			case j == 0:
				Sup = _encodeValues(0, i-1, false)
				Sdiag = _encodeValues(0, 30000, false)
				Sleft = _encodeValues(0, 30000, false)
			default:
				Sdiag = previous[x]
				if x >= even {
					Sleft = current[x-even]
					Sup = current[x-even+1]
				} else {
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
			}

			// log.Println("scores @",i,j,": ",Sdiag,Sup,Sleft)
			// ds,dl,ol := _decodeValues(Sdiag)
			// log.Println(ds,dl,ol)
			// ds,dl,ol = _decodeValues(Sup)
			// log.Println(ds,dl,ol)
			// ds,dl,ol = _decodeValues(Sleft)
			// log.Println(ds,dl,ol)
			var score uint64
			switch {
			case Sdiag >= Sup && Sdiag >= Sleft:
				score = Sdiag
				if seqA.Sequence()[j-1] == seqB.Sequence()[i-1] {
					score = _incscore(score)
				}
			case Sup >= Sleft:
				score = Sup
			default:
				score = Sleft
			}

			if _isout(Sdiag) || _isout(Sup) || _isout(Sleft) {
				score = _setout(score)
			}

			// if i < 5 && j < 5 {
			// 	ds, dl, ol := _decodeValues(_incpath(score))
			// 	log.Println("@", i, j,":", ds, dl, ol)
			// }

			current[x] = _incpath(score)

			// if i == lB && j == lA {
			// 	s, l, o := _decodeValues(current[x])
			// 	log.Println("Results in ", x, y, "(", i, j, ") values : ", s, l, o)
			// }

			X++
		}
		// if !in_matrix {
		// 	log.Fatalln("Never entred in the matrix", y, "/", N)
		// }
		previous, current = current, previous

	}

	s, l, o := _decodeValues(previous[(delta%2)*even+extra+(delta>>1)])

	if o {
		log.Println("Too much error", s, l, (lA%2)*even+(lA-lB), lA, lB, width, even, N)
		return -1, -1
	}

	return s, l
}

// width * j + modulo + width * extra - width * modulo= X
// i = (X - modulo)/width  - modulo + extra
