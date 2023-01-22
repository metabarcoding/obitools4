package obiseq

type SeqAnnotator func(*BioSequence)

type SeqWorker func(*BioSequence) *BioSequence
type SeqSliceWorker func(BioSequenceSlice) BioSequenceSlice

func AnnotatorToSeqWorker(function SeqAnnotator) SeqWorker {
	f := func(seq *BioSequence) *BioSequence {
		function(seq)
		return seq
	}
	return f
}

func SeqToSliceWorker(worker SeqWorker, inplace bool) SeqSliceWorker {
	f := func(input BioSequenceSlice) BioSequenceSlice {
		output := input
		if (! inplace) {
			output = MakeBioSequenceSlice() 
		}
			for i,s := range(input) {
				output[i] = worker(s)
			}

		return output
	}

	return f
}

