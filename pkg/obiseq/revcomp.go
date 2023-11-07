package obiseq

// ".ABCDEFGHIJKLMNOPQRSTUVWXYZ#![]"
var _revcmpDNA = []byte(".TVGHNNCDNNMNKNNNNYSAABWNRN]N[NNN")

func complement(n byte) byte {
	switch {
	case n == '.' || n == '-':
		return n
	case (n >= 'A' && n <= 'z'):
		return _revcmpDNA[n&31] | (n & 0x20)
	}
	return 'n'
}

// Reverse complements a DNA sequence.
// If the inplace parametter is true, that operation is done in place.
func (sequence *BioSequence) ReverseComplement(inplace bool) *BioSequence {

	if !inplace {
		sequence = sequence.Copy()
	}

	s := sequence.sequence

	for i, j := sequence.Len()-1, 0; i >= j; i-- {

		// ASCII code & 31 -> builds an index in witch (a|A) is 1
		// ASCII code & 0x20 -> Foce lower case

		s[j], s[i] = complement(s[i]), complement(s[j])
		j++
	}

	if sequence.HasQualities() {
		s := sequence.qualities
		for i, j := sequence.Len()-1, 0; i >= j; i-- {
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
		b[1], b[9] = complement(b[9]), complement(b[1])

		// Exchange sequencing scores
		b[3], b[4], b[11], b[12] = b[11], b[12], b[3], b[4]

		return string(b)
	}

	lseq := sequence.Len()

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

func ReverseComplementWorker(inplace bool) SeqWorker {
	f := func(input *BioSequence) *BioSequence {
		return input.ReverseComplement(inplace)
	}

	return f
}
