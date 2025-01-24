package main

import (
	"log"
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obifind"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

func main() {
	optionParser := obioptions.GenerateOptionParser(obifind.OptionSet)

	_, args := optionParser(os.Args)

	var iterator *obitax.ITaxon

	switch {
	case obifind.CLIRequestsPathForTaxid() != "NA":

		taxon := obitax.DefaultTaxonomy().Taxon(obifind.CLIRequestsPathForTaxid())

		if taxon == nil {
			log.Fatalf("Cannot identify the requested taxon: %s",
				obifind.CLIRequestsPathForTaxid())
		}

		s := taxon.Path()

		if s == nil {
			log.Fatalf("Cannot extract taxonomic path describing %s", taxon.String())
		}

		iterator = s.Iterator()

		if obifind.CLIWithQuery() {
			iterator = iterator.AddMetadata("query", taxon.String())
		}

	case len(args) == 0:
		iterator = obitax.DefaultTaxonomy().Iterator()
	default:
		iters := make([]*obitax.ITaxon, len(args))

		for i, pat := range args {
			ii := obitax.DefaultTaxonomy().IFilterOnName(pat, obifind.CLIFixedPattern(), true)
			if obifind.CLIWithQuery() {
				ii = ii.AddMetadata("query", pat)
			}
			iters[i] = ii
		}

		iterator = iters[0]

		if len(iters) > 1 {
			iterator = iterator.Concat(iters[1:]...)
		}
	}

	iterator = obifind.CLITaxonRestrictions(iterator)
	obifind.CLICSVTaxaWriter(iterator, true)

	obiutils.WaitForLastPipe()

}
