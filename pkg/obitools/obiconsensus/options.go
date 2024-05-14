package obiconsensus

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var _saveGraph = "__@@NOSAVE@@__"
var _kmerSize = -1

func ObiconsensusOptionSet(options *getoptions.GetOpt) {

	options.StringVar(&_saveGraph, "save-graph", _saveGraph,
		options.Description("Creates a directory containing the set of De Bruijn graphs used by "+
			"the obiconsensus algorithm. "+
			"The graph files follow the graphml format."),
	)

	options.IntVar(&_kmerSize, "kmer-size", _kmerSize,
		options.ArgName("SIZE"),
		options.Description("The size of the kmer used to build the consensus. "+
			"Default value = -1, which means that the kmer size is estimated from the data"),
	)

}

func OptionSet(options *getoptions.GetOpt) {
	obiconvert.InputOptionSet(options)
	obiconvert.OutputOptionSet(options)
	ObiconsensusOptionSet(options)
}

// Returns true it the obliclean graphs must be saved
func CLISaveGraphToFiles() bool {
	return _saveGraph != "__@@NOSAVE@@__"
}

// It returns the directory where the graph files are saved
func CLIGraphFilesDirectory() string {
	return _saveGraph
}

func CLIKmerSize() int {
	return _kmerSize
}
