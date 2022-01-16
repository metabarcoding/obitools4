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

func PairingOptionSet(options *getoptions.GetOpt) {
	options.StringSliceVar(&_ForwardFiles, "forward-reads",
		1, 1000,
		options.Alias("F"),
		options.Description("The file names containing the forward reads"))
	options.StringSliceVar(&_ReverseFiles, "reverse-reads",
		1, 1000,
		options.Alias("R"),
		options.Description("The file names containing the reverse reads"))
	options.IntVar(&_Delta, "delta", 5,
		options.Alias("D"),
		options.Description("Length added to the fast detected overlap for the precise alignement (default 5)."))
	options.IntVar(&_MinOverlap, "min-overlap", 20,
		options.Alias("O"),
		options.Description("Minimum ovelap between both the reads to consider the aligment (default 20)."))
	options.Float64Var(&_GapPenality, "gap-penality", 2,
		options.Alias("G"),
		options.Description("Gap penality expressed as the multiply factor applied to the mismatch score between two nucleotides with a quality of 40 (default 2)."))
	options.BoolVar(&_WithoutStats, "without-stat", false,
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

func GapPenality() float64 {
	return _GapPenality
}

func WithStats() bool {
	return !_WithoutStats
}
