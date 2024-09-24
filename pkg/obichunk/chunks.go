package obichunk

import (
	"sync"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

func ISequenceChunk(iterator obiiter.IBioSequence,
	classifier *obiseq.BioSequenceClassifier) (obiiter.IBioSequence, error) {

	newIter := obiiter.MakeIBioSequence()

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
		sources := make(map[int]string, 1000)

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

				source := ""
				for data.Next() {
					b := data.Get()
					source = b.Source()
					*chunk = append(*chunk, b.Slice()...)
				}

				lock.Lock()
				sources[newflux] = source
				lock.Unlock()

				jobDone.Done()
			}(newflux)
		}

		jobDone.Wait()
		order := 0

		for i, chunk := range chunks {

			if len(*chunk) > 0 {
				newIter.Push(obiiter.MakeBioSequenceBatch(sources[i], order, *chunk))
				order++
			}

		}
		newIter.Done()
	}()

	return newIter, nil
}
