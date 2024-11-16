package obifind

import (
	"fmt"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
	"github.com/DavidGamba/go-getoptions"
)

var __rank_list__ = false
var __taxonomical_restriction__ = make([]string, 0)

var __fixed_pattern__ = false
var __with_path__ = false
var __taxid_path__ = "NA"
var __taxid_sons__ = "NA"
var __restrict_rank__ = ""

func FilterTaxonomyOptionSet(options *getoptions.GetOpt) {
	options.BoolVar(&__rank_list__, "rank-list", false,
		options.Alias("l"),
		options.Description("List every taxonomic rank available in the taxonomy."))

	options.StringSliceVar(&__taxonomical_restriction__, "restrict-to-taxon", 1, 1,
		options.Alias("r"),
		options.Description("Restrict output to some subclades."))
}

func CLITaxonomicalRestrictions() (*obitax.TaxonSet, error) {
	taxonomy := obitax.DefaultTaxonomy()

	if taxonomy == nil {
		return nil, fmt.Errorf("no taxonomy loaded")
	}

	ts := taxonomy.NewTaxonSet()
	for _, taxid := range __taxonomical_restriction__ {
		tx := taxonomy.Taxon(taxid)

		if tx == nil {
			return nil, fmt.Errorf(
				"cannot find taxon %s in taxonomy %s",
				taxid,
				taxonomy.Name(),
			)
		}

		ts.InsertTaxon(tx)
	}

	return ts, nil
}

func OptionSet(options *getoptions.GetOpt) {
	obioptions.LoadTaxonomyOptionSet(options, true, true)
	FilterTaxonomyOptionSet(options)
	options.BoolVar(&__fixed_pattern__, "fixed", false,
		options.Alias("F"),
		options.Description("Match taxon names using a fixed pattern, not a regular expression"))
	options.BoolVar(&__with_path__, "with-path", false,
		options.Alias("P"),
		options.Description("Adds a column containing the full path for each displayed taxon."))
	options.StringVar(&__taxid_path__, "parents", "NA",
		options.Alias("p"),
		options.Description("Displays every parental tree's information for the provided taxid."))
	options.StringVar(&__restrict_rank__, "rank", "",
		options.Description("Restrict to the given taxonomic rank."))
}

func CLIRequestsPathForTaxid() string {
	return __taxid_path__
}

func CLIRequestsSonsForTaxid() string {
	return __taxid_sons__
}
