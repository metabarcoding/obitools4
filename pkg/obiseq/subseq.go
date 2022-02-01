package obiseq

import (
	"errors"
	"fmt"
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

	var newSeq BioSequence

	if from < to {
		newSeq = MakeEmptyBioSequence()
		newSeq.Write(sequence.Sequence()[from:to])

		if sequence.HasQualities() {
			newSeq.WriteQualities(sequence.Qualities()[from:to])
		}

		newSeq.sequence.id = fmt.Sprintf("%s_sub[%d..%d]", sequence.Id(), from+1, to)
		newSeq.sequence.definition = sequence.sequence.definition
	} else {
		newSeq, _ = sequence.Subsequence(from, sequence.Length(), false)
		newSeq.Write(sequence.Sequence()[0:to])

		if sequence.HasQualities() {
			newSeq.WriteQualities(sequence.Qualities()[0:to])
		}

	}

	if len(sequence.Annotations()) > 0 {
		newSeq.sequence.annotations = GetAnnotation(sequence.Annotations())
	}

	return newSeq, nil
}
