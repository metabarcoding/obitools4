package obikmersim

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiformats"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var _KmerSize = 30
var _Sparse = false
var _References = []string{}
var _MinSharedKmers = 1
var _Self = false

var _Delta = 5
var _PenalityScale = 1.0
var _GapPenality = 2.0
var _FastScoreAbs = false

// PCROptionSet defines every options related to a simulated PCR.
//
// The function adds to a CLI every options proposed to the user
// to tune the parametters of the PCR simulation algorithm.
//
// # Parameters
//
// - option : is a pointer to a getoptions.GetOpt instance normaly
// produced by the
func KmerSimCountOptionSet(options *getoptions.GetOpt) {

	options.IntVar(&_KmerSize, "kmer-size", _KmerSize,
		options.Alias("k"),
		options.Description("Kmer size to use."))

	options.BoolVar(&_Sparse, "sparse", _Sparse,
		options.Alias("S"),
		options.Description("Set sparse kmer mode."))

	options.StringSliceVar(&_References, "reference", 1, 1,
		options.Alias("r"),
		options.Description("Reference sequence."))

	options.IntVar(&_MinSharedKmers, "min-shared-kmers", _MinSharedKmers,
		options.Alias("m"),
		options.Description("Minimum number of shared kmers between two sequences."))

	options.BoolVar(&_Self, "self", _Self,
		options.Alias("s"),
		options.Description("Compare references with themselves."))

}

func KmerSimMatchOptionSet(options *getoptions.GetOpt) {
	options.IntVar(&_Delta, "delta", _Delta,
		options.Alias("d"),
		options.Description("Delta value for the match."))

	options.Float64Var(&_PenalityScale, "penality-scale", _PenalityScale,
		options.Alias("X"),
		options.Description("Scale factor applied to the mismatch score and the gap penality (default 1)."))

	options.Float64Var(&_GapPenality, "gap-penality", _GapPenality,
		options.Alias("G"),
		options.Description("Gap penality expressed as the multiply factor applied to the mismatch score between two nucleotides with a quality of 40 (default 2)."))

	options.BoolVar(&_FastScoreAbs, "fast-absolute", _FastScoreAbs,
		options.Alias("a"),
		options.Description("Use fast absolute score mode."))
}

func CountOptionSet(options *getoptions.GetOpt) {
	obiconvert.OptionSet(options)
	KmerSimCountOptionSet(options)
}

func MatchOptionSet(options *getoptions.GetOpt) {
	obiconvert.OptionSet(options)
	KmerSimCountOptionSet(options)
	KmerSimMatchOptionSet(options)
}

func CLIKmerSize() uint {
	return uint(_KmerSize)
}

func CLISparseMode() bool {
	return _Sparse
}

func CLIReference() (string, obiseq.BioSequenceSlice) {

	refnames, err := obiconvert.ExpandListOfFiles(false, _References...)

	if err != nil {
		return "", obiseq.BioSequenceSlice{}
	}

	nreader := 1

	if obiconvert.CLINoInputOrder() {
		nreader = obioptions.StrictReadWorker()
	}

	source, references := obiformats.ReadSequencesBatchFromFiles(
		refnames,
		obiformats.ReadSequencesFromFile,
		nreader).Load()

	return source, references
}

func CLIMinSharedKmers() int {
	return _MinSharedKmers
}

func CLISelf() bool {
	return _Self
}

func CLIDelta() int {
	return _Delta
}

func CLIScale() float64 {
	return _PenalityScale
}

func CLIGapPenality() float64 {
	return _GapPenality
}

func CLIGap() float64 {
	return _GapPenality
}

func CLIFastRelativeScore() bool {
	return !_FastScoreAbs
}
