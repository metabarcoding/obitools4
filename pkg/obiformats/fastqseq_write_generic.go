package obiformats

import (
	"bytes"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

type BioSequenceBatchFormater func(batch obiiter.BioSequenceBatch) []byte
type BioSequenceFormater func(sequence *obiseq.BioSequence) string

func BuildFastxSeqFormater(format string, header FormatHeader) BioSequenceFormater {
	var f BioSequenceFormater

	switch format {
	case "fastq":
		f = func(sequence *obiseq.BioSequence) string {
			return FormatFastq(sequence, header)
		}
	case "fasta":
		f = func(sequence *obiseq.BioSequence) string {
			return FormatFasta(sequence, header)
		}
	default:
		log.Fatal("Unknown output format")
	}

	return f
}

func BuildFastxFormater(format string, header FormatHeader) BioSequenceBatchFormater {
	fs := BuildFastxSeqFormater(format, header)

	f := func(batch obiiter.BioSequenceBatch) []byte {
		var bs bytes.Buffer
		for _, seq := range batch.Slice() {
			bs.WriteString(fs(seq))
			bs.WriteString("\n")
		}
		return bs.Bytes()
	}

	return f
}
