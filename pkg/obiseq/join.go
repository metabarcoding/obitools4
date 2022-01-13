package obiseq

import "git.metabarcoding.org/lecasofts/go/obitools/pkg/goutils"

func (sequence BioSequence) Join(seq2 BioSequence, copy_annot bool) (BioSequence, error) {

	new_seq := MakeEmptyBioSequence()
	new_seq.SetId(sequence.Id())
	new_seq.SetDefinition(sequence.Definition())

	new_seq.Write(sequence.Sequence())
	new_seq.Write(seq2.Sequence())

	if copy_annot {
		goutils.CopyMap(new_seq.Annotations(), sequence.Annotations())
	}

	return new_seq, nil
}
