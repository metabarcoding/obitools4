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

	new_iter := MakeIBioSequence(buffsize)

	new_iter.Add(1)

	go func() {
		new_iter.Wait()
		close(new_iter.pointer.channel)
	}()

	go func() {
		for iterator.Next() {
			seq := iterator.Get()
			seq = worker(seq)
			new_iter.pointer.channel <- seq
		}
		new_iter.Done()
	}()

	return new_iter
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

	new_iter := MakeIBioSequenceBatch(buffsize)

	new_iter.Add(nworkers)

	go func() {
		new_iter.Wait()
		for len(new_iter.Channel()) > 0 {
			time.Sleep(time.Millisecond)
		}
		close(new_iter.pointer.channel)
		log.Println("End of the batch workers")

	}()

	f := func(iterator IBioSequenceBatch) {
		for iterator.Next() {
			batch := iterator.Get()
			for i, seq := range batch.slice {
				batch.slice[i] = worker(seq)
			}
			new_iter.pointer.channel <- batch
		}
		new_iter.Done()
	}

	log.Println("Start of the batch workers")
	for i := 0; i < nworkers; i++ {
		go f(iterator.Split())
	}

	return new_iter
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

	new_iter := MakeIBioSequenceBatch(buffsize)

	new_iter.Add(nworkers)

	go func() {
		new_iter.Wait()
		for len(new_iter.Channel()) > 0 {
			time.Sleep(time.Millisecond)
		}
		close(new_iter.pointer.channel)
		log.Println("End of the batch slice workers")
	}()

	f := func(iterator IBioSequenceBatch) {
		for iterator.Next() {
			batch := iterator.Get()
			batch.slice = worker(batch.slice)
			new_iter.pointer.channel <- batch
		}
		new_iter.Done()
	}

	log.Println("Start of the batch slice workers")
	for i := 0; i < nworkers; i++ {
		go f(iterator.Split())
	}

	return new_iter
}
