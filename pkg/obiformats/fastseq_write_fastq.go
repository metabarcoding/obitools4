package obiformats

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

func FormatFastq(seq *obiseq.BioSequence, quality_shift int, formater FormatHeader) string {

	l := seq.Length()
	q := seq.Qualities()
	ascii := make([]byte, seq.Length())

	for j := 0; j < l; j++ {
		ascii[j] = uint8(q[j]) + uint8(quality_shift)
	}

	info := ""
	if formater != nil {
		info = formater(seq)
	}

	return fmt.Sprintf("@%s %s %s\n%s\n+\n%s",
		seq.Id(), info,
		seq.Definition(),
		string(seq.Sequence()),
		string(ascii),
	)
}

func FormatFastqBatch(batch obiiter.BioSequenceBatch, quality_shift int,
	formater FormatHeader) []byte {
	var bs bytes.Buffer
	for _, seq := range batch.Slice() {
		bs.WriteString(FormatFastq(seq, quality_shift, formater))
		bs.WriteString("\n")
	}
	return bs.Bytes()
}

func WriteFastq(iterator obiiter.IBioSequence, file io.Writer, options ...WithOption) error {
	opt := MakeOptions(options)

	header_format := opt.FormatFastSeqHeader()
	quality := opt.QualityShift()

	for iterator.Next() {
		seq := iterator.Get()
		fmt.Fprintln(file, FormatFastq(seq, quality, header_format))
	}

	if opt.CloseFile() {
		switch file := file.(type) {
		case *os.File:
			file.Close()
		}
	}

	return nil
}

func WriteFastqToFile(iterator obiiter.IBioSequence,
	filename string,
	options ...WithOption) error {

	file, err := os.Create(filename)

	if err != nil {
		log.Fatalf("open file error: %v", err)
		return err
	}

	options = append(options, OptionCloseFile())
	return WriteFastq(iterator, file, options...)
}

func WriteFastqToStdout(iterator obiiter.IBioSequence, options ...WithOption) error {
	options = append(options, OptionDontCloseFile())
	return WriteFastq(iterator, os.Stdout, options...)
}

type FileChunck struct {
	text  []byte
	order int
}

func WriteFastqBatch(iterator obiiter.IBioSequenceBatch,
	file io.Writer,
	options ...WithOption) (obiiter.IBioSequenceBatch, error) {
	opt := MakeOptions(options)

	buffsize := iterator.BufferSize()
	newIter := obiiter.MakeIBioSequenceBatch(buffsize)

	nwriters := opt.ParallelWorkers()

	chunkchan := make(chan FileChunck)

	header_format := opt.FormatFastSeqHeader()
	quality := opt.QualityShift()

	newIter.Add(nwriters)

	go func() {
		newIter.WaitAndClose()
		for len(chunkchan) > 0 {
			time.Sleep(time.Millisecond)
		}
		close(chunkchan)
	}()

	ff := func(iterator obiiter.IBioSequenceBatch) {
		for iterator.Next() {
			batch := iterator.Get()
			chunk := FileChunck{
				FormatFastqBatch(batch, quality, header_format),
				batch.Order(),
			}
			chunkchan <- chunk
			newIter.Push(batch)
		}
		newIter.Done()
	}

	log.Println("Start of the fastq file writing")
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

func WriteFastqBatchToStdout(iterator obiiter.IBioSequenceBatch,
	options ...WithOption) (obiiter.IBioSequenceBatch, error) {
	options = append(options, OptionDontCloseFile())
	return WriteFastqBatch(iterator, os.Stdout, options...)
}

func WriteFastqBatchToFile(iterator obiiter.IBioSequenceBatch,
	filename string,
	options ...WithOption) (obiiter.IBioSequenceBatch, error) {

	file, err := os.Create(filename)

	if err != nil {
		log.Fatalf("open file error: %v", err)
		return obiiter.NilIBioSequenceBatch, err
	}

	options = append(options, OptionCloseFile())

	return WriteFastqBatch(iterator, file, options...)
}
