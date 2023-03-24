package obiformats

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiutils"
	log "github.com/sirupsen/logrus"
)

func CSVRecord(sequence *obiseq.BioSequence, opt Options) []string {
	keys := opt.CSVKeys()
	record := make([]string, 0, len(keys)+4)

	if opt.CSVId() {
		record = append(record, sequence.Id())
	}

	if opt.CSVCount() {
		record = append(record, fmt.Sprint(sequence.Count()))
	}

	if opt.CSVTaxon() {
		taxid := sequence.Taxid()
		sn, ok := sequence.GetAttribute("scientific_name")

		if !ok {
			if taxid == 1 {
				sn = "root"
			} else {
				sn = opt.CSVNAValue()
			}
		}

		record = append(record, fmt.Sprint(taxid), fmt.Sprint(sn))
	}

	if opt.CSVDefinition() {
		record = append(record, sequence.Definition())
	}

	for _, key := range opt.CSVKeys() {
		value, ok := sequence.GetAttribute(key)
		if !ok {
			value = opt.CSVNAValue()
		}

		svalue, _ := obiutils.InterfaceToString(value)
		record = append(record, svalue)
	}

	if opt.CSVSequence() {
		record = append(record, string(sequence.Sequence()))
	}

	if opt.CSVQuality() {
		if sequence.HasQualities() {
			l := sequence.Len()
			q := sequence.Qualities()
			ascii := make([]byte, l)
			quality_shift := opt.QualityShift()
			for j := 0; j < l; j++ {
				ascii[j] = uint8(q[j]) + uint8(quality_shift)
			}
			record = append(record, string(ascii))
		} else {
			record = append(record, opt.CSVNAValue())
		}
	}

	return record
}

func CSVHeader(opt Options) []string {
	keys := opt.CSVKeys()
	record := make([]string, 0, len(keys)+4)

	if opt.CSVId() {
		record = append(record, "id")
	}

	if opt.CSVCount() {
		record = append(record, "count")
	}

	if opt.CSVTaxon() {
		record = append(record, "taxid", "scientific_name")
	}

	if opt.CSVDefinition() {
		record = append(record, "definition")
	}

	record = append(record, opt.CSVKeys()...)

	if opt.CSVSequence() {
		record = append(record, "sequence")
	}

	if opt.CSVQuality() {
		record = append(record, "quality")
	}

	return record
}

func FormatCVSBatch(batch obiiter.BioSequenceBatch, opt Options) []byte {
	buff := new(bytes.Buffer)
	csv := csv.NewWriter(buff)

	if batch.Order() == 0 {
		csv.Write(CSVHeader(opt))
	}
	for _, s := range batch.Slice() {
		csv.Write(CSVRecord(s, opt))
	}

	csv.Flush()

	return buff.Bytes()
}

func WriteCSV(iterator obiiter.IBioSequence,
	file io.WriteCloser,
	options ...WithOption) (obiiter.IBioSequence, error) {
	opt := MakeOptions(options)

	file, _ = obiutils.CompressStream(file, opt.CompressedFile(), opt.CloseFile())

	newIter := obiiter.MakeIBioSequence()

	nwriters := opt.ParallelWorkers()

	obiiter.RegisterAPipe()
	chunkchan := make(chan FileChunck)

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

			chunkchan <- FileChunck{
				FormatCVSBatch(batch, opt),
				batch.Order(),
			}
			newIter.Push(batch)
		}
		newIter.Done()
	}

	log.Debugln("Start of the CSV file writing")
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

		log.Debugln("End of the CSV file writing")
		obiiter.UnregisterPipe()
		waitWriter.Done()

	}()

	return newIter, nil
}

func WriteCSVToStdout(iterator obiiter.IBioSequence,
	options ...WithOption) (obiiter.IBioSequence, error) {
	options = append(options, OptionDontCloseFile())
	return WriteCSV(iterator, os.Stdout, options...)
}

func WriteCSVToFile(iterator obiiter.IBioSequence,
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

	iterator, err = WriteCSV(iterator, file, options...)

	if opt.HaveToSavePaired() {
		var revfile *os.File

		revfile, err = os.OpenFile(opt.PairedFileName(), flags, 0660)
		if err != nil {
			log.Fatalf("open file error: %v", err)
			return obiiter.NilIBioSequence, err
		}
		iterator, err = WriteCSV(iterator.PairedWith(), revfile, options...)
	}

	return iterator, err
}
