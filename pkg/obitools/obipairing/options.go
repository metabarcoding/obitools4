package obipairing

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var _ForwardFiles = make([]string, 0, 10)
var _ReverseFiles = make([]string, 0, 10)
var _Delta = 5
var _MinOverlap = 20
var _GapPenality = float64(2.0)
var _WithoutStats = false
var _MinIdentity = 0.9

func PairingOptionSet(options *getoptions.GetOpt) {
	options.StringSliceVar(&_ForwardFiles, "forward-reads",
		1, 1000,
		options.Alias("F"),
		options.Required("You must provide at least one forward file"),
		options.Description("The file names containing the forward reads"))
	options.StringSliceVar(&_ReverseFiles, "reverse-reads",
		1, 1000,
		options.Alias("R"),
		options.Required("You must provide at least one reverse file"),
		options.Description("The file names containing the reverse reads"))
	options.IntVar(&_Delta, "delta", _Delta,
		options.Alias("D"),
		options.Description("Length added to the fast detected overlap for the precise alignement"))
	options.IntVar(&_MinOverlap, "min-overlap", _MinOverlap,
		options.Alias("O"),
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
	obiconvert.OptionSet(options)
	PairingOptionSet(options)
}

func IBatchPairedSequence() (obiseq.IPairedBioSequenceBatch, error) {
	forward, err := obiconvert.ReadBioSequencesBatch(_ForwardFiles...)
	if err != nil {
		return obiseq.NilIPairedBioSequenceBatch, err
	}

	reverse, err := obiconvert.ReadBioSequencesBatch(_ReverseFiles...)
	if err != nil {
		return obiseq.NilIPairedBioSequenceBatch, err
	}

	paired := forward.PairWith(reverse)

	return paired, nil
}

func Delta() int {
	return _Delta
}

func MinOverlap() int {
	return _MinOverlap
}

func MinIdentity() float64 {
	return _MinIdentity
}

func GapPenality() float64 {
	return _GapPenality
}

func WithStats() bool {
	return !_WithoutStats
}
