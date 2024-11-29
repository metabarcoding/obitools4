package obiformats

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/goccy/go-json"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	log "github.com/sirupsen/logrus"
)

func _UnescapeUnicodeCharactersInJSON(_jsonRaw []byte) ([]byte, error) {
	str, err := strconv.Unquote(strings.Replace(strconv.Quote(string(_jsonRaw)), `\\u`, `\u`, -1))
	if err != nil {
		return nil, err
	}
	return []byte(str), nil
}

func JSONRecord(sequence *obiseq.BioSequence) []byte {
	record := make(map[string]interface{}, 4)

	record["id"] = sequence.Id()

	if sequence.HasSequence() {
		record["sequence"] = sequence.String()
	}

	if sequence.HasQualities() {
		record["qualities"] = sequence.QualitiesString()
	}

	if sequence.HasAnnotation() {
		record["annotations"] = sequence.Annotations()
	}

	text, error := json.MarshalIndent(record, "  ", "  ")

	if error != nil {
		log.Panicf("conversion to JSON error on sequence id %s", sequence.Id())
	}

	text, error = _UnescapeUnicodeCharactersInJSON(text)

	if error != nil {
		log.Panicf("conversion to JSON error on sequence id %s", sequence.Id())
	}

	return text
}

func FormatJSONBatch(batch obiiter.BioSequenceBatch) *bytes.Buffer {
	buff := new(bytes.Buffer)

	json := bufio.NewWriter(buff)

	if batch.Order() == 0 {
		json.WriteString("[\n")
	} else {
		json.WriteString(",\n")
	}

	n := batch.Slice().Len() - 1
	for i, s := range batch.Slice() {
		json.WriteString("  ")
		json.Write(JSONRecord(s))
		if i < n {
			json.WriteString(",\n")
		}
	}

	json.Flush()
	return buff
}

func WriteJSON(iterator obiiter.IBioSequence,
	file io.WriteCloser,
	options ...WithOption) (obiiter.IBioSequence, error) {
	var latestChunk atomic.Int64

	opt := MakeOptions(options)

	file, _ = obiutils.CompressStream(file, opt.CompressedFile(), opt.CloseFile())

	newIter := obiiter.MakeIBioSequence()
	nwriters := opt.ParallelWorkers()

	chunkchan := WriteFileChunk(file, opt.CloseFile())
	newIter.Add(nwriters)

	go func() {
		newIter.WaitAndClose()

		chunkchan <- FileChunk{
			Source: "end",
			Raw:    bytes.NewBuffer([]byte("\n]\n")),
			Order:  int(latestChunk.Load()) + 1,
		}
		for len(chunkchan) > 0 {
			time.Sleep(time.Millisecond)
		}
		close(chunkchan)
	}()

	ff := func(iterator obiiter.IBioSequence) {
		for iterator.Next() {

			batch := iterator.Get()

			ss := FileChunk{
				Source: batch.Source(),
				Raw:    FormatJSONBatch(batch),
				Order:  batch.Order(),
			}

			chunkchan <- ss
			latestChunk.Store(int64(batch.Order()))
			newIter.Push(batch)
		}
		newIter.Done()
	}

	log.Debugln("Start of the JSON file writing")
	for i := 1; i < nwriters; i++ {
		go ff(iterator.Split())
	}
	go ff(iterator)

	return newIter, nil
}

func WriteJSONToStdout(iterator obiiter.IBioSequence,
	options ...WithOption) (obiiter.IBioSequence, error) {
	options = append(options, OptionCloseFile())

	return WriteJSON(iterator, os.Stdout, options...)
}

func WriteJSONToFile(iterator obiiter.IBioSequence,
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

	iterator, err = WriteJSON(iterator, file, options...)

	if opt.HaveToSavePaired() {
		var revfile *os.File

		revfile, err = os.OpenFile(opt.PairedFileName(), flags, 0660)
		if err != nil {
			log.Fatalf("open file error: %v", err)
			return obiiter.NilIBioSequence, err
		}
		iterator, err = WriteJSON(iterator.PairedWith(), revfile, options...)
	}

	return iterator, err
}
