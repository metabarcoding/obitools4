package obiiter

import (
	"log"

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
		log.Println("End of the batch workers")

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

	log.Println("Start of the batch workers")
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

	log.Println("Start of the batch slice workers")
	for i := 0; i < nworkers-1; i++ {
		go f(iterator.Split())
	}
	go f(iterator)

	return newIter
}


func (iterator IBioSequence) MakeIWorker(worker SeqWorker, sizes ...int) IBioSequence {
	buffsize := iterator.BufferSize()

	if len(sizes) > 0 {
		buffsize = sizes[0]
	}

	newIter := MakeIBioSequence(buffsize)

	newIter.Add(1)

	go func() {
		newIter.Wait()
		close(newIter.pointer.channel)
	}()

	go func() {
		for iterator.Next() {
			seq := iterator.Get()
			seq = worker(seq)
			newIter.pointer.channel <- seq
		}
		newIter.Done()
	}()

	return newIter
}

func WorkerPipe(worker SeqWorker, sizes ...int) Pipeable {
	f := func(iterator IBioSequenceBatch) IBioSequenceBatch {
		return iterator.MakeIWorker(worker,sizes...)
	}

	return f
}

func SliceWorkerPipe(worker SeqSliceWorker, sizes ...int) Pipeable {
	f := func(iterator IBioSequenceBatch) IBioSequenceBatch {
		return iterator.MakeISliceWorker(worker,sizes...)
	}

	return f
}
