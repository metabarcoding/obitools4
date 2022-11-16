package obiformats

import (
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
)

func WriteSequenceBatch(iterator obiiter.IBioSequenceBatch,
	file io.Writer,
	options ...WithOption) (obiiter.IBioSequenceBatch, error) {

	iterator = iterator.Rebatch(1000)

	ok := iterator.Next()

	if ok {
		batch := iterator.Get()
		iterator.PushBack()

		var newIter obiiter.IBioSequenceBatch
		var err error

		if len(batch.Slice()) > 0 {
			if batch.Slice()[0].HasQualities() {
				newIter, err = WriteFastqBatch(iterator, file, options...)
			} else {
				newIter, err = WriteFastaBatch(iterator, file, options...)
			}
		} else {
			newIter, err = WriteFastaBatch(iterator, file, options...)
		}

		return newIter, err
	}

	if iterator.Finished() {
		return iterator, nil
	}

	return obiiter.NilIBioSequenceBatch, fmt.Errorf("input iterator not ready")
}

func WriteSequencesBatchToStdout(iterator obiiter.IBioSequenceBatch,
	options ...WithOption) (obiiter.IBioSequenceBatch, error) {
	options = append(options, OptionDontCloseFile())
	return WriteSequenceBatch(iterator, os.Stdout, options...)
}

func WriteSequencesBatchToFile(iterator obiiter.IBioSequenceBatch,
	filename string,
	options ...WithOption) (obiiter.IBioSequenceBatch, error) {

	file, err := os.Create(filename)

	if err != nil {
		log.Fatalf("open file error: %v", err)
		return obiiter.NilIBioSequenceBatch, err
	}

	options = append(options, OptionCloseFile())
	return WriteSequenceBatch(iterator, file, options...)
}
