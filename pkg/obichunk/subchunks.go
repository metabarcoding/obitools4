package obichunk

import (
	"sync"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

func ISequenceSubChunk(iterator obiseq.IBioSequenceBatch,
	classifier *obiseq.BioSequenceClassifier,
	sizes ...int) (obiseq.IBioSequenceBatch, error) {

	bufferSize := iterator.BufferSize()
	nworkers := 4

	if len(sizes) > 0 {
		nworkers = sizes[0]
	}

	if len(sizes) > 1 {
		bufferSize = sizes[1]
	}

	newIter := obiseq.MakeIBioSequenceBatch(bufferSize)

	newIter.Add(nworkers)

	go func() {
		newIter.Wait()
		close(newIter.Channel())
	}()

	omutex := sync.Mutex{}
	order := 0

	nextOrder := func() int {
		omutex.Lock()
		neworder := order
		order++
		omutex.Unlock()
		return neworder
	}

	ff := func(iterator obiseq.IBioSequenceBatch) {
		chunks := make(map[int]*obiseq.BioSequenceSlice, 100)

		for iterator.Next() {

			batch := iterator.Get()

			for _, s := range batch.Slice() {
				key := classifier.Code(s)

				slice, ok := chunks[key]

				if !ok {
					slice = obiseq.GetBioSequenceSlicePtr()
					chunks[key] = slice
				}

				*slice = append(*slice, s)
			}

			for k, chunck := range chunks {
				newIter.Channel() <- obiseq.MakeBioSequenceBatch(nextOrder(), *chunck...)
				delete(chunks, k)
			}

			batch.Recycle()
		}

		newIter.Done()
	}

	for i := 0; i < nworkers-1; i++ {
		go ff(iterator.Split())
	}
	go ff(iterator)

	return newIter, nil
}
