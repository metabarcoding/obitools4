package obimultiplex

import (
	"log"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obingslibrary"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

func IExtractBarcodeBatches(iterator obiseq.IBioSequenceBatch) (obiseq.IBioSequenceBatch, error) {

	opts := make([]obingslibrary.WithOption, 0, 10)

	opts = append(opts,
		obingslibrary.OptionAllowedMismatches(CLIAllowedMismatch()),
		obingslibrary.OptionUnidentified(CLIUnidentifiedFileName()),
		obingslibrary.OptionDiscardErrors(!CLIConservedErrors()),
		obingslibrary.OptionParallelWorkers(obioptions.CLIParallelWorkers()),
		obingslibrary.OptionBatchSize(obioptions.CLIBatchSize()),
		obingslibrary.OptionBufferSize(obioptions.CLIBufferSize()),
	)

	ngsfilter, err := CLINGSFIlter()
	if err != nil {
		log.Fatalf("%v", err)
	}

	worker := obingslibrary.ExtractBarcodeSliceWorker(ngsfilter, opts...)

	newIter := iterator.MakeISliceWorker(worker)

	if !CLIConservedErrors() {
		newIter = newIter.Rebatch(obioptions.CLIBatchSize())
	}

	log.Printf("Sequence demultiplexing using %d workers\n", obioptions.CLIParallelWorkers())

	return newIter, nil
}
