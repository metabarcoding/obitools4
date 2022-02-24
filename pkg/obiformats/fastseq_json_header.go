package obiformats

import (
	log "github.com/sirupsen/logrus"
	"strings"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"github.com/goccy/go-json"
)

func _parse_json_header_(header string, annotations obiseq.Annotation) string {

	start := -1
	stop := -1
	level := 0
	lh := len(header)

	for i := 0; (i < lh) && (stop < 0); i++ {
		// fmt.Printf("[%d,%d-%d] : %d (%c) (%d,%c)\n", i, start, stop, header[i], header[i], '{', '{')
		if level == 0 && header[i] == '{' {
			start = i
		}

		if header[i] == '{' {
			level++
		}

		if header[i] == '}' {
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
	if err != nil {
		log.Fatalf("annotation parsing error on %s : %v\n", header, err)
	}

	return strings.TrimSpace(header[stop:])
}

func ParseFastSeqJsonHeader(sequence *obiseq.BioSequence) {
	sequence.SetDefinition(_parse_json_header_(sequence.Definition(),
		sequence.Annotations()))
}

func FormatFastSeqJsonHeader(sequence *obiseq.BioSequence) string {
	annotations := sequence.Annotations()

	if annotations != nil {
		text, err := json.Marshal(sequence.Annotations())

		if err != nil {
			panic(err)
		}

		return string(text)
	}

	return ""
}
