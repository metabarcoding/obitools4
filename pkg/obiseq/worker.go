package obiseq

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

func SeqToSliceWorker(worker SeqWorker, inplace bool) SeqSliceWorker {
	var f SeqSliceWorker

	if worker == nil {
		if inplace {
			f = func(input BioSequenceSlice) BioSequenceSlice {
				return input
			}
		} else  {
			f = func(input BioSequenceSlice) BioSequenceSlice {
				output := MakeBioSequenceSlice(len(input))
				copy(output,input)
				return output
			}		
		}
	} else {
		f = func(input BioSequenceSlice) BioSequenceSlice {
			output := input
			if !inplace {
				output = MakeBioSequenceSlice(len(input))
			}
			for i, s := range input {
				output[i] = worker(s)
			}
	
			return output
		}	
	}

	return f
}

func SeqToSliceConditionalWorker(worker SeqWorker,
	condition SequencePredicate,
	inplace bool) SeqSliceWorker {

	if condition == nil {
		return SeqToSliceWorker(worker,inplace)
	}

	f := func(input BioSequenceSlice) BioSequenceSlice {
		output := input
		if !inplace {
			output = MakeBioSequenceSlice(len(input))
		}
		for i, s := range input {
			if condition(s) {
				output[i] = worker(s)
			} else {
				output[i] = s
			}
		}

		return output
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
		return next(worker(seq))
	}

	return f
}
