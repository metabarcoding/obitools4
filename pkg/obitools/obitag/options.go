package obitag

import (
	"log"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiformats"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obifind"
	"github.com/DavidGamba/go-getoptions"
)

var _RefDB = ""
var _RunExact = false

func TagOptionSet(options *getoptions.GetOpt) {
	options.StringVar(&_RefDB, "reference-db", _RefDB,
		options.Alias("R"),
		options.Required(),
		options.ArgName("FILENAME"),
		options.Description("The name of the file containing the reference DB"))

	options.BoolVar(&_RunExact, "exact", _RunExact,
		options.Alias("E"),
		options.Description("Unactivate the heuristic limatitating the sequence comparisons"))

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
	refdb, err := obiformats.ReadSequencesBatchFromFile(_RefDB)

	if err != nil {
		log.Panicf("Cannot open the reference library file : %s\n", _RefDB)
	}

	return refdb.Load()
}

func CLIRunExact() bool {
	return _RunExact
}
