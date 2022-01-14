package obiseq

import (
	"log"
	"time"
)

type SeqAnnotator func(BioSequence)

type SeqWorker func(BioSequence) BioSequence
type SeqSliceWorker func(BioSequenceSlice) BioSequenceSlice

func AnnotatorToSeqWorker(function SeqAnnotator) SeqWorker {
	f := func(seq BioSequence) BioSequence {
		function(seq)
		return seq
	}
	return f
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
		newIter.Wait()
		for len(newIter.Channel()) > 0 {
			time.Sleep(time.Millisecond)
		}
		close(newIter.pointer.channel)
		log.Println("End of the batch workers")

	}()

	f := func(iterator IBioSequenceBatch) {
		for iterator.Next() {
			batch := iterator.Get()
			for i, seq := range batch.slice {
				batch.slice[i] = worker(seq)
			}
			newIter.pointer.channel <- batch
		}
		newIter.Done()
	}

	log.Println("Start of the batch workers")
	for i := 0; i < nworkers; i++ {
		go f(iterator.Split())
	}

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
		newIter.Wait()
		for len(newIter.Channel()) > 0 {
			time.Sleep(time.Millisecond)
		}
		close(newIter.pointer.channel)
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
	for i := 0; i < nworkers; i++ {
		go f(iterator.Split())
	}

	return newIter
}
