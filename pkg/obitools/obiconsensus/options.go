package obiconsensus

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var _saveGraph = "__@@NOSAVE@@__"
var _kmerSize = -1
var _threshold = 0.99
var _mindepth = -1.0

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

	options.Float64Var(&_threshold, "threshold", _threshold,
		options.ArgName("RATIO"),
		options.Description("A threshold between O and 1 used to determine the optimal "+
			"kmer size"),
	)

	options.Float64Var(&_mindepth, "min-depth", _mindepth,
		options.ArgName("DEPTH"),
		options.Description("if DEPTH is between 0 and 1, it corresponds to fraction of the "+
			"reads in which a kmer must occurs to be conserved in the graph. If DEPTH is greater "+
			"than 1, indicate the minimum count of occurrence for a kmer to be kept. "+
			"Default value = -1, which means that the DEPTH is estimated from the data"),
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

func CLIKmerDepth() float64 {
	return _mindepth
}

func CLIThreshold() float64 {
	return _threshold
}
