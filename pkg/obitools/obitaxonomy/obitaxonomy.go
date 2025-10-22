package obitaxonomy

import (
	"fmt"
	"time"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiformats"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiitercsv"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obicsv"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"

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

	options := make([]obiformats.WithOption, 0)

	options = append(options,
		obiformats.OptionsWithPattern(CLIWithQuery()),
		obiformats.OptionsWithParent(CLIWithParent()),
		obiformats.OptionsWithRank(CLIWithRank()),
		obiformats.OptionsWithScientificName(CLIWithScientificName()),
		obiformats.OptionsWithPath(CLIWithPath()),
		obiformats.OptionsRawTaxid(obidefault.UseRawTaxids()),
		obiformats.OptionsSource(obidefault.SelectedTaxonomy()),
	)

	return obiformats.CSVTaxaIterator(iterator, options...)
}

func CLICSVTaxaWriter(iterator *obitax.ITaxon, terminalAction bool) *obiitercsv.ICSVRecord {
	return obicsv.CLICSVWriter(CLICSVTaxaIterator(iterator), terminalAction)
}

func CLINewickWriter(iterator *obitax.ITaxon,
	terminalAction bool) *obitax.ITaxon {

	var err error
	var newIter *obitax.ITaxon

	options := make([]obiformats.WithOption, 0)
	options = append(options, obiformats.OptionsCompressed(obidefault.CompressOutput()),
		obiformats.OptionsWithRank(CLIWithRank()),
		obiformats.OptionsWithScientificName(CLIWithScientificName()),
		obiformats.OptionsWithTaxid(true),
		obiformats.OptionWithoutRootPath(CLINewickWithoutRoot()),
	)

	filename := obiconvert.CLIOutPutFileName()

	if filename != "-" {
		newIter, err = obiformats.WriteNewickToFile(iterator, filename, options...)

		if err != nil {
			log.Fatalf("Cannot write to file : %+v", err)
		}

	} else {
		newIter, err = obiformats.WriteNewickToStdout(iterator, options...)

		if err != nil {
			log.Fatalf("Cannot write to stdout : %+v", err)
		}

	}

	if terminalAction {
		newIter.Consume()
		return nil
	}

	return newIter
}

func CLIDownloadNCBITaxdump() error {
	now := time.Now()
	dateStr := now.Format("20060102") // In Go, this specific date is used as reference for formatting

	filename := fmt.Sprintf("ncbitaxo_%s.tgz", dateStr)

	if obiconvert.CLIOutPutFileName() != "-" {
		filename = obiconvert.CLIOutPutFileName()
	}

	log.Infof("Downloading NCBI Taxdump to %s", filename)
	return obiutils.DownloadFile("https://ftp.ncbi.nlm.nih.gov/pub/taxonomy/taxdump.tar.gz", filename)

}
