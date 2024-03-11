package obiseq

// ".ABCDEFGHIJKLMNOPQRSTUVWXYZ#![]"
var _revcmpDNA = []byte(".TVGHNNCDNNMNKNNNNYSAABWNRN]N[NNN")

// nucComplement returns the complement of a nucleotide.
//
// It takes a byte as input and returns the complement of the nucleotide.
// The function handles various cases:
//   - If the input is '.' or '-', it returns the same character.
//   - If the input is '[', it returns ']'.
//   - If the input is ']', it returns '['.
//   - If the input is a letter from 'A' to 'z', it returns the complement of the nucleotide.
//     The complement is calculated using the _revcmpDNA lookup table.
//   - If none of the above cases match, it returns 'n'.
func nucComplement(n byte) byte {
	switch {
	case n == '.' || n == '-':
		return n
	case n == '[':
		return ']'
	case n == ']':
		return '['
	case (n >= 'A' && n <= 'z'):
		return _revcmpDNA[n&31] | 0x20
	}
	return 'n'
}

// ReverseComplement reverses and complements a BioSequence.
//
// If `inplace` is `false`, a new copy of the BioSequence is created before
// performing the reverse complement. If `inplace` is `true`, the reverse
// complement is performed directly on the original BioSequence.
//
// The function first reverses the sequence by swapping the characters from the
// beginning and end of the sequence. Then, it complements each character in the
// sequence by finding its complement using the `nucComplement` function.
//
// If the BioSequence has qualities, the function also reverse the qualities in
// the same way as the sequence.
//
// The function returns the reverse complemented BioSequence.
func (sequence *BioSequence) ReverseComplement(inplace bool) *BioSequence {

	if sequence == nil {
		return nil
	}

	if !inplace {
		sequence = sequence.Copy()
	}

	s := sequence.sequence

	for i, j := sequence.Len()-1, 0; i >= j; i-- {

		// ASCII code & 31 -> builds an index in witch (a|A) is 1
		// ASCII code & 0x20 -> Foce lower case

		s[j], s[i] = nucComplement(s[i]), nucComplement(s[j])
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
		b[1], b[9] = nucComplement(b[9]), nucComplement(b[1])

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

/**
* ReverseComplementWorker is a function that returns a SeqWorker which performs reverse complement operation on given BioSequence.
* @param inplace {bool}: If true, changes will be made to original sequence object else new sequence object will be created. Default value is false.
* @returns {SeqWorker} A function that accepts *BioSequence and returns its reversed-complement form.
 */
func ReverseComplementWorker(inplace bool) SeqWorker {
	f := func(input *BioSequence) (BioSequenceSlice, error) {
		return BioSequenceSlice{input.ReverseComplement(inplace)}, nil
	}

	return f
}
