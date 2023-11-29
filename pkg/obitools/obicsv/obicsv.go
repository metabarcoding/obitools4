package obicsv

import (
	"log"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiformats"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
)

func CLIWriteCSV(iterator obiiter.IBioSequence,
	terminalAction bool, filenames ...string) (obiiter.IBioSequence, error) {

	if obiconvert.CLIProgressBar() {
		iterator = iterator.Speed()
	}

	var newIter obiiter.IBioSequence

	opts := make([]obiformats.WithOption, 0, 10)

	nworkers := obioptions.CLIParallelWorkers() / 4
	if nworkers < 2 {
		nworkers = 2
	}

	opts = append(opts, obiformats.OptionsParallelWorkers(nworkers))
	opts = append(opts, obiformats.OptionsBatchSize(obioptions.CLIBatchSize()))
	opts = append(opts, obiformats.OptionsCompressed(obiconvert.CLICompressed()))

	opts = append(opts, obiformats.CSVId(CLIPrintId()),
		obiformats.CSVCount(CLIPrintCount()),
		obiformats.CSVTaxon(CLIPrintTaxon()),
		obiformats.CSVDefinition(CLIPrintDefinition()),
		obiformats.CSVKeys(CLIToBeKeptAttributes()),
		obiformats.CSVSequence(CLIPrintSequence()),
		obiformats.CSVAutoColumn(CLIAutoColumns()),
	)

	var err error

	if len(filenames) == 0 {
		newIter, err = obiformats.WriteCSVToStdout(iterator, opts...)
	} else {
		newIter, err = obiformats.WriteCSVToFile(iterator, filenames[0], opts...)
	}

	if err != nil {
		log.Fatalf("Write file error: %v", err)
		return obiiter.NilIBioSequence, err
	}

	if terminalAction {
		newIter.Recycle()
		return obiiter.NilIBioSequence, nil
	}

	return newIter, nil

}
