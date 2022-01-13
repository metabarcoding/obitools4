package obiformats

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

func FormatFastq(seq obiseq.BioSequence, quality_shift int, formater FormatHeader) string {

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

func FormatFastqBatch(batch obiseq.BioSequenceBatch, quality_shift int,
	formater FormatHeader) []byte {
	var bs bytes.Buffer
	for _, seq := range batch.Slice() {
		bs.WriteString(FormatFastq(seq, quality_shift, formater))
		bs.WriteString("\n")
	}
	return bs.Bytes()
}

func WriteFastq(iterator obiseq.IBioSequence, file io.Writer, options ...WithOption) error {
	opt := MakeOptions(options)

	header_format := opt.FormatFastSeqHeader()
	quality := opt.QualityShift()

	for iterator.Next() {
		seq := iterator.Get()
		fmt.Fprintln(file, FormatFastq(seq, quality, header_format))
	}

	return nil
}

func WriteFastqToFile(iterator obiseq.IBioSequence,
	filename string,
	options ...WithOption) error {

	file, err := os.Create(filename)

	if err != nil {
		log.Fatalf("open file error: %v", err)
		return err
	}

	return WriteFastq(iterator, file, options...)
}

func WriteFastqToStdout(iterator obiseq.IBioSequence, options ...WithOption) error {
	return WriteFastq(iterator, os.Stdout, options...)
}

type FileChunck struct {
	text  []byte
	order int
}

func WriteFastqBatch(iterator obiseq.IBioSequenceBatch, file io.Writer, options ...WithOption) (obiseq.IBioSequenceBatch, error) {
	buffsize := iterator.BufferSize()
	new_iter := obiseq.MakeIBioSequenceBatch(buffsize)

	opt := MakeOptions(options)
	nwriters := 4

	chunkchan := make(chan FileChunck)

	header_format := opt.FormatFastSeqHeader()
	quality := opt.QualityShift()

	new_iter.Add(nwriters)

	go func() {
		new_iter.Wait()
		for len(chunkchan) > 0 {
			time.Sleep(time.Millisecond)
		}
		close(chunkchan)
		for len(new_iter.Channel()) > 0 {
			time.Sleep(time.Millisecond)
		}
		close(new_iter.Channel())
	}()

	ff := func(iterator obiseq.IBioSequenceBatch) {
		for iterator.Next() {
			batch := iterator.Get()
			chunkchan <- FileChunck{
				FormatFastqBatch(batch, quality, header_format),
				batch.Order(),
			}
			new_iter.Channel() <- batch
		}
		new_iter.Done()
	}

	log.Println("Start of the fastq file reading")
	for i := 0; i < nwriters; i++ {
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
	}()

	return new_iter, nil
}

func WriteFastqBatchToStdout(iterator obiseq.IBioSequenceBatch, options ...WithOption) (obiseq.IBioSequenceBatch, error) {
	return WriteFastqBatch(iterator, os.Stdout, options...)
}

func WriteFastqBatchToFile(iterator obiseq.IBioSequenceBatch,
	filename string,
	options ...WithOption) (obiseq.IBioSequenceBatch, error) {

	file, err := os.Create(filename)

	if err != nil {
		log.Fatalf("open file error: %v", err)
		return obiseq.NilIBioSequenceBatch, err
	}

	return WriteFastqBatch(iterator, file, options...)
}
