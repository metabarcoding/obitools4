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
	breakOnError bool) SeqSliceWorker {
	var f SeqSliceWorker

	if worker == nil {
		f = func(input BioSequenceSlice) (BioSequenceSlice, error) {
			return input, nil
		}
	} else {
		f = func(input BioSequenceSlice) (BioSequenceSlice, error) {
			output := MakeBioSequenceSlice(len(input))
			i := 0
			for _, s := range input {
				r, err := worker(s)
				if err == nil {
					for _, rs := range r {
						if i == len(output) {
							output = slices.Grow(output, cap(output))
							output = output[:cap(output)]
						}
						output[i] = rs
						i++
					}

				} else {
					if breakOnError {
						err = fmt.Errorf("got an error on sequence %s processing : %v",
							s.Id(), err)
						return BioSequenceSlice{}, err
					} else {
						log.Warnf("got an error on sequence %s processing : %v",
							s.Id(), err)
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
	worker SeqWorker, breakOnError bool) SeqSliceWorker {

	if condition == nil {
		return SeqToSliceWorker(worker, breakOnError)
	}

	f := func(input BioSequenceSlice) (BioSequenceSlice, error) {
		output := MakeBioSequenceSlice(len(input))

		i := 0

		for _, s := range input {
			if condition(s) {
				r, err := worker(s)
				if err == nil {
					for _, rs := range r {
						if i == len(output) {
							output = slices.Grow(output, cap(output))
							output = output[:cap(output)]
						}
						output[i] = rs
						i++
					}
				} else {
					if breakOnError {
						err = fmt.Errorf("got an error on sequence %s processing : %v",
							s.Id(), err)
						return BioSequenceSlice{}, err
					} else {
						log.Warnf("got an error on sequence %s processing : %v",
							s.Id(), err)
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

	sw := SeqToSliceWorker(next, false)

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
