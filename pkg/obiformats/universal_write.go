package obiformats

import (
	"fmt"
	"io"
	"log"
	"os"

	"git.metabarcoding.org/lecasofts/go/oa2/pkg/obiseq"
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

// func WriteSequenceBatch(iterator obiseq.IBioSequenceBatch,
// 	file io.Writer,
// 	options ...WithOption) error {

// 	opts := MakeOptions(options)

// 	header_format := opts.FormatFastSeqHeader()
// 	quality := opts.QualityShift()

// 	ok := iterator.Next()

// 	if ok {
// 		batch := iterator.Get()
// 		if batch.Slice()[0].HasQualities() {
// 			file.Write()
// 			fmt.Fprintln(file, FormatFastq(seq, quality, header_format))
// 			WriteFastq(iterator, file, options...)
// 		} else {
// 			fmt.Fprintln(file, FormatFasta(seq, header_format))
// 			WriteFasta(iterator, file, options...)
// 		}
// 	}

// 	return nil
// }
