package obiformats

import (
	"encoding/csv"
	"io"
	"os"
	"path"
	"strings"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	"github.com/goccy/go-json"
	log "github.com/sirupsen/logrus"
)

func _ParseCsvFile(source string,
	reader io.Reader,
	out obiiter.IBioSequence,
	shift byte,
	batchSize int) {

	file := csv.NewReader(reader)

	file.Comma = ','
	file.ReuseRecord = false
	file.LazyQuotes = true
	file.Comment = '#'
	file.FieldsPerRecord = -1
	file.TrimLeadingSpace = true

	header, err := file.Read()

	if err != nil {
		if err == io.EOF {
			out.Done()
			return
		}
		log.Fatal(err)
	}

	sequenceColIndex := -1
	idColIndex := -1
	qualitiesColIndex := -1
	o := 0

	for i, colName := range header {
		switch colName {
		case "sequence":
			sequenceColIndex = i
		case "id":
			idColIndex = i
		case "qualities":
			qualitiesColIndex = i
		}
	}

	file.ReuseRecord = true
	slice := obiseq.MakeBioSequenceSlice()

	for {
		rec, err := file.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		sequence := obiseq.NewEmptyBioSequence(0)

		if sequenceColIndex >= 0 {
			sequence.SetSequence([]byte(rec[sequenceColIndex]))
		}

		if idColIndex >= 0 {
			sequence.SetId(rec[idColIndex])
		}

		if qualitiesColIndex >= 0 {
			q := []byte(rec[qualitiesColIndex])

			for i := 0; i < len(q); i++ {
				q[i] -= shift
			}
			sequence.SetQualities(q)
		}

		for i, field := range rec {
			var val interface{}

			if i == sequenceColIndex || i == idColIndex || i == qualitiesColIndex {
				continue
			}

			ft := header[i]

			switch {
			case ft == "taxid":
				sequence.SetTaxid(field)
			case strings.HasSuffix(ft, "_taxid"):
				sequence.SetTaxid(field, strings.TrimSuffix(ft, "_taxid"))
			default:
				err := json.Unmarshal(obiutils.UnsafeBytes(field), &val)

				if err != nil {
					val = field
				} else {
					if _, ok := val.(float64); ok {
						if obiutils.IsIntegral(val.(float64)) {
							val = int(val.(float64))
						}
					}
				}

				sequence.SetAttribute(ft, val)
			}
		}

		slice = append(slice, sequence)
		if len(slice) >= batchSize {
			out.Push(obiiter.MakeBioSequenceBatch(source, o, slice))
			o++
			slice = obiseq.MakeBioSequenceSlice()
		}
	}

	if len(slice) > 0 {
		out.Push(obiiter.MakeBioSequenceBatch(source, o, slice))
	}

	out.Done()

}

func ReadCSV(reader io.Reader, options ...WithOption) (obiiter.IBioSequence, error) {

	opt := MakeOptions(options)
	out := obiiter.MakeIBioSequence()

	out.Add(1)
	go _ParseCsvFile(opt.Source(),
		reader,
		out,
		byte(obioptions.InputQualityShift()),
		opt.BatchSize())

	go func() {
		out.WaitAndClose()
	}()

	return out, nil

}

func ReadCSVFromFile(filename string, options ...WithOption) (obiiter.IBioSequence, error) {

	options = append(options, OptionsSource(obiutils.RemoveAllExt((path.Base(filename)))))
	file, err := Ropen(filename)

	if err == ErrNoContent {
		log.Infof("file %s is empty", filename)
		return ReadEmptyFile(options...)
	}

	if err != nil {
		return obiiter.NilIBioSequence, err
	}

	return ReadCSV(file, options...)
}

func ReadCSVFromStdin(reader io.Reader, options ...WithOption) (obiiter.IBioSequence, error) {
	options = append(options, OptionsSource(obiutils.RemoveAllExt("stdin")))
	input, err := Buf(os.Stdin)

	if err == ErrNoContent {
		log.Infof("stdin is empty")
		return ReadEmptyFile(options...)
	}

	if err != nil {
		log.Fatalf("open file error: %v", err)
		return obiiter.NilIBioSequence, err
	}

	return ReadCSV(input, options...)
}
