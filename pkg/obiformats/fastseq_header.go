package obiformats

import (
	"strings"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

func ParseGuessedFastSeqHeader(sequence *obiseq.BioSequence) {
	if strings.HasPrefix(sequence.Definition(), "{") {
		ParseFastSeqJsonHeader(sequence)
	} else {
		ParseFastSeqOBIHeader(sequence)
	}
}

func IParseFastSeqHeaderBatch(iterator obiseq.IBioSequenceBatch, options ...WithOption) obiseq.IBioSequenceBatch {
	opt := MakeOptions(options)
	return iterator.MakeIWorker(obiseq.AnnotatorToSeqWorker(opt.ParseFastSeqHeader()),
		opt.ParallelWorkers(),
		opt.BufferSize())
}

func IParseFastSeqHeader(iterator obiseq.IBioSequence, options ...WithOption) obiseq.IBioSequence {
	opt := MakeOptions(options)

	return IParseFastSeqHeaderBatch(iterator.IBioSequenceBatch(opt.BatchSize(),
		opt.BufferSize()),
		options...).SortBatches().IBioSequence()
}
