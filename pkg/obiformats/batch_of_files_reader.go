package obiformats

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

func ReadSequencesBatchFromFiles(filenames []string,
	reader IBatchReader,
	concurrent_readers int,
	options ...WithOption) obiiter.IBioSequence {

	if reader == nil {
		reader = ReadSequencesFromFile
	}

	batchiter := obiiter.MakeIBioSequence()
	nextCounter := obiutils.AtomicCounter()

	batchiter.Add(concurrent_readers)

	go func() {
		batchiter.WaitAndClose()
		log.Println("Finnished to read every files")
	}()

	filenameChan := make(chan string)

	go func() {
		for _, filename := range filenames {
			filenameChan <- filename
		}

		close(filenameChan)
	}()

	for i := 0; i < concurrent_readers; i++ {
		go func() {

			for filename := range filenameChan {
				iter, err := reader(filename, options...)

				if err != nil {
					log.Panicf("Cannot open file %s : %v", filename, err)
				}

				log.Printf("Start reading of file : %s", filename)

				// The batches of a file reach us in any order (parallel parsing
				// of the headers). They are numbered here in arrival order, so
				// the order of the file must be restored first, unless the
				// order of the data is not wanted (several concurrent readers).
				if concurrent_readers == 1 {
					iter = iter.SortBatches()
				}

				for iter.Next() {
					batch := iter.Get()
					batchiter.Push(batch.Reorder(nextCounter()))
				}

				log.Printf("End of reading of file : %s", filename)

			}
			batchiter.Done()
		}()
	}

	return batchiter
}
