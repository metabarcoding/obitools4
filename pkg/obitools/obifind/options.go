package obifind

import (
	"errors"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiformats/ncbitaxdump"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitax"
	"github.com/DavidGamba/go-getoptions"
)

var __taxdump__ = ""
var __alternative_name__ = false
var __rank_list__ = false
var __selected_taxonomy__ = (*obitax.Taxonomy)(nil)
var __taxonomical_restriction__ = make([]int, 0)

var __fixed_pattern__ = false
var __with_path__ = false
var __taxid_path__ = -1
var __taxid_sons__ = -1
var __restrict_rank__ = ""

func LoadTaxonomyOptionSet(options *getoptions.GetOpt, required, alternatiive bool) {
	if required {
		options.StringVar(&__taxdump__, "taxdump", "",
			options.Alias("t"),
			options.Required(),
			options.Description("Points to the directory containing the NCBI Taxonomy database dump."))
	} else {
		options.StringVar(&__taxdump__, "taxdump", "",
			options.Alias("t"),
			options.Description("Points to the directory containing the NCBI Taxonomy database dump."))
	}
	if alternatiive {
		options.BoolVar(&__alternative_name__, "alternative-names", false,
			options.Alias("a"),
			options.Description("Enable the search on all alternative names and not only scientific names."))
	}
	options.BoolVar(&__rank_list__, "rank-list", false,
		options.Alias("l"),
		options.Description("List every taxonomic rank available in the taxonomy."))

	options.IntSliceVar(&__taxonomical_restriction__, "restrict-to-taxon", 1, 1,
		options.Alias("r"),
		options.Description("Restrict output to some subclades."))
}

func CLISelectedNCBITaxDump() string {
	return __taxdump__
}

func CLIAreAlternativeNamesSelected() bool {
	return __alternative_name__
}

func CLITaxonomicalRestrictions() (*obitax.TaxonSet, error) {
	taxonomy, err := CLILoadSelectedTaxonomy()

	if err != nil {
		return nil, err
	}

	ts := make(obitax.TaxonSet)
	for _, taxid := range __taxonomical_restriction__ {
		tx, err := taxonomy.Taxon(taxid)

		if err != nil {
			return nil, err
		}

		ts.Inserts(tx)
	}

	return &ts, nil
}

func CLILoadSelectedTaxonomy() (*obitax.Taxonomy, error) {
	if CLISelectedNCBITaxDump() != "" {
		if __selected_taxonomy__ == nil {
			var err error
			__selected_taxonomy__, err = ncbitaxdump.LoadNCBITaxDump(CLISelectedNCBITaxDump(),
				!CLIAreAlternativeNamesSelected())
			if err != nil {
				return nil, err
			}
		}
		return __selected_taxonomy__, nil
	}

	return nil, errors.New("no NCBI taxdump selected using option -t|--taxdump")
}

func OptionSet(options *getoptions.GetOpt) {
	LoadTaxonomyOptionSet(options, true, true)
	options.BoolVar(&__fixed_pattern__, "fixed", false,
		options.Alias("F"),
		options.Description("Match taxon names using a fixed pattern, not a regular expression"))
	options.BoolVar(&__with_path__, "with-path", false,
		options.Alias("P"),
		options.Description("Adds a column containing the full path for each displayed taxon."))
	options.IntVar(&__taxid_path__, "parents", -1,
		options.Alias("p"),
		options.Description("Displays every parental tree's information for the provided taxid."))
	options.StringVar(&__restrict_rank__, "rank", "",
		options.Description("Restrict to the given taxonomic rank."))
}

func CLIRequestsPathForTaxid() int {
	return __taxid_path__
}

func CLIRequestsSonsForTaxid() int {
	return __taxid_sons__
}
