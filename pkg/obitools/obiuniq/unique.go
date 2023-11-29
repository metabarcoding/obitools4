package obiuniq

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obichunk"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
)

func Unique(sequences obiiter.IBioSequence) obiiter.IBioSequence {

	options := make([]obichunk.WithOption, 0, 30)

	options = append(options,
		obichunk.OptionBatchCount(CLINumberOfChunks()),
	)

	//
	// Considers if data splitting must be done on disk or in memory
	//
	// --on-disk command line option

	if CLIUniqueInMemory() {
		log.Printf("Running dereplication in memory on %d chunks", CLINumberOfChunks())
		options = append(options, obichunk.OptionSortOnMemory())
	} else {
		log.Printf("Running dereplication on disk with %d chunks", CLINumberOfChunks())
		options = append(options, obichunk.OptionSortOnDisk())
	}

	//
	// Considers if sequences observed a single time in the dataset have to
	// be conserved in the output
	//
	// --no-singleton

	if CLINoSingleton() {
		log.Printf("Removing sigletons from the output")
		options = append(options, obichunk.OptionsNoSingleton())
	} else {
		log.Printf("Keep sigletons in the output")
	}

	options = append(options,
		obichunk.OptionStatOn(CLIStatsOn()...))

	options = append(options,
		obichunk.OptionSubCategory(CLIKeys()...))

	options = append(options,
		obichunk.OptionsParallelWorkers(
			obioptions.CLIParallelWorkers()),
		obichunk.OptionsBatchSize(
			obioptions.CLIBatchSize()),
		obichunk.OptionNAValue(CLINAValue()),
	)

	iUnique, err := obichunk.IUniqueSequence(sequences, options...)

	if err != nil {
		log.Fatal(err)
	}

	return iUnique
}
