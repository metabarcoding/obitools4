package obipairing

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var _ForwardFile = ""
var _ReverseFile = ""
var _Delta = 5
var _MinOverlap = 20
var _GapPenality = float64(2.0)
var _WithoutStats = false
var _MinIdentity = 0.9

func PairingOptionSet(options *getoptions.GetOpt) {
	options.StringVar(&_ForwardFile, "forward-reads", "",
		options.Alias("F"),
		options.ArgName("FILENAME_F"),
		options.Required("You must provide at a forward file"),
		options.Description("The file names containing the forward reads"))
	options.StringVar(&_ReverseFile, "reverse-reads", "",
		options.Alias("R"),
		options.ArgName("FILENAME_R"),
		options.Required("You must provide a reverse file"),
		options.Description("The file names containing the reverse reads"))
	options.IntVar(&_Delta, "delta", _Delta,
		options.Alias("D"),
		options.Description("Length added to the fast detected overlap for the precise alignement"))
	options.IntVar(&_MinOverlap, "min-overlap", _MinOverlap,
		options.Description("Minimum ovelap between both the reads to consider the aligment"))
	options.Float64Var(&_MinIdentity, "min-identity", _MinIdentity,
		options.Alias("X"),
		options.Description("Minimum identity between ovelaped regions of the reads to consider the aligment"))
	options.Float64Var(&_GapPenality, "gap-penality", _GapPenality,
		options.Alias("G"),
		options.Description("Gap penality expressed as the multiply factor applied to the mismatch score between two nucleotides with a quality of 40 (default 2)."))
	options.BoolVar(&_WithoutStats, "without-stat", _WithoutStats,
		options.Alias("S"),
		options.Description("Remove alignment statistics from the produced consensus sequences."))
}

func OptionSet(options *getoptions.GetOpt) {
	obiconvert.OutputOptionSet(options)
	obiconvert.InputOptionSet(options)
	PairingOptionSet(options)
}

func CLIPairedSequence() (obiiter.IBioSequence, error) {
	forward, err := obiconvert.CLIReadBioSequences(_ForwardFile)
	if err != nil {
		return obiiter.NilIBioSequence, err
	}

	reverse, err := obiconvert.CLIReadBioSequences(_ReverseFile)
	if err != nil {
		return obiiter.NilIBioSequence, err
	}

	paired := forward.PairTo(reverse)

	return paired, nil
}

func CLIDelta() int {
	return _Delta
}

func CLIMinOverlap() int {
	return _MinOverlap
}

func CLIMinIdentity() float64 {
	return _MinIdentity
}

func CLIGapPenality() float64 {
	return _GapPenality
}

func CLIWithStats() bool {
	return !_WithoutStats
}
