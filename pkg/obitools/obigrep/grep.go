package obigrep

import (
	"log"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
)

func IFilterSequence(iterator obiiter.IBioSequenceBatch) obiiter.IBioSequenceBatch {
	var newIter obiiter.IBioSequenceBatch

	predicate := CLISequenceSelectionPredicate()

	if predicate != nil {
		if CLISaveDiscardedSequences() {
			var discarded obiiter.IBioSequenceBatch

			log.Printf("Discarded sequences saved in file: %s\n", CLIDiscardedFileName())
			newIter, discarded = iterator.DivideOn(predicate,
				obioptions.CLIBatchSize())

			go func() {
				_, err := obiconvert.WriteBioSequences(discarded,
					true,
					CLIDiscardedFileName())

				if err != nil {
					log.Fatalf("%v", err)
				}
			}()

		} else {
			newIter = iterator.FilterOn(predicate,
				obioptions.CLIBatchSize(),
				obioptions.CLIParallelWorkers(),
				obioptions.CLIBufferSize(),
			)
		}
	} else {
		newIter = iterator
	}

	return newIter

}
