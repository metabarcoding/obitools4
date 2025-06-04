package main

import (
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiitercsv"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obicsv"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obitaxonomy"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"

	log "github.com/sirupsen/logrus"
)

func main() {
	optionParser := obioptions.GenerateOptionParser(
		"obitaxonomy",
		"manipulates and queries taxonomy",
		obitaxonomy.OptionSet)

	_, args := optionParser(os.Args)

	var iterator *obitax.ITaxon

	if obitaxonomy.CLIDownloadNCBI() {
		err := obitaxonomy.CLIDownloadNCBITaxdump()
		if err != nil {
			log.Errorf("Cannot download NCBI taxonomy: %s", err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	if !obidefault.HasSelectedTaxonomy() {
		log.Fatal("you must indicate a taxonomy using the -t or --taxonomy option")
	}

	switch {
	case obitaxonomy.CLIAskForRankList():
		newIter := obiitercsv.NewICSVRecord()
		newIter.Add(1)
		newIter.AppendField("rank")
		go func() {
			ranks := obitax.DefaultTaxonomy().RankList()
			data := make([]obiitercsv.CSVRecord, len(ranks))

			for i, rank := range ranks {
				record := make(obiitercsv.CSVRecord)
				record["rank"] = rank
				data[i] = record
			}
			newIter.Push(obiitercsv.MakeCSVRecordBatch(obitax.DefaultTaxonomy().Name(), 0, data))
			newIter.Close()
			newIter.Done()
		}()
		obicsv.CLICSVWriter(newIter, true)
		obiutils.WaitForLastPipe()
		os.Exit(0)

	case obitaxonomy.CLIExtractTaxonomy():
		iter, err := obiconvert.CLIReadBioSequences(args...)
		iter = iter.NumberSequences(1, true)

		if err != nil {
			log.Fatalf("Cannot extract taxonomy: %v", err)
		}

		taxonomy, err := iter.ExtractTaxonomy(obitaxonomy.CLINewickWithLeaves())

		if err != nil {
			log.Fatalf("Cannot extract taxonomy: %v", err)
		}

		taxonomy.SetAsDefault()

		log.Infof("Number of extracted taxa: %d", taxonomy.Len())
		iterator = taxonomy.AsTaxonSet().Sort().Iterator()

	case obitaxonomy.CLIDumpSubtaxonomy():
		iterator = obitaxonomy.CLISubTaxonomyIterator()

	case obitaxonomy.CLIRequestsPathForTaxid() != "NA":

		taxon, isAlias, err := obitax.DefaultTaxonomy().Taxon(obitaxonomy.CLIRequestsPathForTaxid())

		if err != nil {
			log.Fatalf("Cannot identify the requested taxon: %s (%v)",
				obitaxonomy.CLIRequestsPathForTaxid(), err)
		}

		if isAlias {
			if obidefault.FailOnTaxonomy() {
				log.Fatalf("Taxon %s is an alias for %s", taxon.String(), taxon.Parent().String())
			}
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

	if obitaxonomy.CLIAsNewick() {
		obitaxonomy.CLINewickWriter(iterator, true)
	} else {
		obitaxonomy.CLICSVTaxaWriter(iterator, true)
	}

	obiutils.WaitForLastPipe()

}
