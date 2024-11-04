package obitagpcr

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obimultiplex"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obipairing"
	"github.com/DavidGamba/go-getoptions"
)

var _reorientate = false

func TagPCROptionSet(options *getoptions.GetOpt) {
	options.BoolVar(&_reorientate, "reorientate", _reorientate,
		options.Description("Reverse complemente reads if needed to store all the sequences in "+
			"the same orientation respectively to forward and reverse primers"))

}

// OptionSet adds to the basic option set every options declared for
// the obiuniq command
func OptionSet(options *getoptions.GetOpt) {
	obipairing.OptionSet(options)
	obimultiplex.MultiplexOptionSet(options)
	TagPCROptionSet(options)
}

func CLIReorientate() bool {
	return _reorientate
}
