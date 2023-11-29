package obicleandb

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obigrep"
	"github.com/DavidGamba/go-getoptions"
)

var _UpdateTaxids = false

func ObicleanDBOptionSet(options *getoptions.GetOpt) {

	options.BoolVar(&_UpdateTaxids, "update-taxids", _UpdateTaxids,
		options.Description("Indicates if decrecated Taxids must be corrected."))
}

func OptionSet(options *getoptions.GetOpt) {
	obiconvert.InputOptionSet(options)
	obiconvert.OutputOptionSet(options)
	obigrep.TaxonomySelectionOptionSet(options)
	ObicleanDBOptionSet(options)
}

func CLIUpdateTaxids() bool {
	return _UpdateTaxids
}
