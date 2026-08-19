package obiformats

import (
	"io"
	"os"
	"path"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	"github.com/buger/jsonparser"
	"github.com/goccy/go-json"
	log "github.com/sirupsen/logrus"
)

// _parse_json_record parses a single JSON object describing a sequence
// (as produced by JSONRecord in json_writer.go) into a *obiseq.BioSequence.
func _parse_json_record(raw []byte, shift byte) *obiseq.BioSequence {
	sequence := obiseq.NewEmptyBioSequence(0)

	if id, err := jsonparser.GetString(raw, "id"); err == nil {
		sequence.SetId(id)
	}

	if seq, err := jsonparser.GetString(raw, "sequence"); err == nil {
		sequence.SetSequence([]byte(seq))
	}

	if qual, err := jsonparser.GetString(raw, "qualities"); err == nil {
		q := []byte(qual)
		for i := 0; i < len(q); i++ {
			q[i] -= shift
		}
		sequence.SetQualities(q)
	}

	if annot, dataType, _, err := jsonparser.Get(raw, "annotations"); err == nil && dataType == jsonparser.Object {
		jsonparser.ObjectEach(annot,
			func(key []byte, value []byte, valType jsonparser.ValueType, offset int) error {
				return _parse_json_annotation_field(key, value, valType, sequence)
			},
		)
	}

	return sequence
}

// _ParseJsonFile streams the top-level JSON array, decoding and pushing one
// batch of sequences at a time, without ever loading the whole document in
// memory. Only one raw record at a time is buffered by the decoder.
func _ParseJsonFile(source string,
	reader io.Reader,
	out obiiter.IBioSequence,
	shift byte,
	batchSize int) {

	dec := json.NewDecoder(reader)

	if _, err := dec.Token(); err != nil {
		if err == io.EOF {
			out.Done()
			return
		}
		log.Fatalf("cannot parse JSON data: %v", err)
	}

	slice := obiseq.MakeBioSequenceSlice()
	o := 0

	for dec.More() {
		var raw json.RawMessage

		if err := dec.Decode(&raw); err != nil {
			log.Fatalf("cannot parse JSON data: %v", err)
		}

		sequence := _parse_json_record(raw, shift)

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

func ReadJSON(reader io.Reader, options ...WithOption) (obiiter.IBioSequence, error) {

	opt := MakeOptions(options)
	out := obiiter.MakeIBioSequence()

	out.Add(1)
	go _ParseJsonFile(opt.Source(),
		reader,
		out,
		obidefault.ReadQualitiesShift(),
		opt.BatchSize())

	go func() {
		out.WaitAndClose()
	}()

	return out, nil

}

func ReadJSONFromFile(filename string, options ...WithOption) (obiiter.IBioSequence, error) {

	options = append(options, OptionsSource(obiutils.RemoveAllExt((path.Base(filename)))))
	file, err := obiutils.Ropen(filename)

	if err == obiutils.ErrNoContent {
		log.Infof("file %s is empty", filename)
		return ReadEmptyFile(options...)
	}

	if err != nil {
		return obiiter.NilIBioSequence, err
	}

	return ReadJSON(file, options...)
}

func ReadJSONFromStdin(reader io.Reader, options ...WithOption) (obiiter.IBioSequence, error) {
	options = append(options, OptionsSource(obiutils.RemoveAllExt("stdin")))
	input, err := obiutils.Buf(os.Stdin)

	if err == obiutils.ErrNoContent {
		log.Infof("stdin is empty")
		return ReadEmptyFile(options...)
	}

	if err != nil {
		log.Fatalf("open file error: %v", err)
		return obiiter.NilIBioSequence, err
	}

	return ReadJSON(input, options...)
}
