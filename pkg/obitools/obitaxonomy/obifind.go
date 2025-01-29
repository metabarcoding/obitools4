package obitaxonomy

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiitercsv"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obicsv"

	log "github.com/sirupsen/logrus"
)

func CLITaxonRestrictions(iterator *obitax.ITaxon) *obitax.ITaxon {

	if iterator == nil {
		return nil
	}

	clades, err := CLITaxonomicalRestrictions()

	if err != nil {
		log.Errorf("Error on taxonomy restriction: %v", err)
		return nil
	}

	iterator = CLIFilterRankRestriction(iterator.Split().IFilterBelongingSubclades(clades))
	return iterator
}

func CLIFilterRankRestriction(iterator *obitax.ITaxon) *obitax.ITaxon {
	if iterator == nil {
		return nil
	}

	rr := CLIRankRestriction()

	if rr != "" {
		iterator = iterator.IFilterOnTaxRank(rr)
	}

	return iterator
}

func CLISubTaxonomyIterator() *obitax.ITaxon {

	if CLIDumpSubtaxonomy() {
		return obitax.DefaultTaxonomy().ISubTaxonomy(CLISubTaxonomyNode())
	}

	log.Fatalf("No sub-taxonomy specified use the --dump option")
	return nil
}

func CLICSVTaxaIterator(iterator *obitax.ITaxon) *obiitercsv.ICSVRecord {
	if iterator == nil {
		return nil
	}

	options := make([]obitax.WithOption, 0)

	options = append(options,
		obitax.OptionsWithPattern(CLIWithQuery()),
		obitax.OptionsWithParent(CLIWithParent()),
		obitax.OptionsWithRank(CLIWithRank()),
		obitax.OptionsWithScientificName(CLIWithScientificName()),
		obitax.OptionsWithPath(CLIWithPath()),
		obitax.OptionsRawTaxid(CLIRawTaxid()),
		obitax.OptionsSource(obidefault.SelectedTaxonomy()),
	)

	return iterator.CSVTaxaIterator(options...)
}

func CLICSVTaxaWriter(iterator *obitax.ITaxon, terminalAction bool) *obiitercsv.ICSVRecord {
	return obicsv.CLICSVWriter(CLICSVTaxaIterator(iterator), terminalAction)
}
