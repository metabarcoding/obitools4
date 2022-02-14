package obiformats

import (
	"fmt"
	"io"
	"log"
	"os"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

func WriteSequences(iterator obiseq.IBioSequence,
	file io.Writer,
	options ...WithOption) error {

	opts := MakeOptions(options)

	header_format := opts.FormatFastSeqHeader()
	quality := opts.QualityShift()

	ok := iterator.Next()

	if ok {
		seq := iterator.Get()
		if seq.HasQualities() {
			fmt.Fprintln(file, FormatFastq(seq, quality, header_format))
			WriteFastq(iterator, file, options...)
		} else {
			fmt.Fprintln(file, FormatFasta(seq, header_format))
			WriteFasta(iterator, file, options...)
		}
	}

	return nil
}

func WriteSequencesToFile(iterator obiseq.IBioSequence,
	filename string,
	options ...WithOption) error {

	file, err := os.Create(filename)

	if err != nil {
		log.Fatalf("open file error: %v", err)
		return err
	}

	return WriteSequences(iterator, file, options...)
}

func WriteSequencesToStdout(iterator obiseq.IBioSequence, options ...WithOption) error {
	return WriteSequences(iterator, os.Stdout, options...)
}

func WriteSequenceBatch(iterator obiseq.IBioSequenceBatch,
	file io.Writer,
	options ...WithOption) (obiseq.IBioSequenceBatch, error) {

	iterator = iterator.Rebatch(1000)

	ok := iterator.Next()

	if ok {
		batch := iterator.Get()
		iterator.PushBack()

		var newIter obiseq.IBioSequenceBatch
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

	return obiseq.NilIBioSequenceBatch, fmt.Errorf("input iterator not ready")
}

func WriteSequencesBatchToStdout(iterator obiseq.IBioSequenceBatch,
	options ...WithOption) (obiseq.IBioSequenceBatch, error) {
	options = append(options, OptionDontCloseFile())
	return WriteSequenceBatch(iterator, os.Stdout, options...)
}

func WriteSequencesBatchToFile(iterator obiseq.IBioSequenceBatch,
	filename string,
	options ...WithOption) (obiseq.IBioSequenceBatch, error) {

	file, err := os.Create(filename)

	if err != nil {
		log.Fatalf("open file error: %v", err)
		return obiseq.NilIBioSequenceBatch, err
	}

	options = append(options, OptionCloseFile())
	return WriteSequenceBatch(iterator, file, options...)
}
