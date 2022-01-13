package obipairing

import (
	"git.metabarcoding.org/lecasofts/go/oa2/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/oa2/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var __forward_files__ = make([]string, 0, 10)
var __reverse_files__ = make([]string, 0, 10)
var __delta__ = 5
var __min_overlap__ = 20
var __gap_penality__ = 2
var __without_stats__ = false

func PairingOptionSet(options *getoptions.GetOpt) {
	options.StringSliceVar(&__forward_files__, "forward-reads",
		1, 1000,
		options.Alias("F"),
		options.Description("The file names containing the forward reads"))
	options.StringSliceVar(&__reverse_files__, "reverse-reads",
		1, 1000,
		options.Alias("R"),
		options.Description("The file names containing the reverse reads"))
	options.IntVar(&__delta__, "delta", 5,
		options.Alias("D"),
		options.Description("Length added to the fast detected overlap for the precise alignement (default 5)."))
	options.IntVar(&__min_overlap__, "min-overlap", 20,
		options.Alias("O"),
		options.Description("Minimum ovelap between both the reads to consider the aligment (default 20)."))
	options.IntVar(&__gap_penality__, "gap-penality", 2,
		options.Alias("G"),
		options.Description("Gap penality expressed as the multiply factor applied to the mismatch score between two nucleotides with a quality of 40 (default 2)."))
	options.BoolVar(&__without_stats__, "without-stat", false,
		options.Alias("S"),
		options.Description("Remove alignment statistics from the produced consensus sequences."))
}

func OptionSet(options *getoptions.GetOpt) {
	obiconvert.OptionSet(options)
	PairingOptionSet(options)
}

func IBatchPairedSequence() (obiseq.IPairedBioSequenceBatch, error) {
	forward, err := obiconvert.ReadBioSequencesBatch(__forward_files__...)
	if err != nil {
		return obiseq.NilIPairedBioSequenceBatch, err
	}

	reverse, err := obiconvert.ReadBioSequencesBatch(__reverse_files__...)
	if err != nil {
		return obiseq.NilIPairedBioSequenceBatch, err
	}

	paired := forward.PairWith(reverse)

	return paired, nil
}

func Delta() int {
	return __delta__
}

func MinOverlap() int {
	return __min_overlap__
}

func GapPenality() int {
	return __gap_penality__
}

func WithStats() bool {
	return !__without_stats__
}
