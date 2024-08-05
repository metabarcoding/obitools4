// Package obiformats provides functions for formatting and writing biosequences in various formats.
package obiformats

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

// min returns the minimum of two integers.
//
// Parameters:
// - x: an integer
// - y: an integer
//
// Return:
// - the minimum of x and y (an integer)
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// FormatFasta formats a BioSequence into a FASTA formatted string.
//
// seq is a pointer to the BioSequence to be formatted.
// formater is the FormatHeader function to be used for formatting the sequence header.
// It returns a string containing the formatted FASTA sequence.
func FormatFasta(seq *obiseq.BioSequence, formater FormatHeader) string {
	var fragments strings.Builder

	if seq == nil {
		log.Panicln("try to format a nil BioSequence")
	}

	s := seq.Sequence()
	l := len(s)

	folded := ""
	if l == 0 {
		log.Println("Writing a BioSequence of length zero")
	} else {
		fragments.Grow(l + int(l/60)*2 + 100)

		for i := 0; i < l; i += 60 {
			to := min(i+60, l)
			fmt.Fprintf(&fragments, "%s\n", string(s[i:to]))
		}

		folded = fragments.String()
		folded = folded[:fragments.Len()-1]
	}

	info := formater(seq)
	return fmt.Sprintf(">%s %s\n%s",
		seq.Id(), info,
		folded)
}

// FormatFastaBatch formats a batch of biosequences in FASTA format.
//
// It takes the following parameters:
// - batch: a BioSequenceBatch representing the batch of sequences to format.
// - formater: a FormatHeader function that formats the header of each sequence.
// - skipEmpty: a boolean indicating whether empty sequences should be skipped or not.
//
// It returns a byte array containing the formatted sequences.
func FormatFastaBatch(batch obiiter.BioSequenceBatch, formater FormatHeader, skipEmpty bool) *bytes.Buffer {
	// Create a buffer to store the formatted sequences
	var bs bytes.Buffer

	lt := 0

	for _, seq := range batch.Slice() {
		lt += seq.Len()
	}

	// Iterate over each sequence in the batch
	log.Debugf("FormatFastaBatch: #%d : %d seqs", batch.Order(), batch.Len())
	first := true
	for _, seq := range batch.Slice() {
		// Check if the sequence is empty
		if seq.Len() > 0 {
			// Format the sequence using the provided formater function
			formattedSeq := FormatFasta(seq, formater)

			if first {
				bs.Grow(lt + (len(formattedSeq)-seq.Len())*batch.Len()*5/4)
				first = false
			}

			// Append the formatted sequence to the buffer
			bs.WriteString(formattedSeq)
			bs.WriteByte('\n')
		} else {
			// Handle empty sequences
			if skipEmpty {
				// Skip empty sequences if skipEmpty is true
				log.Warnf("Sequence %s is empty and skipped in output", seq.Id())
			} else {
				// Terminate the program if skipEmpty is false
				log.Fatalf("Sequence %s is empty", seq.Id())
			}
		}
	}

	// Return the byte array representation of the buffer
	return &bs
}

// WriteFasta writes a given iterator of bio sequences to a file in FASTA format.
//
// The function takes an iterator of bio sequences, a file to write to, and
// optional options. It returns a new iterator of bio sequences and an error.
func WriteFasta(iterator obiiter.IBioSequence,
	file io.WriteCloser,
	options ...WithOption) (obiiter.IBioSequence, error) {
	opt := MakeOptions(options)

	file, _ = obiutils.CompressStream(file, opt.CompressedFile(), opt.CloseFile())

	newIter := obiiter.MakeIBioSequence()

	nwriters := opt.ParallelWorkers()

	chunkchan := WriteSeqFileChunk(file, opt.CloseFile())

	header_format := opt.FormatFastSeqHeader()

	newIter.Add(nwriters)

	go func() {
		newIter.WaitAndClose()
		close(chunkchan)
		log.Debugf("Writing fasta file done")
	}()

	ff := func(iterator obiiter.IBioSequence) {
		for iterator.Next() {

			batch := iterator.Get()

			log.Debugf("Formating fasta chunk %d", batch.Order())

			chunkchan <- SeqFileChunk{
				Source: batch.Source(),
				Raw:    FormatFastaBatch(batch, header_format, opt.SkipEmptySequence()),
				Order:  batch.Order(),
			}

			log.Debugf("Fasta chunk %d formated", batch.Order())

			newIter.Push(batch)
		}
		newIter.Done()
	}

	log.Debugln("Start of the fasta file writing")
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
func WriteFastaToStdout(iterator obiiter.IBioSequence,
	options ...WithOption) (obiiter.IBioSequence, error) {
	options = append(options, OptionDontCloseFile())
	return WriteFasta(iterator, os.Stdout, options...)
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
func WriteFastaToFile(iterator obiiter.IBioSequence,
	filename string,
	options ...WithOption) (obiiter.IBioSequence, error) {

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
		return obiiter.NilIBioSequence, err
	}

	options = append(options, OptionCloseFile())

	iterator, err = WriteFasta(iterator, file, options...)

	if opt.HaveToSavePaired() {
		var revfile *os.File

		revfile, err = os.OpenFile(opt.PairedFileName(), flags, 0660)
		if err != nil {
			log.Fatalf("open file error: %v", err)
			return obiiter.NilIBioSequence, err
		}
		iterator, err = WriteFasta(iterator.PairedWith(), revfile, options...)
	}

	return iterator, err
}
