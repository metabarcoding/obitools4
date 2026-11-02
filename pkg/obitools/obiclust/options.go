package obiclust

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
	log "github.com/sirupsen/logrus"
)

// reference is the length used to normalize a score (or to express a raw
// distance): the alignment length, the longest or the shortest sequence.
type reference int

const (
	refAlignment reference = iota
	refLongest
	refShortest
)

var _threshold = 0.97
var _ratio = 1.0
var _exact = false
var _distance = false
var _raw = false
var _normalized = false
var _refShortest = false
var _refLongest = false
var _refAlignment = false
var _sortKey = "count"
var _ascending = false
var _acceptAmbiguous = false
var _otuMap = ""
var _noSequenceOutput = false

func ObiclustOptionSet(options *getoptions.GetOpt) {
	options.Float64Var(&_threshold, "threshold", _threshold,
		options.Alias("t"),
		options.ArgName("THRESHOLD"),
		options.Description("Score threshold for clustering. If the score is normalized and expressed in similarity (default), "+
			"it is an identity, e.g. 0.95 for an identity of 95%. If the score is normalized and expressed in distance, "+
			"it is (1.0 - identity), e.g. 0.05 for an identity of 95%. If the score is not normalized and expressed in similarity, "+
			"it is the length of the Longest Common Subsequence. If the score is not normalized and expressed in distance, "+
			"it is (reference length - LCS length)."),
	)

	options.BoolVar(&_refShortest, "shortest", _refShortest,
		options.Alias("l"),
		options.Description("Reference sequence length is the shortest."))

	options.BoolVar(&_refLongest, "longest", _refLongest,
		options.Alias("L"),
		options.Description("Reference sequence length is the longest."))

	options.BoolVar(&_refAlignment, "alignment", _refAlignment,
		options.Alias("a"),
		options.Description("Reference sequence length is the alignment length (default)."))

	options.BoolVar(&_normalized, "normalized", _normalized,
		options.Alias("n"),
		options.Description("Score is normalized by the reference sequence length (default)."))

	options.BoolVar(&_raw, "raw", _raw,
		options.Alias("r"),
		options.Description("Raw score, not normalized."))

	options.BoolVar(&_distance, "distance", _distance,
		options.Alias("d"),
		options.Description("Score is expressed in distance (default: score is expressed in similarity)."))

	options.BoolVar(&_exact, "exact", _exact,
		options.Alias("e"),
		options.Description("A sequence is assigned to the cluster whose center presents the highest score "+
			"above the threshold, as opposed to the default 'fast' mode where it is assigned to the first "+
			"cluster found whose center presents a score above the threshold."))

	options.Float64Var(&_ratio, "ratio", _ratio,
		options.Alias("R"),
		options.ArgName("RATIO"),
		options.Description("Maximum ratio between the counts of two sequences so that the less abundant one "+
			"can be considered as a variant of the more abundant one (default: %3.2f)."))

	options.StringVar(&_sortKey, "sort", _sortKey,
		options.Alias("s"),
		options.ArgName("KEY"),
		options.Description("Sort the sequences on KEY before clustering. KEY must be 'None' for no sorting, "+
			"or an attribute of the sequences, except for the count that can be computed (default: sorting on count). "+
			"The sequences are sorted in decreasing order (see --ascending), ties being ordered "+
			"as in the input."))

	options.BoolVar(&_ascending, "ascending", _ascending,
		options.Description("Sorting is in ascending order (default: descending)."))

	options.BoolVar(&_acceptAmbiguous, "accept-ambiguous", _acceptAmbiguous,
		options.Alias("g"),
		options.Description("Accept the sequences containing IUPAC letters other than a, c, g and t. "+
			"Such a sequence can be attached to a cluster (the alignment handles the IUPAC codes), "+
			"but cannot be the center of a cluster, and is left without cluster if it matches no center "+
			"(default: these sequences are discarded)."))

	options.StringVar(&_otuMap, "otu-map", _otuMap,
		options.ArgName("FILENAME"),
		options.Description("Write the clusters to FILENAME as a CSV table, one row per sequence "+
			"(id, cluster, cluster_center, cluster_score, count), grouped by cluster."))

	options.BoolVar(&_noSequenceOutput, "no-sequence-output", _noSequenceOutput,
		options.Alias("f"),
		options.Description("Do not write the sequences on the standard output."))
}

func OptionSet(options *getoptions.GetOpt) {
	obiconvert.InputOptionSet(options)
	obiconvert.OutputOptionSet(options)
	ObiclustOptionSet(options)
}

// CLIThreshold returns the clustering threshold as given on the command line.
func CLIThreshold() float64 { return _threshold }

// CLIRatio returns the maximum abundance ratio between a sequence and the center it is attached to.
func CLIRatio() float64 { return _ratio }

// CLIExact returns true when a sequence must be attached to the best center and not to the first matching one.
func CLIExact() bool { return _exact }

// CLIDistance returns true when the score is expressed in distance.
func CLIDistance() bool { return _distance }

// CLINormalized returns true when the score is normalized by the reference length.
func CLINormalized() bool {
	if _raw && _normalized {
		log.Fatal("The options --raw and --normalized are exclusive")
	}
	return !_raw
}

// CLIReference returns the length used as reference to normalize the score.
func CLIReference() reference {
	n := 0
	ref := refAlignment
	if _refShortest {
		n++
		ref = refShortest
	}
	if _refLongest {
		n++
		ref = refLongest
	}
	if _refAlignment {
		n++
	}
	if n > 1 {
		log.Fatal("The options --shortest, --longest and --alignment are exclusive")
	}
	return ref
}

// CLISortKey returns the key used to sort the sequences, or "None".
func CLISortKey() string {
	if _sortKey == "none" {
		return "None"
	}
	return _sortKey
}

// CLIAscending returns true when the sort must be in ascending order.
func CLIAscending() bool { return _ascending }

// CLIAcceptAmbiguous returns true when the sequences with IUPAC letters other than acgt are kept.
func CLIAcceptAmbiguous() bool { return _acceptAmbiguous }

// CLIOTUMap returns the name of the OTU map file, empty if none is requested.
func CLIOTUMap() string { return _otuMap }

// CLINoSequenceOutput returns true when the sequences must not be written.
func CLINoSequenceOutput() bool { return _noSequenceOutput }
