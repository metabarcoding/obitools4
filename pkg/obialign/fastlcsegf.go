package obialign

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiutils"
	log "github.com/sirupsen/logrus"
)

var _iupac = [26]byte{
	//  a   b  c  d   e  f
	1, 14, 2, 13, 0, 0,
	//  g   h  i  j  k   l
	4, 11, 0, 0, 12, 0,
	//  m   n  o  p  q  r
	3, 15, 0, 0, 0, 5,
	//  s  t  u   v  w  x
	6, 8, 8, 13, 9, 0,
	//  y   z
	10, 0,
}

func _samenuc(a, b byte) bool {
	if (a >= 'A') && (a <= 'Z') {
		a |= 32
	}
	if (b >= 'A') && (b <= 'Z') {
		b |= 32
	}

	if (a >= 'a') && (a <= 'z') && (b >= 'a') && (b <= 'z') {
		return (_iupac[a-'a'] & _iupac[b-'a']) > 0
	}
	return a == b
}
func FastLCSEGFScoreByte(bA, bB []byte, maxError int, endgapfree bool, buffer *[]uint64) (int, int) {

	lA := len(bA)
	lB := len(bB)

	// Ensure that A is the longest
	if lA < lB {
		bA, bB = bB, bA
		lA, lB = lB, lA
	}

	if maxError == -1 {
		maxError = lA * 2
	}

	delta := lA - lB

	if endgapfree {
		maxError += delta
	}

	// The difference of length is larger the maximum allowed errors
	if delta > maxError {
		return -1, -1
	}

	// // BEGINNING OF DEBUG CODE //
	// debug_scores := make([][]int, lB+1)
	// for i := range debug_scores {
	// 	debug_scores[i] = make([]int, lA+1)
	// }

	// debug_path := make([][]int, lB+1)
	// for i := range debug_path {
	// 	debug_path[i] = make([]int, lA+1)
	// }

	// debug_out := make([][]bool, lB+1)
	// for i := range debug_out {
	// 	debug_out[i] = make([]bool, lA+1)
	// }
	// // END OF DEBUG CODE //

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

	previous[extra] = _empty // Initialise cell 0,0
	if endgapfree {          // Initialise cell 0,1
		previous[extra+even] = encodeValues(0, 0, false)
	} else {
		previous[extra+even] = encodeValues(0, 1, false)
	}
	previous[extra+even-1] = encodeValues(0, 1, false) // Initialise cell 1,0

	N := lB + ((delta) >> 1)

	log.Debugln("N = ", N, " delta = ", delta, " extra = ", extra, " maxError = ", maxError)

	for y := 1; y <= N; y++ {
		// in_matrix := false
		x1 := y - lB + extra
		x2 := extra - y
		xs := obiutils.MaxInt(obiutils.MaxInt(x1, x2), 0)

		x1 = y + extra
		x2 = lA + extra - y
		xf := obiutils.MinInt(obiutils.MinInt(x1, x2), even-1) + 1

		for x := xs; x < xf; x++ {

			// i span along B and j along A
			i := y - x + extra
			j := y + x - extra

			//			log.Debugln("Coord : ", i, j)
			var Sdiag, Sleft, Sup uint64

			switch {

			case i == 0:
				Sup = _notavail
				Sdiag = _notavail
				if endgapfree {
					Sleft = encodeValues(0, 0, false)
				} else {
					Sleft = encodeValues(0, j, false)
				}
			case j == 0:
				Sup = encodeValues(0, i, false)
				Sdiag = _notavail
				Sleft = _notavail
			default:
				Sdiag = _incpath(previous[x])
				if _samenuc(bA[j-1], bB[i-1]) {
					Sdiag = _incscore(Sdiag)
				}

				if x < (even - 1) {
					Sup = _incpath(previous[x+even])
				} else {
					Sup = _out
				}
				if x > 0 {
					Sleft = previous[x+even-1]
					if (i > 0 && i < lB) || !endgapfree {
						Sleft = _incpath(Sleft)
					}
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

			// I supose the bug was here
			// if _isout(Sdiag) || _isout(Sup) || _isout(Sleft) {
			// 	score = _setout(score)
			// }

			if x == 0 || x == (even-1) {
				score = _setout(score)
			}
			// // BEGINNING OF DEBUG CODE //
			// if i < 2 && j < 5 {
			// 	log.Debugf("[%d,%d]\n",i,j)
			// 	s, p, o := decodeValues(Sdiag)
			// 	log.Debugf("+Sdiag (%v) : %d, %d, %v\n",Sdiag,s,p,o)
			// 	s, p, o = decodeValues(Sup)
			// 	log.Debugf("+Sup   (%v) : %d, %d, %v\n",Sup,s,p,o)
			// 	s, p, o = decodeValues(Sleft)
			// 	log.Debugf("+Sleft (%v) : %d, %d, %v\n",Sleft,s,p,o)
			// 	s, p, o = decodeValues(score)
			// 	log.Debugf("+score (%v) : %d, %d, %v\n",score,s,p,o)
			// 	log.Debugln("-------------------")
			// }
			// s, p, o := decodeValues(score)
			// debug_scores[i][j] = s
			// debug_path[i][j] = p
			// debug_out[i][j] = o
			// // END OF DEBUG CODE //

			current[x] = score
		}
		// . 9   10 + 2 - 1
		x1 = y - lB + extra + even
		x2 = extra - y + even - 1
		xs = obiutils.MaxInt(obiutils.MaxInt(x1, x2), even)

		x1 = y + extra + even
		x2 = lA + extra - y + even - 1
		xf = obiutils.MinInt(obiutils.MinInt(x1, x2), width-1) + 1

		for x := xs; x < xf; x++ {

			i := y - x + extra + even
			j := y + x - extra - even + 1

			//log.Debugln("Coord : ", i, j)
			var Sdiag, Sleft, Sup uint64

			switch {

			case i == 0:
				Sup = _notavail
				Sdiag = _notavail
				if endgapfree {
					Sleft = encodeValues(0, 0, false)
				} else {
					Sleft = encodeValues(0, j, false)
				}
			case j == 0:
				Sup = encodeValues(0, i, false)
				Sdiag = _notavail
				Sleft = _notavail
			default:
				Sdiag = _incpath(previous[x])
				if _samenuc(bA[j-1], bB[i-1]) {
					Sdiag = _incscore(Sdiag)
				}

				Sleft = current[x-even]
				if (i > 0 && i < lB) || !endgapfree {
					Sleft = _incpath(Sleft)
				}
				Sup = _incpath(current[x-even+1])

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

			// I supose the bug was here
			// if _isout(Sdiag) || _isout(Sup) || _isout(Sleft) {
			// 	score = _setout(score)
			// }

			// // BEGINNING OF DEBUG CODE //
			// if i < 2 && j < 5 {
			// 	log.Debugf("[%d,%d]\n",i,j)
			// 	s, p, o := decodeValues(Sdiag)
			// 	log.Debugf("-Sdiag (%v) : %d, %d, %v\n",Sdiag,s,p,o)
			// 	s, p, o = decodeValues(Sup)
			// 	log.Debugf("-Sup   (%v) : %d, %d, %v\n",Sup,s,p,o)
			// 	s, p, o = decodeValues(Sleft)
			// 	log.Debugf("-Sleft (%v) : %d, %d, %v\n",Sleft,s,p,o)
			// 	s, p, o = decodeValues(score)
			// 	log.Debugf("-score (%v) : %d, %d, %v\n",score,s,p,o)
			// 	log.Debugln("-------------------")
			// }
			// s, p, o := decodeValues(score)
			// debug_scores[i][j] = s
			// debug_path[i][j] = p
			// debug_out[i][j] = o
			// // END OF DEBUG CODE //
			current[x] = score
		}

		previous, current = current, previous

	}

	s, l, o := decodeValues(previous[(delta%2)*even+extra+(delta>>1)])

	// // BEGINNING OF DEBUG CODE //
	// fmt.Printf("%2c\t", ' ')
	// for j := 0; j <= lA; j++ {
	// 	if j > 0 {
	// 		fmt.Printf("%11c\t", bA[j-1])
	// 	} else {
	// 		fmt.Printf("%11c\t", '-')
	// 	}
	// }
	// fmt.Printf("\n")
	// for i := 0; i <= lB; i++ {
	// 	if i > 0 {
	// 		fmt.Printf("%2c\t", bB[i-1])
	// 	} else {
	// 		fmt.Printf("%2c\t", '-')
	// 	}

	// 	for j := 0; j <= lA; j++ {
	// 		fmt.Printf("%2d:%2d:%v\t", debug_scores[i][j],
	// 			debug_path[i][j], debug_out[i][j])
	// 	}
	// 	fmt.Printf("\n")
	// }
	// // end OF DEBUG CODE //

	if o {
		return -1, -1
	}

	return s, l
}

func FastLCSEGFScore(seqA, seqB *obiseq.BioSequence, maxError int, buffer *[]uint64) (int, int) {
	return FastLCSEGFScoreByte(seqA.Sequence(), seqB.Sequence(), maxError, true, buffer)
}

func FastLCSScore(seqA, seqB *obiseq.BioSequence, maxError int, buffer *[]uint64) (int, int) {
	return FastLCSEGFScoreByte(seqA.Sequence(), seqB.Sequence(), maxError, false, buffer)
}