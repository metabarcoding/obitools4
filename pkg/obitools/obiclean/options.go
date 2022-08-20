package obiclean

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var _distStepMax = 1
var _sampleAttribute = "sample"

var _ratioMax = 1.0
var _minEvalRate = 1000

var _clusterMode = false
var _onlyHead = false

var _saveGraph = "__@@NOSAVE@@__"
var _saveRatio = "__@@NOSAVE@@__"

func ObicleanOptionSet(options *getoptions.GetOpt) {
	options.StringVar(&_sampleAttribute, "sample", _sampleAttribute,
		options.Alias("s"),
		options.Description("Attribute containing sample descriptions (default %s)."))

	options.IntVar(&_distStepMax, "distance", _distStepMax,
		options.Alias("d"),
		options.Description("Maximum numbers of differences between two variant sequences (default: %d)."))

	options.Float64Var(&_ratioMax, "ratio", _ratioMax,
		options.Alias("r"),
		options.Description("Threshold ratio between counts (rare/abundant counts)"+
			" of two sequence records so that the less abundant one is a variant of "+
			"the more abundant (default: %3.2f)."))

	options.IntVar(&_minEvalRate, "min-eval-rate", _minEvalRate,
		options.Description("Minimum abundance of a sequence to be used to evaluate mutation rate."))

	// options.BoolVar(&_clusterMode, "cluster", _clusterMode,
	// 	options.Alias("C"),
	// 	options.Description("Switch obiclean into its clustering mode. This adds information to each sequence about the true."),
	// )

	options.BoolVar(&_onlyHead, "head", _onlyHead,
		options.Alias("H"),
		options.Description("Select only sequences with the head status in a least one sample."),
	)

	options.StringVar(&_saveGraph, "save-graph", _saveGraph,
		options.Description("Creates a directory containing the set of DAG used by the obiclean clustering algorithm. "+
			"The graph files follow the graphml format."),
	)

	options.StringVar(&_saveRatio, "save-ratio", _saveRatio,
		options.Description("Creates a file containing the set of abundance ratio on the graph edge. "+
			"The ratio file follows the csv format."),
	)

}

func OptionSet(options *getoptions.GetOpt) {
	obiconvert.InputOptionSet(options)
	obiconvert.OutputOptionSet(options)
	ObicleanOptionSet(options)
}

func DistStepMax() int {
	return _distStepMax
}

// It returns the name of the attibute used to store sample name
func SampleAttribute() string {
	return _sampleAttribute
}

// Return the maximum abundance ratio between read counts to consider sequences as error.
func RatioMax() float64 {
	return _ratioMax
}

// > The function `MinCountToEvalMutationRate()` returns the minimum number of reads that must be
// observed before the mutation rate can be evaluated
func MinCountToEvalMutationRate() int {
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
func SaveGraphToFiles() bool {
	return _saveGraph != "__@@NOSAVE@@__"
}

// It returns the directory where the graph files are saved
func GraphFilesDirectory() string {
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
