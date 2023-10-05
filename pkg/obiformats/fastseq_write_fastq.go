package obiformats

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiutils"
)

// The function FormatFastq takes a BioSequence object, a quality shift value, and a header formatter
// function as input, and returns a formatted string in FASTQ format.
func FormatFastq(seq *obiseq.BioSequence, quality_shift int, formater FormatHeader) string {

	l := seq.Len()
	q := seq.Qualities()
	ascii := make([]byte, seq.Len())

	for j := 0; j < l; j++ {
		ascii[j] = uint8(q[j]) + uint8(quality_shift)
	}

	info := ""
	if formater != nil {
		info = formater(seq)
	}

	return fmt.Sprintf("@%s %s\n%s\n+\n%s",
		seq.Id(), info,
		string(seq.Sequence()),
		string(ascii),
	)
}

func FormatFastqBatch(batch obiiter.BioSequenceBatch, quality_shift int,
	formater FormatHeader, skipEmpty bool) []byte {
	var bs bytes.Buffer
	for _, seq := range batch.Slice() {
		if seq.Len() > 0 {
			bs.WriteString(FormatFastq(seq, quality_shift, formater))
			bs.WriteString("\n")
		} else {
			if skipEmpty {
				log.Warnf("Sequence %s is empty and skiped in output", seq.Id())
			} else {
				log.Fatalf("Sequence %s is empty", seq.Id())
			}
		}

	}
	return bs.Bytes()
}

type FileChunck struct {
	text  []byte
	order int
}

func WriteFastq(iterator obiiter.IBioSequence,
	file io.WriteCloser,
	options ...WithOption) (obiiter.IBioSequence, error) {

	iterator = iterator.Rebatch(1000)

	opt := MakeOptions(options)

	file, _ = obiutils.CompressStream(file, opt.CompressedFile(), opt.CloseFile())

	newIter := obiiter.MakeIBioSequence()

	nwriters := opt.ParallelWorkers()

	obiiter.RegisterAPipe()
	chunkchan := make(chan FileChunck)

	header_format := opt.FormatFastSeqHeader()
	quality := opt.QualityShift()

	newIter.Add(nwriters)

	var waitWriter sync.WaitGroup

	go func() {
		newIter.WaitAndClose()
		for len(chunkchan) > 0 {
			time.Sleep(time.Millisecond)
		}
		close(chunkchan)
		waitWriter.Wait()
	}()

	ff := func(iterator obiiter.IBioSequence) {
		for iterator.Next() {
			batch := iterator.Get()
			chunk := FileChunck{
				FormatFastqBatch(batch, quality, header_format, opt.SkipEmptySequence()),
				batch.Order(),
			}
			chunkchan <- chunk
			newIter.Push(batch)
		}
		newIter.Done()
	}

	log.Debugln("Start of the fastq file writing")
	go ff(iterator)
	for i := 0; i < nwriters-1; i++ {
		go ff(iterator.Split())
	}

	next_to_send := 0
	received := make(map[int]FileChunck, 100)

	waitWriter.Add(1)
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

		file.Close()

		log.Debugln("End of the fastq file writing")
		obiiter.UnregisterPipe()
		waitWriter.Done()
	}()

	return newIter, nil
}

func WriteFastqToStdout(iterator obiiter.IBioSequence,
	options ...WithOption) (obiiter.IBioSequence, error) {
	options = append(options, OptionDontCloseFile())
	return WriteFastq(iterator, os.Stdout, options...)
}

func WriteFastqToFile(iterator obiiter.IBioSequence,
	filename string,
	options ...WithOption) (obiiter.IBioSequence, error) {

	opt := MakeOptions(options)
	flags := os.O_WRONLY | os.O_CREATE

	if opt.AppendFile() {
		flags |= os.O_APPEND
	}
	file, err := os.OpenFile(filename, flags, 0660)

	if err != nil {
		log.Fatalf("open file error: %v", err)
		return obiiter.NilIBioSequence, err
	}

	options = append(options, OptionCloseFile())

	iterator, err = WriteFastq(iterator, file, options...)

	if opt.HaveToSavePaired() {
		var revfile *os.File

		revfile, err = os.OpenFile(opt.PairedFileName(), flags, 0660)
		if err != nil {
			log.Fatalf("open file error: %v", err)
			return obiiter.NilIBioSequence, err
		}
		iterator, err = WriteFastq(iterator.PairedWith(), revfile, options...)
	}

	return iterator, err
}
