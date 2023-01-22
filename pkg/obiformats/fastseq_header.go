package obiformats

import (
	"strings"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

func ParseGuessedFastSeqHeader(sequence *obiseq.BioSequence) {
	if strings.HasPrefix(sequence.Definition(), "{") {
		ParseFastSeqJsonHeader(sequence)
	} else {
		ParseFastSeqOBIHeader(sequence)
	}
}

func IParseFastSeqHeaderBatch(iterator obiiter.IBioSequence,
	options ...WithOption) obiiter.IBioSequence {
	opt := MakeOptions(options)
	return iterator.MakeIWorker(obiseq.AnnotatorToSeqWorker(opt.ParseFastSeqHeader()),
		opt.ParallelWorkers(),
		opt.BufferSize())
}
