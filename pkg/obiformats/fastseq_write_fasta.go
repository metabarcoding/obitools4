package obiformats

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func FormatFasta(seq *obiseq.BioSequence, formater FormatHeader) string {
	var fragments strings.Builder

	if seq==nil {
		log.Panicln("try to format a nil BioSequence")
	}

	s := seq.Sequence()
	l := len(s)

	fragments.Grow(l + int(l/60) + 10)

	for i := 0; i < l; i += 60 {
		to := min(i+60, l)
		fmt.Fprintf(&fragments, "%s\n", string(s[i:to]))
	}

	folded := fragments.String()
	folded = folded[:fragments.Len()-1]
	info := formater(seq)
	return fmt.Sprintf(">%s %s %s\n%s",
		seq.Id(), info,
		seq.Definition(),
		folded)
}

func FormatFastaBatch(batch obiseq.BioSequenceBatch, formater FormatHeader) []byte {
	var bs bytes.Buffer
	for _, seq := range batch.Slice() {
		bs.WriteString(FormatFasta(seq, formater))
		bs.WriteString("\n")
	}
	return bs.Bytes()
}

func WriteFasta(iterator obiseq.IBioSequence, file io.Writer, options ...WithOption) error {
	opt := MakeOptions(options)

	header_format := opt.FormatFastSeqHeader()

	for iterator.Next() {
		seq := iterator.Get()
		fmt.Fprintln(file, FormatFasta(seq, header_format))
	}

	if opt.CloseFile() {
		switch file := file.(type) {
		case *os.File:
			file.Close()
		}
	}

	return nil
}

func WriteFastaToFile(iterator obiseq.IBioSequence,
	filename string,
	options ...WithOption) error {

	file, err := os.Create(filename)

	if err != nil {
		log.Fatalf("open file error: %v", err)
		return err
	}

	options = append(options, OptionCloseFile())

	return WriteFasta(iterator, file, options...)
}

func WriteFastaToStdout(iterator obiseq.IBioSequence, options ...WithOption) error {
	options = append(options, OptionDontCloseFile())
	return WriteFasta(iterator, os.Stdout, options...)
}

func WriteFastaBatch(iterator obiseq.IBioSequenceBatch, file io.Writer, options ...WithOption) (obiseq.IBioSequenceBatch, error) {
	opt := MakeOptions(options)

	buffsize := iterator.BufferSize()
	newIter := obiseq.MakeIBioSequenceBatch(buffsize)

	nwriters := opt.ParallelWorkers()

	chunkchan := make(chan FileChunck)

	header_format := opt.FormatFastSeqHeader()

	newIter.Add(nwriters)

	go func() {
		newIter.WaitAndClose()
		close(chunkchan)
	}()

	ff := func(iterator obiseq.IBioSequenceBatch) {
		for iterator.Next() {
			batch := iterator.Get()
			chunkchan <- FileChunck{
				FormatFastaBatch(batch, header_format),
				batch.Order(),
			}
			newIter.Push(batch)
		}
		newIter.Done()
	}

	log.Println("Start of the fasta file writing")
	go ff(iterator)
	for i := 0; i < nwriters-1; i++ {
		go ff(iterator.Split())
	}

	next_to_send := 0
	received := make(map[int]FileChunck, 100)

	go func() {
		for chunk := range chunkchan {
			if chunk.order == next_to_send {
				file.Write(chunk.text)
				next_to_send++
				chunk, ok := received[next_to_send]
				for ok {
					file.Write(chunk.text)
					delete(received, next_to_send)
					next_to_send++
					chunk, ok = received[next_to_send]
				}
			} else {
				received[chunk.order] = chunk
			}

		}

		if opt.CloseFile() {
			switch file := file.(type) {
			case *os.File:
				file.Close()
			}
		}

	}()

	return newIter, nil
}

func WriteFastaBatchToStdout(iterator obiseq.IBioSequenceBatch, options ...WithOption) (obiseq.IBioSequenceBatch, error) {
	options = append(options, OptionDontCloseFile())
	return WriteFastaBatch(iterator, os.Stdout, options...)
}

func WriteFastaBatchToFile(iterator obiseq.IBioSequenceBatch,
	filename string,
	options ...WithOption) (obiseq.IBioSequenceBatch, error) {

	file, err := os.Create(filename)

	if err != nil {
		log.Fatalf("open file error: %v", err)
		return obiseq.NilIBioSequenceBatch, err
	}

	options = append(options, OptionCloseFile())

	return WriteFastaBatch(iterator, file, options...)
}
