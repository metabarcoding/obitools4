package obifind

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

func CLICSVTaxaIterator(iterator *obitax.ITaxon) *obiitercsv.ICSVRecord {
	if iterator == nil {
		return nil
	}

	options := make([]WithOption, 0)

	options = append(options,
		OptionsWithPattern(CLIWithQuery()),
		OptionsWithParent(CLIWithParent()),
		OptionsWithRank(CLIWithRank()),
		OptionsWithScientificName(CLIWithScientificName()),
		OptionsWithPath(CLIWithPath()),
		OptionsRawTaxid(CLIRawTaxid()),
		OptionsSource(obidefault.SelectedTaxonomy()),
	)

	return NewCSVTaxaIterator(iterator, options...)
}

func CLICSVTaxaWriter(iterator *obitax.ITaxon, terminalAction bool) *obiitercsv.ICSVRecord {
	return obicsv.CLICSVWriter(CLICSVTaxaIterator(iterator), terminalAction)
}
