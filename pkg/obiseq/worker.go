package obiseq

import log "github.com/sirupsen/logrus"

type SeqAnnotator func(*BioSequence)

type SeqWorker func(*BioSequence) *BioSequence
type SeqSliceWorker func(BioSequenceSlice) BioSequenceSlice

func NilSeqWorker(seq *BioSequence) *BioSequence {
	return seq
}

func AnnotatorToSeqWorker(function SeqAnnotator) SeqWorker {
	f := func(seq *BioSequence) *BioSequence {
		function(seq)
		return seq
	}
	return f
}

func SeqToSliceWorker(worker SeqWorker,
	inplace, breakOnError bool) SeqSliceWorker {
	var f SeqSliceWorker

	if worker == nil {
		if inplace {
			f = func(input BioSequenceSlice) BioSequenceSlice {
				return input
			}
		} else {
			f = func(input BioSequenceSlice) BioSequenceSlice {
				output := MakeBioSequenceSlice(len(input))
				copy(output, input)
				return output
			}
		}
	} else {
		f = func(input BioSequenceSlice) BioSequenceSlice {
			output := input
			if !inplace {
				output = MakeBioSequenceSlice(len(input))
			}
			i := 0
			for _, s := range input {
				r := worker(s)
				if r != nil {
					output[i] = r
					i++
				} else if breakOnError {
					log.Fatalf("got an error on sequence %s processing",
						r.Id())
				}
			}

			return output[0:i]
		}

	}

	return f
}

func SeqToSliceConditionalWorker(worker SeqWorker,
	condition SequencePredicate,
	inplace, breakOnError bool) SeqSliceWorker {

	if condition == nil {
		return SeqToSliceWorker(worker, inplace, breakOnError)
	}

	f := func(input BioSequenceSlice) BioSequenceSlice {
		output := input
		if !inplace {
			output = MakeBioSequenceSlice(len(input))
		}

		i := 0

		for _, s := range input {
			if condition(s) {
				r := worker(s)
				if r != nil {
					output[i] = r
					i++
				} else if breakOnError {
					log.Fatalf("got an error on sequence %s processing",
						r.Id())
				}
			}
		}

		return output[0:i]
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

	f := func(seq *BioSequence) *BioSequence {
		if seq == nil {
			return nil
		}
		return next(worker(seq))
	}

	return f
}
