package obirefidx

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

// OptionSet adds to the basic option set every options declared for
// the obiuniq command
func OptionSet(options *getoptions.GetOpt) {
	obiconvert.OptionSet(options)
	obioptions.LoadTaxonomyOptionSet(options, true, false)
}
