package obiseq

func (sequence *BioSequence) Join(seq2 *BioSequence, inplace bool) *BioSequence {

	if !inplace {
		sequence = sequence.Copy()
	}

	sequence.Write(seq2.Sequence())

	return sequence
}
