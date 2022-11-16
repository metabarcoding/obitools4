package obiiter

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

type SeqAnnotator func(*obiseq.BioSequence)

type SeqWorker func(*obiseq.BioSequence) *obiseq.BioSequence
type SeqSliceWorker func(obiseq.BioSequenceSlice) obiseq.BioSequenceSlice

func AnnotatorToSeqWorker(function SeqAnnotator) SeqWorker {
	f := func(seq *obiseq.BioSequence) *obiseq.BioSequence {
		function(seq)
		return seq
	}
	return f
}

// That method allows for applying a SeqWorker function on every sequences.
//
// Sequences are provided by the iterator and modified sequences are pushed
// on the returned IBioSequenceBatch.
//
// Moreover the SeqWorker function, the method accepted two optional integer parameters.
//   - First is allowing to indicates the number of workers running in parallele (default 4)
//   - The second the size of the chanel buffer. By default set to the same value than the input buffer.
func (iterator IBioSequenceBatch) MakeIWorker(worker SeqWorker, sizes ...int) IBioSequenceBatch {
	nworkers := 4
	buffsize := iterator.BufferSize()

	if len(sizes) > 0 {
		nworkers = sizes[0]
	}

	if len(sizes) > 1 {
		buffsize = sizes[1]
	}

	newIter := MakeIBioSequenceBatch(buffsize)

	newIter.Add(nworkers)

	go func() {
		newIter.WaitAndClose()
		log.Debugln("End of the batch workers")

	}()

	f := func(iterator IBioSequenceBatch) {
		for iterator.Next() {
			batch := iterator.Get()
			for i, seq := range batch.slice {
				batch.slice[i] = worker(seq)
			}
			newIter.Push(batch)
		}
		newIter.Done()
	}

	log.Debugln("Start of the batch workers")
	for i := 0; i < nworkers-1; i++ {
		go f(iterator.Split())
	}
	go f(iterator)

	return newIter
}

func (iterator IBioSequenceBatch) MakeIConditionalWorker(predicate obiseq.SequencePredicate,
	worker SeqWorker, sizes ...int) IBioSequenceBatch {
	nworkers := 4
	buffsize := iterator.BufferSize()

	if len(sizes) > 0 {
		nworkers = sizes[0]
	}

	if len(sizes) > 1 {
		buffsize = sizes[1]
	}

	newIter := MakeIBioSequenceBatch(buffsize)

	newIter.Add(nworkers)

	go func() {
		newIter.WaitAndClose()
		log.Debugln("End of the batch workers")

	}()

	f := func(iterator IBioSequenceBatch) {
		for iterator.Next() {
			batch := iterator.Get()
			for i, seq := range batch.slice {
				if predicate(batch.slice[i]) {
					batch.slice[i] = worker(seq)
				}
			}
			newIter.Push(batch)
		}
		newIter.Done()
	}

	log.Debugln("Start of the batch workers")
	for i := 0; i < nworkers-1; i++ {
		go f(iterator.Split())
	}
	go f(iterator)

	return newIter
}

func (iterator IBioSequenceBatch) MakeISliceWorker(worker SeqSliceWorker, sizes ...int) IBioSequenceBatch {
	nworkers := 4
	buffsize := iterator.BufferSize()

	if len(sizes) > 0 {
		nworkers = sizes[0]
	}

	if len(sizes) > 1 {
		buffsize = sizes[1]
	}

	newIter := MakeIBioSequenceBatch(buffsize)

	newIter.Add(nworkers)

	go func() {
		newIter.WaitAndClose()
		log.Println("End of the batch slice workers")
	}()

	f := func(iterator IBioSequenceBatch) {
		for iterator.Next() {
			batch := iterator.Get()
			batch.slice = worker(batch.slice)
			newIter.pointer.channel <- batch
		}
		newIter.Done()
	}

	log.Printf("Start of the batch slice workers on %d workers (buffer : %d)\n", nworkers, buffsize)
	for i := 0; i < nworkers-1; i++ {
		go f(iterator.Split())
	}
	go f(iterator)

	return newIter
}

func WorkerPipe(worker SeqWorker, sizes ...int) Pipeable {
	f := func(iterator IBioSequenceBatch) IBioSequenceBatch {
		return iterator.MakeIWorker(worker, sizes...)
	}

	return f
}

func SliceWorkerPipe(worker SeqSliceWorker, sizes ...int) Pipeable {
	f := func(iterator IBioSequenceBatch) IBioSequenceBatch {
		return iterator.MakeISliceWorker(worker, sizes...)
	}

	return f
}

func ReverseComplementWorker(inplace bool) SeqWorker {
	f := func(input *obiseq.BioSequence) *obiseq.BioSequence {
		return input.ReverseComplement(inplace)
	}

	return f
}
