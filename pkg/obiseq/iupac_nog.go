package obiseq

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

func SameIUPACNuc(a, b byte) bool {
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
