package obiformats

import (
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
)

func WriteSequence(iterator obiiter.IBioSequence,
	file io.Writer,
	options ...WithOption) (obiiter.IBioSequence, error) {

	iterator = iterator.Rebatch(1000)

	ok := iterator.Next()

	if ok {
		batch := iterator.Get()
		iterator.PushBack()

		var newIter obiiter.IBioSequence
		var err error

		if len(batch.Slice()) > 0 {
			if batch.Slice()[0].HasQualities() {
				newIter, err = WriteFastq(iterator, file, options...)
			} else {
				newIter, err = WriteFasta(iterator, file, options...)
			}
		} else {
			newIter, err = WriteFasta(iterator, file, options...)
		}

		return newIter, err
	}

	if iterator.Finished() {
		return iterator, nil
	}

	return obiiter.NilIBioSequence, fmt.Errorf("input iterator not ready")
}

func WriteSequencesToStdout(iterator obiiter.IBioSequence,
	options ...WithOption) (obiiter.IBioSequence, error) {
	options = append(options, OptionDontCloseFile())
	return WriteSequence(iterator, os.Stdout, options...)
}

func WriteSequencesToFile(iterator obiiter.IBioSequence,
	filename string,
	options ...WithOption) (obiiter.IBioSequence, error) {

	var file *os.File
	var err error

	opt := MakeOptions(options)

	if opt.AppendFile() {
		log.Debug("Open files in appending mode")
		file, err = os.OpenFile(filename,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	} else {
		file, err = os.Create(filename)
	}

	if err != nil {
		log.Fatalf("open file error: %v", err)
		return obiiter.NilIBioSequence, err
	}

	options = append(options, OptionCloseFile())
	return WriteSequence(iterator, file, options...)
}
