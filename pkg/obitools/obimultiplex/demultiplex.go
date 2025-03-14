package obimultiplex

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obingslibrary"
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
		obingslibrary.OptionParallelWorkers(obidefault.ParallelWorkers()),
		obingslibrary.OptionBatchSize(obidefault.BatchSize()),
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
		out = out.FilterOn(obiseq.HasAttribute("obimultiplex_error").Not(), obidefault.BatchSize())
	}

	var unidentified obiiter.IBioSequence
	if CLIUnidentifiedFileName() != "" {
		log.Printf("Unassigned sequences saved in file: %s\n", CLIUnidentifiedFileName())
		unidentified, out = newIter.DivideOn(obiseq.HasAttribute("obimultiplex_error"),
			obidefault.BatchSize())

		go func() {
			_, err := obiconvert.CLIWriteBioSequences(unidentified,
				true,
				CLIUnidentifiedFileName())

			if err != nil {
				log.Fatalf("%v", err)
			}
		}()

	}
	log.Printf("Sequence demultiplexing using %d workers\n", obidefault.ParallelWorkers())

	return out, nil
}
