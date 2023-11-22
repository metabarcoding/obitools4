package obiseq

import (
	"fmt"
)

// Subsequence returns a subsequence of the BioSequence.
//
// Parameters:
// - from: starting position of the subsequence.
// - to: ending position of the subsequence.
// - circular: indicates whether the subsequence should be circular.
//
// Return:
// - *BioSequence: the subsequence of the BioSequence.
// - error: an error if the subsequence parameters are invalid.
func (sequence *BioSequence) Subsequence(from, to int, circular bool) (*BioSequence, error) {
	if from >= to && !circular {
		return nil, fmt.Errorf("from: %d greater than to: %d", from, to)
	}

	if from < 0 {
		return nil, fmt.Errorf("from out of bounds %d < 0", from)
	}

	if from >= sequence.Len() {
		return nil,
			fmt.Errorf("from out of bounds %d >= %d", from, sequence.Len())
	}

	if to > sequence.Len() {
		return nil,
			fmt.Errorf("to out of bounds %d > %d", to, sequence.Len())
	}

	var newSeq *BioSequence

	if from < to {
		newSeq = NewEmptyBioSequence(0)
		newSeq.sequence = CopySlice(sequence.Sequence()[from:to])

		if sequence.HasQualities() {
			newSeq.qualities = CopySlice(sequence.Qualities()[from:to])
		}

		newSeq.id = fmt.Sprintf("%s_sub[%d..%d]", sequence.Id(), from+1, to)
		//	newSeq.definition = sequence.definition
	} else {
		newSeq, _ = sequence.Subsequence(from, sequence.Len(), false)
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

	lseq := sequence.Len()

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
