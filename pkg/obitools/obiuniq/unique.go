package obiuniq

import (
	"log"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obichunk"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

func Unique(sequences obiseq.IBioSequenceBatch) obiseq.IBioSequenceBatch {

	options := make([]obichunk.WithOption, 0, 30)

	options = append(options,
		obichunk.OptionBatchCount(CLINumberOfChunks()),
	)

	if CLIUniqueInMemory() {
		log.Printf("Running dereplication in memory on %d chunks", CLINumberOfChunks())
		options = append(options, obichunk.OptionSortOnMemory())
	} else {
		log.Printf("Running dereplication on disk with %d chunks", CLINumberOfChunks())
		options = append(options, obichunk.OptionSortOnDisk())
	}

	options = append(options,
		obichunk.OptionStatOn(CLIStatsOn()...))

	options = append(options,
		obichunk.OptionSubCategory(CLIKeys()...))

	options = append(options,
		obichunk.OptionsParallelWorkers(
			obioptions.CLIParallelWorkers()),
		obichunk.OptionsBufferSize(
			obioptions.CLIBufferSize()),
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
