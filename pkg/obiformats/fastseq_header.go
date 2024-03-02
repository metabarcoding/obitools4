package obiformats

import (
	"strings"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// ParseGuessedFastSeqHeader parses the guessed fast sequence header.
//
// The function takes a pointer to a BioSequence object as its parameter.
// It determines whether the sequence definition starts with "{" or not.
// If it does, it calls the ParseFastSeqJsonHeader function.
// If it doesn't, it calls the ParseFastSeqOBIHeader function.
func ParseGuessedFastSeqHeader(sequence *obiseq.BioSequence) {
	if strings.HasPrefix(sequence.Definition(), "{") {
		// Sequence definition starts with "{"
		ParseFastSeqJsonHeader(sequence)
	} else {
		// Sequence definition does not start with "{"
		ParseFastSeqOBIHeader(sequence)
	}
}

// IParseFastSeqHeaderBatch is a function that processes a batch of biosequences and returns an iterator of biosequences.
//
// It takes an iterator of biosequences as the first parameter and accepts optional options as variadic arguments.
// The function returns an iterator of biosequences.
func IParseFastSeqHeaderBatch(iterator obiiter.IBioSequence,
	options ...WithOption) obiiter.IBioSequence {
	opt := MakeOptions(options)
	return iterator.MakeIWorker(obiseq.AnnotatorToSeqWorker(opt.ParseFastSeqHeader()),
		false,
		opt.ParallelWorkers())
}
