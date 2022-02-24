package obichunk

import (
	log "github.com/sirupsen/logrus"
	"sync"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

func ISequenceChunk(iterator obiiter.IBioSequenceBatch,
	classifier *obiseq.BioSequenceClassifier,
	sizes ...int) (obiiter.IBioSequenceBatch, error) {

	bufferSize := iterator.BufferSize()

	if len(sizes) > 0 {
		bufferSize = sizes[0]
	}

	newIter := obiiter.MakeIBioSequenceBatch(bufferSize)

	newIter.Add(1)

	go func() {
		newIter.Wait()
		newIter.Close()
	}()

	go func() {
		lock := sync.Mutex{}

		dispatcher := iterator.Distribute(classifier)

		jobDone := sync.WaitGroup{}
		chunks := make(map[int]*obiseq.BioSequenceSlice, 1000)

		for newflux := range dispatcher.News() {
			jobDone.Add(1)
			go func(newflux int) {
				data, err := dispatcher.Outputs(newflux)

				if err != nil {
					log.Fatalf("Cannot retreive the new chanel : %v", err)
				}

				chunk := obiseq.NewBioSequenceSlice()
				lock.Lock()
				chunks[newflux] = chunk
				lock.Unlock()

				for data.Next() {
					b := data.Get()
					*chunk = append(*chunk, b.Slice()...)
					b.Recycle()
				}

				jobDone.Done()
			}(newflux)
		}

		jobDone.Wait()
		order := 0

		for _, chunck := range chunks {

			if len(*chunck) > 0 {
				newIter.Push(obiiter.MakeBioSequenceBatch(order, *chunck))
				order++
			}

		}
		newIter.Done()
	}()

	return newIter, nil
}
