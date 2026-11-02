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

// IUPACNucMask returns the set of nucleotides denoted by the IUPAC code b
// (case insensitive, U is read as T) as a 4 bits mask: bit 0 is A, bit 1 is C,
// bit 2 is G and bit 3 is T. It returns 0 if b is not an IUPAC nucleotide code.
func IUPACNucMask(b byte) byte {
	if (b >= 'A') && (b <= 'Z') {
		b |= 32
	}
	if (b < 'a') || (b > 'z') {
		return 0
	}
	return _iupac[b-'a']
}
