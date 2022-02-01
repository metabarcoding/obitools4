package obimultiplex

import (
	"log"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obingslibrary"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
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
		log.Println("Discards unassigned sequences")
		newIter = newIter.Rebatch(obioptions.CLIBatchSize())
	}

	var unidentified obiseq.IBioSequenceBatch
	if CLIUnidentifiedFileName() != "" {
		log.Printf("Unassigned sequences saved in file: %s\n", CLIUnidentifiedFileName())
		unidentified, newIter = newIter.DivideOn(obiseq.HasAttribute("demultiplex_error"),
			obioptions.CLIBatchSize())

		go func() {
			_, err := obiconvert.WriteBioSequencesBatch(unidentified,
				true,
				CLIUnidentifiedFileName())

			if err != nil {
				log.Fatalf("%v", err)
			}
		}()

	}
	log.Printf("Sequence demultiplexing using %d workers\n", obioptions.CLIParallelWorkers())

	return newIter, nil
}
