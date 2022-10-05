package obiseq

// ".ABCDEFGHIJKLMNOPQRSTUVWXYZ#![]"
var __revcmp_dna__ = []byte(".TVGHEFCDIJMLKNOPQYSAABWXRZ#!][")

// Reverse complements a DNA sequence.
// If the inplace parametter is true, that operation is done in place.
func (sequence *BioSequence) ReverseComplement(inplace bool) *BioSequence {

	if !inplace {
		sequence = sequence.Copy()
	}

	s := sequence.sequence

	for i, j := sequence.Length()-1, 0; i >= j; i-- {

		s[j], s[i] = __revcmp_dna__[s[i]&31]|(s[i]&0x20),
			__revcmp_dna__[s[j]&31]|(s[j]&0x20)
		j++
	}

	if sequence.HasQualities() {
		s := sequence.qualities
		for i, j := sequence.Length()-1, 0; i >= j; i-- {
			s[j], s[i] = s[i], s[j]
			j++
		}
	}

	return sequence
}
