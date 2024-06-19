package obiformats

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

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

func FormatJSONBatch(batch obiiter.BioSequenceBatch) []byte {
	buff := new(bytes.Buffer)
	json := bufio.NewWriter(buff)
	n := batch.Slice().Len() - 1
	for i, s := range batch.Slice() {
		json.WriteString("  ")
		json.Write(JSONRecord(s))
		if i < n {
			json.WriteString(",\n")
		}
	}

	json.Flush()

	return buff.Bytes()
}

func WriteJSON(iterator obiiter.IBioSequence,
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
				FormatJSONBatch(batch),
				batch.Order(),
			}
			newIter.Push(batch)
		}
		newIter.Done()
	}

	next_to_send := 0
	received := make(map[int]FileChunck, 100)

	waitWriter.Add(1)
	go func() {
		for chunk := range chunkchan {
			if chunk.order == next_to_send {
				if next_to_send > 0 {
					file.Write([]byte(",\n"))
				}
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

		file.Write([]byte("\n]\n"))
		file.Close()

		log.Debugln("End of the JSON file writing")
		obiiter.UnregisterPipe()
		waitWriter.Done()

	}()

	log.Debugln("Start of the JSON file writing")
	file.Write([]byte("[\n"))
	go ff(iterator)
	for i := 0; i < nwriters-1; i++ {
		go ff(iterator.Split())
	}

	return newIter, nil
}

func WriteJSONToStdout(iterator obiiter.IBioSequence,
	options ...WithOption) (obiiter.IBioSequence, error) {
	options = append(options, OptionDontCloseFile())
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
