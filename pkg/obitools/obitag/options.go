package obitag

import (
	"log"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiformats"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obifind"
	"github.com/DavidGamba/go-getoptions"
)

var _RefDB = ""
var _SaveRefDB = ""
var _RunExact = false

func TagOptionSet(options *getoptions.GetOpt) {
	options.StringVar(&_RefDB, "reference-db", _RefDB,
		options.Alias("R"),
		options.Required(),
		options.ArgName("FILENAME"),
		options.Description("The name of the file containing the reference DB"))

	options.StringVar(&_SaveRefDB, "save-db", _SaveRefDB,
		options.ArgName("FILENAME"),
		options.Description("The name of a file where to save the reference DB with its indices"))

	// options.BoolVar(&_RunExact, "exact", _RunExact,
	// 	options.Alias("E"),
	// 	options.Description("Unactivate the heuristic limatitating the sequence comparisons"))

}

// OptionSet adds to the basic option set every options declared for
// the obiuniq command
func OptionSet(options *getoptions.GetOpt) {
	obiconvert.OptionSet(options)
	obifind.LoadTaxonomyOptionSet(options, true, false)
	TagOptionSet(options)
}

func CLIRefDBName() string {
	return _RefDB
}

func CLIRefDB() obiseq.BioSequenceSlice {
	refdb, err := obiformats.ReadSequencesFromFile(_RefDB)

	if err != nil {
		log.Panicf("Cannot open the reference library file : %s\n", _RefDB)
	}

	return refdb.Load()
}

func CLIShouldISaveRefDB() bool {
	return _SaveRefDB != ""
}

func CLISaveRefetenceDB(db obiseq.BioSequenceSlice) {
	if CLIShouldISaveRefDB() {
		idb := obiiter.IBatchOver(db, 1000)

		var newIter obiiter.IBioSequence

		opts := make([]obiformats.WithOption, 0, 10)

		switch obiconvert.CLIOutputFastHeaderFormat() {
		case "json":
			opts = append(opts, obiformats.OptionsFastSeqHeaderFormat(obiformats.FormatFastSeqJsonHeader))
		case "obi":
			opts = append(opts, obiformats.OptionsFastSeqHeaderFormat(obiformats.FormatFastSeqOBIHeader))
		default:
			opts = append(opts, obiformats.OptionsFastSeqHeaderFormat(obiformats.FormatFastSeqJsonHeader))
		}

		nworkers := obioptions.CLIParallelWorkers() / 4
		if nworkers < 2 {
			nworkers = 2
		}

		opts = append(opts, obiformats.OptionsParallelWorkers(nworkers))
		opts = append(opts, obiformats.OptionsBatchSize(obioptions.CLIBatchSize()))

		opts = append(opts, obiformats.OptionsQualityShift(obiconvert.CLIOutputQualityShift()))

		opts = append(opts, obiformats.OptionsCompressed(obiconvert.CLICompressed()))

		var err error

		switch obiconvert.CLIOutputFormat() {
		case "fastq":
			newIter, err = obiformats.WriteFastqToFile(idb, _SaveRefDB, opts...)
		case "fasta":
			newIter, err = obiformats.WriteFastaToFile(idb, _SaveRefDB, opts...)
		default:
			newIter, err = obiformats.WriteSequencesToFile(idb, _SaveRefDB, opts...)
		}

		if err != nil {
			log.Fatalf("Write file error: %v", err)
		}

		newIter.Recycle()
		obiiter.WaitForLastPipe()
	}
}

func CLIRunExact() bool {
	return _RunExact
}
