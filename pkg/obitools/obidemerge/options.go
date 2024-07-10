package obidemerge

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var _Demerge = ""

func DemergeOptionSet(options *getoptions.GetOpt) {

	options.StringVar(&_Demerge, "demerge", _Demerge,
		options.Alias("d"),
		options.Description("Indicates which slot has to be demerged."))
}

// OptionSet adds to the basic option set every options declared for
// the obipcr command
func OptionSet(options *getoptions.GetOpt) {
	obiconvert.OptionSet(options)
	DemergeOptionSet(options)
}

func CLIDemergeSlot() string {
	return _Demerge
}

func CLIHasSlotToDemerge() bool {
	return _Demerge != ""
}
