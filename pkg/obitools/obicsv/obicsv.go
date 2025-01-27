package obicsv

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiitercsv"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
)

func CLIWriteSequenceCSV(iterator obiiter.IBioSequence,
	terminalAction bool, filenames ...string) *obiitercsv.ICSVRecord {

	if obiconvert.CLIProgressBar() {
		iterator = iterator.Speed("Writing CSV")
	}

	opts := make([]WithOption, 0, 10)

	nworkers := obidefault.ParallelWorkers() / 4
	if nworkers < 2 {
		nworkers = 2
	}

	opts = append(opts, OptionsParallelWorkers(nworkers))
	opts = append(opts, OptionsBatchSize(obidefault.BatchSize()))
	opts = append(opts, OptionsCompressed(obidefault.CompressOutput()))

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

func CLICSVWriter(iterator *obiitercsv.ICSVRecord,
	terminalAction bool,
	options ...WithOption) *obiitercsv.ICSVRecord {

	var err error
	var newIter *obiitercsv.ICSVRecord

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
