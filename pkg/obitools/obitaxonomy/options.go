package obitaxonomy

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
var __with_query__ = false
var __without_rank__ = false
var __without_parent__ = false
var __without_scientific_name__ = false
var __raw_taxid__ = false
var __taxid_path__ = "NA"
var __taxid_sons__ = "NA"
var __restrict_rank__ = ""
var __to_dump__ = ""
var __download_ncbi__ = false

func FilterTaxonomyOptionSet(options *getoptions.GetOpt) {
	options.BoolVar(&__rank_list__, "rank-list", false,
		options.Alias("l"),
		options.Description("List every taxonomic rank available in the taxonomy."))

	options.StringSliceVar(&__taxonomical_restriction__, "restrict-to-taxon", 1, 1,
		options.Alias("r"),
		options.Description("Restrict output to some subclades."))
}

func OptionSet(options *getoptions.GetOpt) {
	obioptions.LoadTaxonomyOptionSet(options, false, true)
	FilterTaxonomyOptionSet(options)
	options.BoolVar(&__fixed_pattern__, "fixed", false,
		options.Alias("F"),
		options.Description("Match taxon names using a fixed pattern, not a regular expression"))
	options.StringVar(&__taxid_path__, "parents", "NA",
		options.Alias("p"),
		options.ArgName("TAXID"),
		options.Description("Displays every parental tree's information for the provided taxid."))
	options.StringVar(&__restrict_rank__, "rank", "",
		options.ArgName("RANK"),
		options.Description("Restrict to the given taxonomic rank."))
	options.BoolVar(&__without_parent__, "without-parent", __without_parent__,
		options.Description("Supress the column containing the parent's taxonid from the output."))
	options.StringVar(&__taxid_sons__, "sons", "NA",
		options.Alias("s"),
		options.ArgName("TAXID"),
		options.Description("Displays every sons' tree's information for the provided taxid."))
	options.BoolVar(&__with_path__, "with-path", false,
		options.Description("Adds a column containing the full path for each displayed taxon."))
	options.BoolVar(&__without_rank__, "without-rank", __without_rank__,
		options.Alias("R"),
		options.Description("Supress the column containing the taxonomic rank from the output."))
	options.BoolVar(&__with_query__, "with-query", false,
		options.Alias("P"),
		options.Description("Adds a column containing query used to filter taxon name for each displayed taxon."))
	options.BoolVar(&__without_scientific_name__, "without-scientific-name", __without_scientific_name__,
		options.Alias("S"),
		options.Description("Supress the column containing the scientific name from the output."))
	options.BoolVar(&__raw_taxid__, "raw-taxid", false,
		options.Description("Displays the raw taxid for each displayed taxon."))
	options.StringVar(&__to_dump__, "dump", __to_dump__,
		options.Alias("D"),
		options.ArgName("TAXID"),
		options.Description("Dump a sub-taxonomy corresponding to the precised clade"),
	)
	options.BoolVar(&__download_ncbi__, "download-ncbi", __download_ncbi__,
		options.Description("Download the current NCBI taxonomy taxdump"),
	)

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

func CLIRequestsPathForTaxid() string {
	return __taxid_path__
}

func CLIRequestsSonsForTaxid() string {
	return __taxid_sons__
}

func CLIWithParent() bool {
	return !__without_parent__
}

func CLIWithPath() bool {
	return __with_path__
}

func CLIWithRank() bool {
	return !__without_rank__
}

func CLIWithScientificName() bool {
	return !__without_scientific_name__
}

func CLIRawTaxid() bool {
	return __raw_taxid__
}

func CLIRankRestriction() string {
	return __restrict_rank__
}

func CLIFixedPattern() bool {
	return __fixed_pattern__
}

func CLIWithQuery() bool {
	return __with_query__
}

func CLIDumpSubtaxonomy() bool {
	return __to_dump__ != ""
}

func CLISubTaxonomyNode() string {
	return __to_dump__
}

func CLIDownloadNCBI() bool {
	return __download_ncbi__
}
