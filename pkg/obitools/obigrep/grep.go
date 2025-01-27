package obigrep

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
)

func CLIFilterSequence(iterator obiiter.IBioSequence) obiiter.IBioSequence {
	var newIter obiiter.IBioSequence

	predicate := CLISequenceSelectionPredicate()

	if obiconvert.CLIHasPairedFile() {
		predicate = predicate.PairedPredicat(CLIPairedReadMode())
	}

	if predicate != nil {
		if CLISaveDiscardedSequences() {
			var discarded obiiter.IBioSequence

			log.Printf("Discarded sequences saved in file: %s\n", CLIDiscardedFileName())
			newIter, discarded = iterator.DivideOn(predicate,
				obidefault.BatchSize())

			go func() {
				_, err := obiconvert.CLIWriteBioSequences(discarded,
					true,
					CLIDiscardedFileName())

				if err != nil {
					log.Fatalf("%v", err)
				}
			}()

		} else {
			newIter = iterator.FilterOn(predicate,
				obidefault.BatchSize(),
				obidefault.ParallelWorkers(),
			)
		}
	} else {
		newIter = iterator
	}

	return newIter

}
