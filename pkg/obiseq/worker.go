package obiseq

import (
	"fmt"
	"slices"

	log "github.com/sirupsen/logrus"
)

type SeqAnnotator func(*BioSequence)

type SeqWorker func(*BioSequence) (BioSequenceSlice, error)
type SeqSliceWorker func(BioSequenceSlice) (BioSequenceSlice, error)

func NilSeqWorker(seq *BioSequence) (BioSequenceSlice, error) {
	return BioSequenceSlice{seq}, nil
}

func AnnotatorToSeqWorker(function SeqAnnotator) SeqWorker {
	f := func(seq *BioSequence) (BioSequenceSlice, error) {
		function(seq)
		return BioSequenceSlice{seq}, nil
	}
	return f
}

func SeqToSliceWorker(worker SeqWorker,
	inplace, breakOnError bool) SeqSliceWorker {
	var f SeqSliceWorker

	if worker == nil {
		if inplace {
			f = func(input BioSequenceSlice) (BioSequenceSlice, error) {
				return input, nil
			}
		} else {
			f = func(input BioSequenceSlice) (BioSequenceSlice, error) {
				output := MakeBioSequenceSlice(len(input))
				copy(output, input)
				return output, nil
			}
		}
	} else {
		f = func(input BioSequenceSlice) (BioSequenceSlice, error) {
			output := input
			if !inplace {
				output = MakeBioSequenceSlice(len(input))
			}
			i := 0
			for _, s := range input {
				r, err := worker(s)
				if err == nil {
					for _, rs := range r {
						output[i] = rs
						i++
						if i == cap(output) {
							slices.Grow(output, cap(output))
						}
					}

				} else {
					if breakOnError {
						err = fmt.Errorf("got an error on sequence %s processing : %v",
							s.Id(), err)
						return BioSequenceSlice{}, err
					} else {
						log.Warnf("got an error on sequence %s processing",
							s.Id())
					}
				}
			}

			return output[0:i], nil
		}

	}

	return f
}

func SeqToSliceConditionalWorker(
	condition SequencePredicate,
	worker SeqWorker,
	inplace, breakOnError bool) SeqSliceWorker {

	if condition == nil {
		return SeqToSliceWorker(worker, inplace, breakOnError)
	}

	f := func(input BioSequenceSlice) (BioSequenceSlice, error) {
		output := input
		if !inplace {
			output = MakeBioSequenceSlice(len(input))
		}

		i := 0

		for _, s := range input {
			if condition(s) {
				r, err := worker(s)
				if err == nil {
					for _, rs := range r {
						output[i] = rs
						i++
						if i == cap(output) {
							slices.Grow(output, cap(output))
						}
					}
				} else {
					if breakOnError {
						err = fmt.Errorf("got an error on sequence %s processing : %v",
							s.Id(), err)
						return BioSequenceSlice{}, err
					} else {
						log.Warnf("got an error on sequence %s processing",
							s.Id())
					}
				}
			}
		}

		return output[0:i], nil
	}

	return f
}

func (worker SeqWorker) ChainWorkers(next SeqWorker) SeqWorker {
	if worker == nil {
		return next
	} else {
		if next == nil {
			return worker
		}
	}

	sw := SeqToSliceWorker(next, true, false)

	f := func(seq *BioSequence) (BioSequenceSlice, error) {
		if seq == nil {
			return BioSequenceSlice{}, nil
		}
		slice, err := worker(seq)
		if err == nil {
			slice, err = sw(slice)
		}
		return slice, err
	}

	return f
}
