package obiminion

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var _distStepMax = 1
var _sampleAttribute = "sample"

var _ratioMax = 1.0

var _clusterMode = false
var _onlyHead = false

var _kmerSize = -1

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

// CLINoSingleton returns a boolean value indicating whether or not singleton sequences should be discarded.
//
// No parameters.
// Returns a boolean value indicating whether or not singleton sequences should be discarded.
func CLINoSingleton() bool {
	return _NoSingleton
}