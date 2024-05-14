package obiminion

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var _distStepMax = 1
var _sampleAttribute = "sample"

var _ratioMax = 1.0
var _minEvalRate = 1000

var _clusterMode = false
var _onlyHead = false

var _kmerSize = -1
var _threshold = 1.0
var _mindepth = -1.0

var _consensus_max_length = 1000

var _NoSingleton = false

var _saveGraph = "__@@NOSAVE@@__"
var _saveRatio = "__@@NOSAVE@@__"

// ObiminionOptionSet sets the options for obiminion.
//
// options: The options for configuring obiminion.
func ObiminionOptionSet(options *getoptions.GetOpt) {
	options.StringVar(&_sampleAttribute, "sample", _sampleAttribute,
		options.Alias("s"),
		options.Description("Attribute containing sample descriptions (default %s)."))

	options.IntVar(&_distStepMax, "distance", _distStepMax,
		options.Alias("d"),
		options.Description("Maximum numbers of differences between two variant sequences (default: %d)."))

	options.IntVar(&_minEvalRate, "min-eval-rate", _minEvalRate,
		options.Description("Minimum abundance of a sequence to be used to evaluate mutation rate."))

	options.StringVar(&_saveGraph, "save-graph", _saveGraph,
		options.Description("Creates a directory containing the set of DAG used by the obiclean clustering algorithm. "+
			"The graph files follow the graphml format."),
	)

	options.StringVar(&_saveRatio, "save-ratio", _saveRatio,
		options.Description("Creates a file containing the set of abundance ratio on the graph edge. "+
			"The ratio file follows the csv format."),
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

	options.IntVar(&_consensus_max_length, "consensus-max-length", _consensus_max_length,
		options.ArgName("LENGTH"),
		options.Description("Maximum length of the consensus sequence. "+
			"Default value = -1, which means that no limit is applied"),
	)

	options.BoolVar(&_NoSingleton, "no-singleton", _NoSingleton,
		options.Description("If set, sequences occurring a single time in the data set are discarded."))


}

// OptionSet sets up the options for the obiminion package.
//
// It takes a pointer to a getoptions.GetOpt object as a parameter.
// It does not return any value.
func OptionSet(options *getoptions.GetOpt) {
	obiconvert.InputOptionSet(options)
	obiconvert.OutputOptionSet(options)
	ObiminionOptionSet(options)
}

// CLIDistStepMax returns the maximum distance between two sequences.
//
// The value of the distance is set by the user with the `-d` flag.
//
// No parameters.
// Returns an integer.
func CLIDistStepMax() int {
	return _distStepMax
}

// CLISampleAttribute returns the name of the attribute used to store sample name.
//
// The value of the sample attribute is set by the user with the `-s` flag.
//
// No parameters.
// Returns a string.
func CLISampleAttribute() string {
	return _sampleAttribute
}

// > The function `CLIMinCountToEvalMutationRate()` returns the minimum number of reads that must be
// observed before the mutation rate can be evaluated
func CLIMinCountToEvalMutationRate() int {
	return _minEvalRate
}

func ClusterMode() bool {
	return _clusterMode
}

// `OnlyHead()` returns a boolean value that indicates whether the `-h` flag was passed to the program
func OnlyHead() bool {
	return _onlyHead
}

// Returns true it the obliclean graphs must be saved
func CLISaveGraphToFiles() bool {
	return _saveGraph != "__@@NOSAVE@@__"
}

// It returns the directory where the graph files are saved
func CLIGraphFilesDirectory() string {
	return _saveGraph
}

// Returns true it the table of ratio must be saved
func IsSaveRatioTable() bool {
	return _saveRatio != "__@@NOSAVE@@__"
}

// It returns the filename of the file that stores the ratio table
func RatioTableFilename() string {
	return _saveRatio
}

// CLIKmerSize returns the value of the kmer size to use for building the consensus.
//
// The value of the kmer size is set by the user with the `-k` flag.
// The value -1 means that the kmer size is estimated as the minimum value that
// insure that no kmer are present more than one time in a sequence.
//
// No parameters.
// Returns an integer value.
func CLIKmerSize() int {
	return _kmerSize
}

func CLIKmerDepth() float64 {
	return _mindepth
}

func CLIThreshold() float64 {
	return _threshold
}

func CLIMaxConsensusLength() int {
	return _consensus_max_length
}

// CLINoSingleton returns a boolean value indicating whether or not singleton sequences should be discarded.
//
// No parameters.
// Returns a boolean value indicating whether or not singleton sequences should be discarded.
func CLINoSingleton() bool {
	return _NoSingleton
}