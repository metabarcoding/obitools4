package obiformats

import (
	"bytes"
	"math"
	"strings"
	"unsafe"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	"github.com/goccy/go-json"
)

func _parse_json_header_(header string, annotations obiseq.Annotation) string {
	start := -1
	stop := -1
	level := 0
	lh := len(header)
	inquote := false

	for i := 0; (i < lh) && (stop < 0); i++ {
		// fmt.Printf("[%d,%d-%d] : %d (%c) (%d,%c)\n", i, start, stop, header[i], header[i], '{', '{')
		if level == 0 && header[i] == '{' && !inquote {
			start = i
		}

		// TODO: escaped double quotes are not considered
		if start > -1 && header[i] == '"' {
			inquote = !inquote
		}

		if header[i] == '{' && !inquote {
			level++
		}

		if header[i] == '}' && !inquote {
			level--
		}

		if start >= 0 && level == 0 {
			stop = i
		}

	}

	if start < 0 || stop < 0 {
		return header
	}

	stop++

	err := json.Unmarshal([]byte(header)[start:stop], &annotations)

	for k, v := range annotations {
		switch vt := v.(type) {
		case float64:
			if vt == math.Floor(vt) {
				annotations[k] = int(vt)
			}
			{
				annotations[k] = vt
			}
		}
	}

	if err != nil {
		log.Fatalf("annotation parsing error on %s : %v\n", header, err)
	}

	return strings.TrimSpace(header[stop:])
}

func ParseFastSeqJsonHeader(sequence *obiseq.BioSequence) {
	definition := sequence.Definition()
	sequence.SetDefinition("")

	definition_part := _parse_json_header_(
		definition,
		sequence.Annotations())
	if len(definition_part) > 0 {
		if sequence.HasDefinition() {
			definition_part = sequence.Definition() + " " + definition_part
		}
		sequence.SetDefinition(definition_part)
	}
}

func WriteFastSeqJsonHeader(buffer *bytes.Buffer, sequence *obiseq.BioSequence) {

	annotations := sequence.Annotations()

	if len(annotations) > 0 {
		err := obiutils.JsonMarshalByteBuffer(buffer, sequence.Annotations())

		if err != nil {
			log.Fatal(err)
		}
	}
}

func FormatFastSeqJsonHeader(sequence *obiseq.BioSequence) string {
	annotations := sequence.Annotations()
	buffer := bytes.Buffer{}

	if len(annotations) > 0 {
		obiutils.JsonMarshalByteBuffer(&buffer, sequence.Annotations())
		return unsafe.String(unsafe.SliceData(buffer.Bytes()), len(buffer.Bytes()))
	}

	return ""
}
