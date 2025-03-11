package obiseq

import (
	"fmt"
	"slices"

	log "github.com/sirupsen/logrus"
)

type SeqAnnotator func(*BioSequence)

type SeqWorker func(*BioSequence) (BioSequenceSlice, error)
type SeqSliceWorker func(BioSequenceSlice) (BioSequenceSlice, error)

// NilSeqWorker returns a BioSequenceSlice containing the input sequence and a nil error value.
// This function is typically used as a placeholder or default worker in SeqToSliceWorker when no specific worker is needed.
//
// Parameters:
//
//	seq *BioSequence: A pointer to a BioSequence struct that needs processing.
//
// Returns:
//
//	BioSequenceSlice, error: This function returns a slice containing the input sequence and an error value. If no error occurred during the operation, it will be nil; otherwise, it will contain details about the error.
func NilSeqWorker(seq *BioSequence) (BioSequenceSlice, error) {
	return BioSequenceSlice{seq}, nil
}

// AnnotatorToSeqWorker is a higher-order function that takes a SeqAnnotator
// function and returns a SeqWorker function. It is used to wrap a sequence
// annotation function and convert it into a worker function that can be used
// in a pipeline or workflow for processing biological sequences.
//
// Parameters:
//
//	function SeqAnnotator: A function that takes a pointer to a BioSequence
//	struct and performs some annotation or processing on the sequence data.
//	The SeqAnnotator type is expected to be a function with the following
//	signature:
//	  func(seq *BioSequence)
//	This function should modify the input BioSequence struct in-place by adding
//	annotations, metadata, or performing any other desired operations.
//
// Returns:
//
//	SeqWorker: A function that takes a pointer to a BioSequence struct and
//	returns a BioSequenceSlice containing the input BioSequence, along with
//	an error value. The SeqWorker type is expected to be a function with the
//	following signature:
//	  func(seq *BioSequence) (BioSequenceSlice, error)
//	The returned SeqWorker function wraps the provided SeqAnnotator function
//	and applies it to the input BioSequence before returning the modified
//	BioSequence in a BioSequenceSlice. The error value is always nil, as the
//	function does not perform any operations that could potentially fail.
func AnnotatorToSeqWorker(function SeqAnnotator) SeqWorker {
	f := func(seq *BioSequence) (BioSequenceSlice, error) {
		function(seq)
		return BioSequenceSlice{seq}, nil
	}
	return f
}

// SeqToSliceWorker is a higher-order function that takes a SeqWorker and a
// boolean value indicating whether to break on error and returns a SeqSliceWorker.
// It can be used in a pipeline or workflow for processing biological sequences,
// applying the provided worker to each element of a BioSequenceSlice and returning
// a new slice.
//
// Parameters:
//
//	worker SeqWorker: A function that takes a pointer to a BioSequence struct and
//	 performs some processing on it.
//	 The signature for this function is func(seq *BioSequence) (BioSequenceSlice, error).
//	 This function should return a modified BioSequence in a BioSequenceSlice along with
//	 an error value indicating whether the operation was successful or not.
//	breakOnError bool: A boolean flag that determines whether to stop processing further
//	 elements in case of an error. If set to true and an error is encountered while
//	 processing any element, it stops processing remaining elements and returns the processed
//	 slice so far along with the encountered error. If set to false, it logs the error and
//	 continues processing remaining elements.
//
// Returns:
//
//	SeqSliceWorker: A function that takes a BioSequenceSlice (a slice of pointers to
//	 BioSequence structs) as input and returns a processed BioSequenceSlice along with
//	 an error value indicating whether the operation was successful or not.
//	 The signature for this function is func(input BioSequenceSlice) (BioSequenceSlice, error).
//	 If breakOnError is set to true and any element processing results in an error, it stops
//	 further processing and returns the processed slice so far along with the encountered error.
//	 Otherwise, it processes all elements and returns them as a single BioSequenceSlice along with
//	 a nil error value.
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

func SeqToSliceFilterOnWorker(condition SequencePredicate,
	breakOnError bool) SeqSliceWorker {

	if condition == nil {
		return func(slice BioSequenceSlice) (BioSequenceSlice, error) {
			return slice, nil
		}
	}

	f := func(input BioSequenceSlice) (BioSequenceSlice, error) {
		output := MakeBioSequenceSlice(len(input))

		i := 0

		for _, s := range input {
			if condition(s) {
				output[i] = s
				i++
			}
		}

		return output[0:i], nil
	}

	return f

}

// SeqToSliceConditionalWorker creates a new SeqSliceWorker that processes each sequence in a slice based on a condition. It takes a SequencePredicate and a worker function as arguments. The worker function is only applied to sequences that satisfy the condition.
// If `condition` is nil, this function just behaves like SeqToSliceWorker with the provided `worker`.
// If `breakOnError` is true, the pipeline will stop and return an error if any sequence processing fails. Otherwise, it will log a warning message for each failed sequence.
//
// Parameters:
//   - condition SequencePredicate: A predicate function that determines which sequences should be processed by the worker.
//   - worker SeqWorker: The worker to be applied to the sequences that satisfy the condition.
//   - breakOnError bool: If true, the pipeline will stop and return an error if any sequence processing fails. Otherwise, it will log a warning message for each failed sequence.
//
// Returns:
//
//	SeqSliceWorker: A new SeqSliceWorker that processes sequences based on a condition. This function returns a single SeqSliceWorker that can be used to process BioSequences in a workflow or pipeline.
func SeqToSliceConditionalWorker(
	condition SequencePredicate,
	worker SeqWorker, breakOnError bool) SeqSliceWorker {

	if condition == nil {
		return SeqToSliceWorker(worker, breakOnError)
	}

	if worker == nil {
		return nil
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
			} else {
				output[i] = s
				i++
			}
		}

		return output[0:i], nil
	}

	return f
}

// ChainWorkers chains two workers together and returns a new SeqWorker. It takes an existing worker (`worker`) and a next worker as arguments, combines them into a pipeline and applies it to each BioSequence in the sequence slice.
// If `next` is nil, this function just returns the input worker.
// If `worker` is nil, this function just returns the next worker.
//
// Parameters:
//   - worker SeqWorker: The initial worker to be chained. This worker will be executed first on each sequence.
//   - next SeqWorker: The next worker in the pipeline. This worker will be applied to the output of `worker` for each sequence.
//
// Returns:
//
//	SeqWorker: A new SeqWorker that chains the input workers together into a pipeline. This function returns a single SeqWorker that can be used to process BioSequences in a workflow or pipeline.
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
