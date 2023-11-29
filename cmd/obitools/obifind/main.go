package main

import (
	"fmt"
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obifind"
)

func main() {
	optionParser := obioptions.GenerateOptionParser(obifind.OptionSet)

	_, args := optionParser(os.Args)

	//prof, _ := os.Create("obifind.prof")
	//pprof.StartCPUProfile(prof)

	restrictions, err := obifind.ITaxonRestrictions()
	if err != nil {
		fmt.Printf("%+v", err)
	}

	switch {
	case obifind.CLIRequestsPathForTaxid() >= 0:
		taxonomy, err := obifind.CLILoadSelectedTaxonomy()
		if err != nil {
			fmt.Printf("%+v", err)
		}

		taxon, err := taxonomy.Taxon(obifind.CLIRequestsPathForTaxid())

		if err != nil {
			fmt.Printf("%+v", err)
		}

		s, err := taxon.Path()

		if err != nil {
			fmt.Printf("%+v", err)
		}

		obifind.TaxonWriter(s.Iterator(),
			fmt.Sprintf("path:%d", taxon.Taxid()))

	case len(args) == 0:
		taxonomy, err := obifind.CLILoadSelectedTaxonomy()
		if err != nil {
			fmt.Printf("%+v", err)
		}

		obifind.TaxonWriter(restrictions(taxonomy.Iterator()), "")

	default:
		matcher, err := obifind.ITaxonNameMatcher()

		if err != nil {
			fmt.Printf("%+v", err)
		}

		for _, pattern := range args {
			s := restrictions(matcher(pattern))
			obifind.TaxonWriter(s, pattern)
		}
	}

	//pprof.StopCPUProfile()
}
