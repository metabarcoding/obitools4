package obicleandb

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obigrep"
	"github.com/DavidGamba/go-getoptions"
)

func OptionSet(options *getoptions.GetOpt) {
	obioptions.LoadTaxonomyOptionSet(options, true, false)
	obiconvert.InputOptionSet(options)
	obiconvert.OutputOptionSet(options)
	obigrep.TaxonomySelectionOptionSet(options)
}
