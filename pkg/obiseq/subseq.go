package obiseq

import (
	"errors"
	"fmt"
)

// Returns a sub sequence start from position 'from' included,
// to position 'to' excluded. Coordinates start at position 0.
func (sequence *BioSequence) Subsequence(from, to int, circular bool) (*BioSequence, error) {

	if from >= to && !circular {
		return nil, errors.New("from greater than to")
	}

	if from < 0 || from >= sequence.Length() {
		return nil, errors.New("from out of bounds")
	}

	if to <= 0 || to > sequence.Length() {
		return nil, errors.New("to out of bounds")
	}

	var newSeq *BioSequence

	if from < to {
		newSeq = NewEmptyBioSequence()
		newSeq.Write(sequence.Sequence()[from:to])

		if sequence.HasQualities() {
			newSeq.WriteQualities(sequence.Qualities()[from:to])
		}

		newSeq.id = fmt.Sprintf("%s_sub[%d..%d]", sequence.Id(), from+1, to)
		newSeq.definition = sequence.definition
	} else {
		newSeq, _ = sequence.Subsequence(from, sequence.Length(), false)
		newSeq.Write(sequence.Sequence()[0:to])

		if sequence.HasQualities() {
			newSeq.WriteQualities(sequence.Qualities()[0:to])
		}

	}

	if sequence.HasAnnotation() {
		newSeq.annotations = GetAnnotation(sequence.Annotations())
	}

	return newSeq._subseqMutation(from), nil
}

func (sequence *BioSequence) _subseqMutation(shift int) *BioSequence {

	lseq := sequence.Length()

	mut, ok := sequence.GetIntMap("pairing_mismatches")
	if ok && len(mut) > 0 {
		cmut := make(map[string]int, len(mut))

		for m, p := range mut {
			if p < lseq {
				cmut[m] = p - shift
			}
		}

		sequence.SetAttribute("pairing_mismatches", cmut)
	}

	return sequence

}
