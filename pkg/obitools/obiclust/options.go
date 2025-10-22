package obiclust

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

type ScoreNormalizationMode int

const (
	NoNormalization ScoreNormalizationMode = iota
	NormalizedByShortest
	NormalizedByLongest
	NormalizedByAlignment
)

var _sampleAttribute = "sample"
var _threshold = 3.0
var _shortReference = false
var _longReference = false
var _alignReference = true
var _normalizedScore = false
var _distanceMode = false
var _exactMode = false
var _ratioMax = 1.0
var _sorted_by_length = false
var _sorted_by_abundance = true
var _sortingAscending = false
var _onlyHead = false
var _saveGraph = "__@@NOSAVE@@__"
var _saveRatio = "__@@NOSAVE@@__"
var _minSample = 1

func ObiclustOptionSet(options *getoptions.GetOpt) {
	options.StringVar(&_sampleAttribute, "sample", _sampleAttribute,
		options.Alias("s"),
		options.Description("Attribute containing sample descriptions (default %s)."))

	options.Float64Var(&_threshold, "distance", _threshold,
		options.Alias("--threshold"),
		options.Description("Threshold to merge sequences into the same cluster (default: %d)."))

	options.Float64Var(&_ratioMax, "ratio", _ratioMax,
		options.Alias("r"),
		options.Description("Threshold ratio between counts (rare/abundant counts)"+
			" of two sequence records so that the less abundant one is a variant of "+
			"the more abundant (default: %3.2f)."))

	options.BoolVar(&_onlyHead, "head", _onlyHead,
		options.Alias("H"),
		options.Description("Select only sequences with the head status in at least one sample."),
	)

	options.StringVar(&_saveGraph, "save-graph", _saveGraph,
		options.Description("Creates a directory containing the set of DAG used by the obiclean clustering algorithm. "+
			"The graph files follow the graphml format."),
	)

	options.StringVar(&_saveRatio, "save-ratio", _saveRatio,
		options.Description("Creates a file containing the set of abundance ratio on the graph edge. "+
			"The ratio file follows the csv format."),
	)

	options.IntVar(&_minSample, "min-sample-count", _minSample,
		options.Description("Minimum number of samples a sequence must be present in to be considered in the analysis."),
	)

	options.BoolVar(&_normalizedScore, "normalized-score", _normalizedScore,
		options.Alias("n"),
		options.Description("Use alignment score normalized by length"),
	)

	options.BoolVar(&_shortReference, "shortest", _shortReference,
		options.Description("Use length of the shortest sequence to normalize alignment score."),
	)

	options.BoolVar(&_longReference, "longest", _longReference,
		options.Description("Use length of the longest sequence to normalize alignment score."),
	)

	options.BoolVar(&_alignReference, "alignment", _alignReference,
		options.Description("Use alignment length to normalize alignment score."),
	)

	options.BoolVar(&_distanceMode, "distance", _distanceMode,
		options.Description("Use alignment distance instead of similarity score."),
	)

	options.BoolVar(&_exactMode, "exact", _exactMode,
		options.Description("Use exact clustering algorithm (default is greedy)."),
	)

	options.BoolVar(&_sorted_by_length, "length-ordered", _sorted_by_length,
		options.Description("Sort sequence by length before clustering."),
	)
	options.BoolVar(&_sorted_by_abundance, "abundance-ordered", _sorted_by_abundance,
		options.Description("Sort sequence by read counts (abundance) before clustering."),
	)
	options.BoolVar(&_sortingAscending, "ascending-sorting", _sortingAscending,
		options.Description("Sort order is ascending (default is descending)."),
	)
}

func OptionSet(options *getoptions.GetOpt) {
	obiconvert.InputOptionSet(options)
	obiconvert.OutputOptionSet(options)
	ObiclustOptionSet(options)
}

// It returns the name of the attibute used to store sample name
func CLISampleAttribute() string {
	return _sampleAttribute
}

func CLIThreshold() float64 {
	return _threshold
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
func CLIIsSaveRatioTable() bool {
	return _saveRatio != "__@@NOSAVE@@__"
}

// It returns the filename of the file that stores the ratio table
func CLIRatioTableFilename() string {
	return _saveRatio
}

func CLINormalizationMode() ScoreNormalizationMode {
	switch {
	case _alignReference:
		return NormalizedByAlignment
	case _longReference:
		return NormalizedByLongest
	case _shortReference:
		return NormalizedByShortest
	default:
		return NoNormalization
	}
}
