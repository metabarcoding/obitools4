package obiformats

import (
	"bytes"
	"io"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

func _formatFastq(buff *bytes.Buffer, seq *obiseq.BioSequence, formater FormatHeader) {

	info := ""
	if formater != nil {
		info = formater(seq)
	}

	buff.WriteByte('@')
	buff.WriteString(seq.Id())
	buff.WriteByte(' ')

	buff.WriteString(info)
	buff.WriteByte('\n')

	buff.Write(seq.Sequence())
	buff.WriteString("\n+\n")

	q := seq.QualitiesString()
	buff.WriteString(q)
	buff.WriteByte('\n')

}

// The function FormatFastq takes a BioSequence object, a quality shift value, and a header formatter
// function as input, and returns a formatted string in FASTQ format.
func FormatFastq(seq *obiseq.BioSequence, formater FormatHeader) string {

	var buff bytes.Buffer

	_formatFastq(&buff, seq, formater)

	return buff.String()
}

func FormatFastqBatch(batch obiiter.BioSequenceBatch,
	formater FormatHeader, skipEmpty bool) []byte {
	var bs bytes.Buffer

	for i, seq := range batch.Slice() {
		if seq.Len() > 0 {
			_formatFastq(&bs, seq, formater)

			if i == 0 {

				bs.Grow(len(bs.Bytes()) * len(batch.Slice()) * 5 / 4)
			}
		} else {
			if skipEmpty {
				log.Warnf("Sequence %s is empty and skiped in output", seq.Id())
			} else {
				log.Fatalf("Sequence %s is empty", seq.Id())
			}
		}

	}

	chunk := bs.Bytes()

	return chunk
}

type FileChunck struct {
	text  []byte
	order int
}

func WriteFastq(iterator obiiter.IBioSequence,
	file io.WriteCloser,
	options ...WithOption) (obiiter.IBioSequence, error) {

	opt := MakeOptions(options)
	iterator = iterator.Rebatch(opt.BatchSize())

	file, _ = obiutils.CompressStream(file, opt.CompressedFile(), opt.CloseFile())

	newIter := obiiter.MakeIBioSequence()

	nwriters := opt.ParallelWorkers()

	obiiter.RegisterAPipe()
	chunkchan := make(chan FileChunck)

	header_format := opt.FormatFastSeqHeader()

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
				FormatFastqBatch(batch, header_format, opt.SkipEmptySequence()),
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
				if chunk.text[0] != '@' {
					log.Panicln("WriteFastq: FASTQ format error")
				}
				file.Write(chunk.text)
				next_to_send++
				chunk, ok := received[next_to_send]
				for ok {
					if chunk.text[0] != '@' {
						log.Panicln("WriteFastq: FASTQ format error")
					}
					file.Write(chunk.text)
					delete(received, next_to_send)
					next_to_send++
					chunk, ok = received[next_to_send]
				}
			} else {
				if _, ok := received[chunk.order]; ok {
					log.Panicln("WriteFastq: Two chunks with the same number")
				}
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
	} else {
		flags |= os.O_TRUNC
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
