package obirefidx

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obifind"
	"github.com/DavidGamba/go-getoptions"
)

// OptionSet adds to the basic option set every options declared for
// the obiuniq command
func OptionSet(options *getoptions.GetOpt) {
	obiconvert.OptionSet(options)
	obifind.LoadTaxonomyOptionSet(options, true, false)
}
