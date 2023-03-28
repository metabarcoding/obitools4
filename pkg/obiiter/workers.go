package obiiter

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

// That method allows for applying a SeqWorker function on every sequences.
//
// Sequences are provided by the iterator and modified sequences are pushed
// on the returned IBioSequenceBatch.
//
// Moreover the SeqWorker function, the method accepted two optional integer parameters.
//   - First is allowing to indicates the number of workers running in parallele (default 4)
//   - The second the size of the chanel buffer. By default set to the same value than the input buffer.
func (iterator IBioSequence) MakeIWorker(worker obiseq.SeqWorker, sizes ...int) IBioSequence {
	nworkers := 4

	if len(sizes) > 0 {
		nworkers = sizes[0]
	}

	newIter := MakeIBioSequence()

	newIter.Add(nworkers)

	go func() {
		newIter.WaitAndClose()
		log.Debugln("End of the batch workers")

	}()

	f := func(iterator IBioSequence) {
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

	if iterator.IsPaired() {
		newIter.MarkAsPaired()
	}

	return newIter
}

func (iterator IBioSequence) MakeIConditionalWorker(predicate obiseq.SequencePredicate,
	worker obiseq.SeqWorker, sizes ...int) IBioSequence {
	nworkers := 4

	if len(sizes) > 0 {
		nworkers = sizes[0]
	}

	newIter := MakeIBioSequence()

	newIter.Add(nworkers)

	go func() {
		newIter.WaitAndClose()
		log.Debugln("End of the batch workers")

	}()

	f := func(iterator IBioSequence) {
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

	if iterator.IsPaired() {
		newIter.MarkAsPaired()
	}

	return newIter
}

func (iterator IBioSequence) MakeISliceWorker(worker obiseq.SeqSliceWorker, sizes ...int) IBioSequence {
	nworkers := 4

	if len(sizes) > 0 {
		nworkers = sizes[0]
	}

	newIter := MakeIBioSequence()

	newIter.Add(nworkers)

	go func() {
		newIter.WaitAndClose()
		log.Println("End of the batch slice workers")
	}()

	f := func(iterator IBioSequence) {
		for iterator.Next() {
			batch := iterator.Get()
			batch.slice = worker(batch.slice)
			newIter.Push(batch)
		}
		newIter.Done()
	}

	log.Printf("Start of the batch slice workers on %d workers\n", nworkers)
	for i := 0; i < nworkers-1; i++ {
		go f(iterator.Split())
	}
	go f(iterator)

	if iterator.IsPaired() {
		newIter.MarkAsPaired()
	}

	return newIter
}

func WorkerPipe(worker obiseq.SeqWorker, sizes ...int) Pipeable {
	f := func(iterator IBioSequence) IBioSequence {
		return iterator.MakeIWorker(worker, sizes...)
	}

	return f
}

func SliceWorkerPipe(worker obiseq.SeqSliceWorker, sizes ...int) Pipeable {
	f := func(iterator IBioSequence) IBioSequence {
		return iterator.MakeISliceWorker(worker, sizes...)
	}

	return f
}
