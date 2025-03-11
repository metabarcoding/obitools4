package obiiter

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// That method allows for applying a SeqWorker function on every sequences.
//
// Sequences are provided by the iterator and modified sequences are pushed
// on the returned IBioSequenceBatch.
//
// Moreover the SeqWorker function, the method accepted two optional integer parameters.
//   - First is allowing to indicates the number of workers running in parallele (default 4)
//   - The second the size of the chanel buffer. By default set to the same value than the input buffer.
func (iterator IBioSequence) MakeIWorker(worker obiseq.SeqWorker,
	breakOnError bool,
	sizes ...int) IBioSequence {
	nworkers := obidefault.ParallelWorkers()

	if len(sizes) > 0 {
		nworkers = sizes[0]
	}

	sw := obiseq.SeqToSliceWorker(worker, breakOnError)
	return iterator.MakeISliceWorker(sw, breakOnError, nworkers)
}

// MakeIConditionalWorker applies a given worker function to each sequence in the iterator that satisfies the given predicate.
// It creates a new iterator with the modified sequences and returns it.
//
// Parameters:
// - predicate: A function that takes a sequence and returns a boolean value indicating whether the sequence satisfies a certain condition.
// - worker: A function that takes a sequence and returns a modified version of the sequence.
// - sizes: Optional. One or more integers representing the number of workers to be used for parallel processing. If not provided, the number of workers will be determined by the obidefault.ReadParallelWorkers() function.
//
// Return:
// - newIter: A new IBioSequence iterator with the modified sequences.
func (iterator IBioSequence) MakeIConditionalWorker(predicate obiseq.SequencePredicate,
	worker obiseq.SeqWorker, breakOnError bool, sizes ...int) IBioSequence {
	nworkers := obidefault.ReadParallelWorkers()

	if len(sizes) > 0 {
		nworkers = sizes[0]
	}

	sw := obiseq.SeqToSliceConditionalWorker(predicate, worker, breakOnError)

	return iterator.MakeISliceWorker(sw, breakOnError, nworkers)

}

// MakeISliceWorker applies a SeqSliceWorker function to each slice in the IBioSequence iterator,
// creating a new IBioSequence with the modified slices.
//
// The worker function takes a slice as input and returns a modified slice. It is applied to each
// slice in the iterator.
//
// The sizes argument is optional and specifies the number of workers to use. If sizes is not
// provided, the default number of workers is used.
//
// The function returns a new IBioSequence containing the modified slices.
func (iterator IBioSequence) MakeISliceWorker(worker obiseq.SeqSliceWorker, breakOnError bool, sizes ...int) IBioSequence {

	if worker == nil {
		return iterator
	}

	nworkers := obidefault.ParallelWorkers()

	if len(sizes) > 0 {
		nworkers = sizes[0]
	}

	newIter := MakeIBioSequence()

	f := func(iterator IBioSequence) {
		var err error
		for iterator.Next() {
			batch := iterator.Get()
			batch.slice, err = worker(batch.slice)
			if err != nil && breakOnError {
				log.Fatalf("Error on sequence processing : %v", err)
			}
			newIter.Push(batch)
		}
		newIter.Done()
	}

	log.Debugln("Start of the batch workers")
	for i := 1; i < nworkers; i++ {
		newIter.Add(1)
		go f(iterator.Split())
	}
	newIter.Add(1)
	go f(iterator)

	go func() {
		newIter.WaitAndClose()
		log.Debugln("End of the batch workers")

	}()

	if iterator.IsPaired() {
		newIter.MarkAsPaired()
	}

	return newIter
}

// WorkerPipe is a function that takes a SeqWorker and a variadic list of sizes as parameters and returns a Pipeable.
//
// The WorkerPipe function creates a closure that takes an IBioSequence iterator as a parameter and returns an IBioSequence.
// Inside the closure, the MakeIWorker method of the iterator is called with the provided worker and sizes, and the result is returned.
//
// Parameters:
// - worker: A SeqWorker object that represents the worker to be used in the closure.
// - sizes: A variadic list of int values that represents the sizes to be used in the MakeIWorker method.
//
// Return:
// - f: A Pipeable object that represents the closure created by the WorkerPipe function.
func WorkerPipe(worker obiseq.SeqWorker, breakOnError bool, sizes ...int) Pipeable {
	f := func(iterator IBioSequence) IBioSequence {
		return iterator.MakeIWorker(worker, breakOnError, sizes...)
	}

	return f
}

// SliceWorkerPipe creates a Pipeable function that applies a SeqSliceWorker to an iterator.
//
// The worker parameter is the SeqSliceWorker to be applied.
// The sizes parameter is a variadic parameter representing the sizes of the slices.
// The function returns a Pipeable function that applies the SeqSliceWorker to the iterator.
func SliceWorkerPipe(worker obiseq.SeqSliceWorker, breakOnError bool, sizes ...int) Pipeable {
	f := func(iterator IBioSequence) IBioSequence {
		return iterator.MakeISliceWorker(worker, breakOnError, sizes...)
	}

	return f
}
