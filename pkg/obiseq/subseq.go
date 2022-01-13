package obiseq

import (
	"errors"
	"fmt"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/goutils"
)

// Returns a sub sequence start from position 'from' included,
// to position 'to' excluded. Coordinates start at position 0.
func (sequence BioSequence) Subsequence(from, to int, circular bool) (BioSequence, error) {

	if from >= to && !circular {
		return NilBioSequence, errors.New("from greater than to")
	}

	if from < 0 || from >= sequence.Length() {
		return NilBioSequence, errors.New("from out of bounds")
	}

	if to <= 0 || to > sequence.Length() {
		return NilBioSequence, errors.New("to out of bounds")
	}

	var new_seq BioSequence

	if from < to {
		new_seq = MakeEmptyBioSequence()
		new_seq.Write(sequence.Sequence()[from:to])
		fmt.Fprintf(&new_seq.sequence.id, "%s_sub[%d..%d]", sequence.Id(), from+1, to)
		new_seq.sequence.definition.Write(sequence.sequence.definition.Bytes())
	} else {
		new_seq, _ = sequence.Subsequence(from, sequence.Length(), false)
		new_seq.Write(sequence.Sequence()[0:to])
	}

	if len(sequence.Annotations()) > 0 {
		goutils.CopyMap(new_seq.Annotations(), sequence.Annotations())
	}

	return new_seq, nil
}
