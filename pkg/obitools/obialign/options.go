package obialign

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	sim "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obisimilarity"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
	log "github.com/sirupsen/logrus"
)

var _threshold = 0.0
var _refShortest = false
var _refLongest = false
var _refAlignment = false
var _normalized = false
var _raw = false
var _distance = false
var _acceptAmbiguous = false
var _referenceDB = ""
var _extraColumns = false

func ObialignOptionSet(options *getoptions.GetOpt) {
	options.Float64Var(&_threshold, "threshold", _threshold,
		options.Alias("t"),
		options.ArgName("THRESHOLD"),
		options.Description("Score threshold: only the pairs of sequences whose score reaches it are reported "+
			"(default: no threshold, all the pairs are reported). If the score is normalized and expressed in similarity (default), "+
			"it is an identity, e.g. 0.95 for an identity of 95%. If the score is normalized and expressed in distance, "+
			"it is (1.0 - identity), e.g. 0.05 for an identity of 95%. If the score is not normalized and expressed in similarity, "+
			"it is the length of the Longest Common Subsequence. If the score is not normalized and expressed in distance, "+
			"it is (reference length - LCS length)."))

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

	options.BoolVar(&_acceptAmbiguous, "accept-ambiguous", _acceptAmbiguous,
		options.Alias("g"),
		options.Description("Accept the sequences containing IUPAC letters other than a, c, g and t; "+
			"the alignment handles the IUPAC codes (default: these sequences are discarded)."))

	options.StringVar(&_referenceDB, "reference-db", _referenceDB,
		options.Alias("R"),
		options.ArgName("FILENAME"),
		options.Description("Compare every sequence of the input with every sequence of this file "+
			"(default: compare the sequences of the input with each other)."))

	options.BoolVar(&_extraColumns, "extra-columns", _extraColumns,
		options.Alias("x"),
		options.Description("Add the columns count1, count2, length1 and length2 to the table."))
}

func OptionSet(options *getoptions.GetOpt) {
	obiconvert.InputOptionSet(options)
	obiconvert.OutputModeOptionSet(options, true)
	ObialignOptionSet(options)
}

// CLIThreshold returns the score threshold, 0 if there is none.
func CLIThreshold() float64 { return _threshold }

// CLINormalized returns true when the score is normalized by the reference length.
func CLINormalized() bool {
	if _raw && _normalized {
		log.Fatal("The options --raw and --normalized are exclusive")
	}
	return !_raw
}

// CLIDistance returns true when the score is expressed in distance.
func CLIDistance() bool { return _distance }

// CLIReference returns the length used as reference to normalize the score.
func CLIReference() sim.Reference {
	n := 0
	ref := sim.RefAlignment
	if _refShortest {
		n++
		ref = sim.RefShortest
	}
	if _refLongest {
		n++
		ref = sim.RefLongest
	}
	if _refAlignment {
		n++
	}
	if n > 1 {
		log.Fatal("The options --shortest, --longest and --alignment are exclusive")
	}
	return ref
}

// CLIAcceptAmbiguous returns true when the sequences with IUPAC letters other than acgt are kept.
func CLIAcceptAmbiguous() bool { return _acceptAmbiguous }

// CLIExtraColumns returns true when the counts and the lengths must be written.
func CLIExtraColumns() bool { return _extraColumns }

// CLIHasReferenceDB returns true when a second dataset has been given.
func CLIHasReferenceDB() bool { return _referenceDB != "" }

// CLIReferenceDB reads the second dataset.
func CLIReferenceDB() obiseq.BioSequenceSlice {
	// Read as the main input is, with the same input options.
	iterator, err := obiconvert.CLIReadBioSequences(_referenceDB)
	if err != nil {
		log.Fatalf("Cannot open the reference file %s: %v", _referenceDB, err)
	}
	_, db := iterator.Load()
	return db
}
