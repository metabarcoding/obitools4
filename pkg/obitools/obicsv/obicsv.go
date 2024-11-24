package obicsv

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
)

func CLIWriteSequenceCSV(iterator obiiter.IBioSequence,
	terminalAction bool, filenames ...string) *ICSVRecord {

	if obiconvert.CLIProgressBar() {
		iterator = iterator.Speed("Writing CSV")
	}

	opts := make([]WithOption, 0, 10)

	nworkers := obioptions.CLIParallelWorkers() / 4
	if nworkers < 2 {
		nworkers = 2
	}

	opts = append(opts, OptionsParallelWorkers(nworkers))
	opts = append(opts, OptionsBatchSize(obioptions.CLIBatchSize()))
	opts = append(opts, OptionsCompressed(obiconvert.CLICompressed()))

	opts = append(opts, CSVId(CLIPrintId()),
		CSVCount(CLIPrintCount()),
		CSVTaxon(CLIPrintTaxon()),
		CSVDefinition(CLIPrintDefinition()),
		CSVKeys(CLIToBeKeptAttributes()),
		CSVSequence(CLIPrintSequence()),
		CSVAutoColumn(CLIAutoColumns()),
	)

	csvIter := NewCSVSequenceIterator(iterator, opts...)
	newIter := CLICSVWriter(csvIter, terminalAction, opts...)

	return newIter

}

func CLICSVWriter(iterator *ICSVRecord,
	terminalAction bool,
	options ...WithOption) *ICSVRecord {

	var err error
	var newIter *ICSVRecord

	if obiconvert.CLIOutPutFileName() != "-" {
		options = append(options, OptionFileName(obiconvert.CLIOutPutFileName()))
	}

	opt := MakeOptions(options)

	if opt.FileName() != "-" {
		newIter, err = WriteCSVToFile(iterator, opt.FileName(), options...)

		if err != nil {
			log.Fatalf("Cannot write to file : %+v", err)
		}

	} else {
		newIter, err = WriteCSVToStdout(iterator, options...)

		if err != nil {
			log.Fatalf("Cannot write to stdout : %+v", err)
		}

	}

	if terminalAction {
		newIter.Consume()
		return nil
	}

	return newIter
}
