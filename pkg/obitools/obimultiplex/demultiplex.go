package obimultiplex

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obingslibrary"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
)

func IExtractBarcode(iterator obiiter.IBioSequence) (obiiter.IBioSequence, error) {

	opts := make([]obingslibrary.WithOption, 0, 10)

	opts = append(opts,
		obingslibrary.OptionAllowedMismatches(CLIAllowedMismatch()),
		obingslibrary.OptionAllowedIndel(CLIAllowsIndel()),
		obingslibrary.OptionUnidentified(CLIUnidentifiedFileName()),
		obingslibrary.OptionDiscardErrors(!CLIConservedErrors()),
		obingslibrary.OptionParallelWorkers(obioptions.CLIParallelWorkers()),
		obingslibrary.OptionBatchSize(obioptions.CLIBatchSize()),
	)

	ngsfilter, err := CLINGSFIlter()
	if err != nil {
		log.Fatalf("%v", err)
	}

	worker := ngsfilter.ExtractMultiBarcodeSliceWorker(opts...)

	newIter := iterator.MakeISliceWorker(worker, false)
	out := newIter

	if !CLIConservedErrors() {
		log.Infoln("Discards unassigned sequences")
		out = out.FilterOn(obiseq.HasAttribute("demultiplex_error").Not(), obioptions.CLIBatchSize())
	}

	var unidentified obiiter.IBioSequence
	if CLIUnidentifiedFileName() != "" {
		log.Printf("Unassigned sequences saved in file: %s\n", CLIUnidentifiedFileName())
		unidentified, out = newIter.DivideOn(obiseq.HasAttribute("demultiplex_error"),
			obioptions.CLIBatchSize())

		go func() {
			_, err := obiconvert.CLIWriteBioSequences(unidentified,
				true,
				CLIUnidentifiedFileName())

			if err != nil {
				log.Fatalf("%v", err)
			}
		}()

	}
	log.Printf("Sequence demultiplexing using %d workers\n", obioptions.CLIParallelWorkers())

	return out, nil
}
