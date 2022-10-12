package obiseq

// ".ABCDEFGHIJKLMNOPQRSTUVWXYZ#![]"
var _revcmpDNA = []byte(".TVGHEFCDIJMLKNOPQYSAABWXRZ#!][")

// Reverse complements a DNA sequence.
// If the inplace parametter is true, that operation is done in place.
func (sequence *BioSequence) ReverseComplement(inplace bool) *BioSequence {

	if !inplace {
		sequence = sequence.Copy()
	}

	s := sequence.sequence

	for i, j := sequence.Length()-1, 0; i >= j; i-- {

		// ASCII code & 31 -> builds an index in witch (a|A) is 1
		// ASCII code & 0x20 -> Foce lower case

		s[j], s[i] = _revcmpDNA[s[i]&31]|(s[i]&0x20),
			_revcmpDNA[s[j]&31]|(s[j]&0x20)
		j++
	}

	if sequence.HasQualities() {
		s := sequence.qualities
		for i, j := sequence.Length()-1, 0; i >= j; i-- {
			s[j], s[i] = s[i], s[j]
			j++
		}
	}

	return sequence._revcmpMutation()
}

func (sequence *BioSequence) _revcmpMutation() *BioSequence {

	rev := func(m string) string {
		b := []byte(m)

		// Echange and reverse complement symboles
		b[1], b[9] = _revcmpDNA[b[9]&31]|(b[9]&0x20),
			_revcmpDNA[b[1]&31]|(b[1]&0x20)

		// Exchange sequencing scores
		b[3], b[4], b[11], b[12] = b[11], b[12], b[3], b[4]

		return string(b)
	}

	lseq := sequence.Length()

	mut, ok := sequence.GetIntMap("pairing_mismatches")
	if ok && len(mut) > 0 {
		cmut := make(map[string]int, len(mut))

		for m, p := range mut {
			cmut[rev(m)] = lseq - p + 1
		}

		sequence.SetAttribute("pairing_mismatches", cmut)
	}

	return sequence
}
