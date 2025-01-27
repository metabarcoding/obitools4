package obicsv

import (
	"bytes"
	"encoding/csv"
	"io"
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiformats"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"

	log "github.com/sirupsen/logrus"
)

func FormatCVSBatch(batch obiiter.CSVRecordBatch, header obiiter.CSVHeader, navalue string) *bytes.Buffer {
	buff := new(bytes.Buffer)
	csv := csv.NewWriter(buff)

	if batch.Order() == 0 {
		csv.Write(header)
	}
	for _, s := range batch.Slice() {
		data := make([]string, len(header))
		for i, key := range header {
			var sval string
			val, ok := s[key]

			if !ok {
				sval = navalue
			} else {
				var err error
				sval, err = obiutils.InterfaceToString(val)
				if err != nil {
					sval = navalue
				}
			}
			data[i] = sval
		}
		csv.Write(data)
	}

	csv.Flush()

	return buff
}

func WriteCSV(iterator *obiiter.ICSVRecord,
	file io.WriteCloser,
	options ...WithOption) (*obiiter.ICSVRecord, error) {
	opt := MakeOptions(options)

	file, _ = obiutils.CompressStream(file, opt.CompressedFile(), opt.CloseFile())

	newIter := obiiter.NewICSVRecord()

	nwriters := opt.ParallelWorkers()

	chunkchan := obiformats.WriteFileChunk(file, opt.CloseFile())

	newIter.Add(nwriters)

	go func() {
		newIter.WaitAndClose()
		close(chunkchan)
		log.Debugf("Writing CSV file done")
	}()

	ff := func(iterator *obiiter.ICSVRecord) {
		for iterator.Next() {

			batch := iterator.Get()

			log.Debugf("Formating CSV chunk %d", batch.Order())

			ss := obiformats.FileChunk{
				Source: batch.Source(),
				Raw: FormatCVSBatch(
					batch,
					iterator.Header(),
					opt.CSVNAValue(),
				),
				Order: batch.Order(),
			}

			chunkchan <- ss

			log.Debugf("CSV chunk %d formated", batch.Order())

			newIter.Push(batch)
		}
		newIter.Done()
	}

	log.Debugln("Start of the CSV file writing")
	go ff(iterator)
	for i := 0; i < nwriters-1; i++ {
		go ff(iterator.Split())
	}

	return newIter, nil
}

// WriteFastaToStdout writes the given bio sequence iterator to standard output in FASTA format.
//
// The function takes an iterator of bio sequences as the first parameter and optional
// configuration options as variadic arguments. It appends the option to not close the file
// to the options slice and then calls the WriteFasta function passing the iterator,
// os.Stdout as the output file, and the options slice.
//
// The function returns the same bio sequence iterator and an error if any occurred.
func WriteCSVToStdout(iterator *obiiter.ICSVRecord,
	options ...WithOption) (*obiiter.ICSVRecord, error) {
	// options = append(options, OptionDontCloseFile())
	options = append(options, OptionCloseFile())
	return WriteCSV(iterator, os.Stdout, options...)
}

// WriteFastaToFile writes the given iterator of biosequences to a file with the specified filename,
// using the provided options. It returns the updated iterator and any error that occurred.
//
// Parameters:
// - iterator: The biosequence iterator to write to the file.
// - filename: The name of the file to write to.
// - options: Zero or more optional parameters to customize the writing process.
//
// Returns:
// - obiiter.IBioSequence: The updated biosequence iterator.
// - error: Any error that occurred during the writing process.
func WriteCSVToFile(iterator *obiiter.ICSVRecord,
	filename string,
	options ...WithOption) (*obiiter.ICSVRecord, error) {

	opt := MakeOptions(options)
	flags := os.O_WRONLY | os.O_CREATE

	if opt.AppendFile() {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	file, err := os.OpenFile(filename, flags, 0660)

	if err != nil {
		log.Fatalf("open file error: %v", err)
		return nil, err
	}

	options = append(options, OptionCloseFile())

	iterator, err = WriteCSV(iterator, file, options...)

	return iterator, err
}
