package main

import (
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obitaxonomy"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"

	log "github.com/sirupsen/logrus"
)

func main() {
	optionParser := obioptions.GenerateOptionParser(obitaxonomy.OptionSet)

	_, args := optionParser(os.Args)

	var iterator *obitax.ITaxon

	switch {

	case obitaxonomy.CLIDownloadNCBI():
		err := obitaxonomy.CLIDownloadNCBITaxdump()
		if err != nil {
			log.Errorf("Cannot download NCBI taxonomy: %s", err.Error())
			os.Exit(1)
		}

		os.Exit(0)

	case obitaxonomy.CLIDumpSubtaxonomy():
		iterator = obitaxonomy.CLISubTaxonomyIterator()

	case obitaxonomy.CLIRequestsPathForTaxid() != "NA":

		taxon := obitax.DefaultTaxonomy().Taxon(obitaxonomy.CLIRequestsPathForTaxid())

		if taxon == nil {
			log.Fatalf("Cannot identify the requested taxon: %s",
				obitaxonomy.CLIRequestsPathForTaxid())
		}

		s := taxon.Path()

		if s == nil {
			log.Fatalf("Cannot extract taxonomic path describing %s", taxon.String())
		}

		iterator = s.Iterator()

		if obitaxonomy.CLIWithQuery() {
			iterator = iterator.AddMetadata("query", taxon.String())
		}

	case len(args) == 0:
		iterator = obitax.DefaultTaxonomy().Iterator()
	default:
		iters := make([]*obitax.ITaxon, len(args))

		for i, pat := range args {
			ii := obitax.DefaultTaxonomy().IFilterOnName(pat, obitaxonomy.CLIFixedPattern(), true)
			if obitaxonomy.CLIWithQuery() {
				ii = ii.AddMetadata("query", pat)
			}
			iters[i] = ii
		}

		iterator = iters[0]

		if len(iters) > 1 {
			iterator = iterator.Concat(iters[1:]...)
		}
	}

	iterator = obitaxonomy.CLITaxonRestrictions(iterator)
	obitaxonomy.CLICSVTaxaWriter(iterator, true)

	obiutils.WaitForLastPipe()

}
