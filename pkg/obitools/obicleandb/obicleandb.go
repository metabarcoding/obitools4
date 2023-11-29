package obicleandb

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obichunk"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obigrep"
)

func ICleanDB(itertator obiiter.IBioSequence) obiiter.IBioSequence {
	var rankPredicate obiseq.SequencePredicate

	options := make([]obichunk.WithOption, 0, 30)

	// Make sequence dereplication with a constraint on the taxid.
	// To be merged, both sequences must have the same taxid.

	options = append(options,
		obichunk.OptionBatchCount(100),
		obichunk.OptionSortOnMemory(),
		obichunk.OptionSubCategory("taxid"),
		obichunk.OptionsParallelWorkers(
			obioptions.CLIParallelWorkers()),
		obichunk.OptionsBatchSize(
			obioptions.CLIBatchSize()),
		obichunk.OptionNAValue("NA"),
	)

	unique, err := obichunk.IUniqueSequence(itertator, options...)

	if err != nil {
		log.Fatal(err)
	}

	taxonomy := obigrep.CLILoadSelectedTaxonomy()

	if len(obigrep.CLIRequiredRanks()) > 0 {
		rankPredicate = obigrep.CLIHasRankDefinedPredicate()
	} else {
		rankPredicate = taxonomy.HasRequiredRank("species").And(taxonomy.HasRequiredRank("genus")).And(taxonomy.HasRequiredRank("family"))
	}

	goodTaxa := taxonomy.IsAValidTaxon(CLIUpdateTaxids()).And(rankPredicate)

	usable := unique.FilterOn(goodTaxa,
		obioptions.CLIBatchSize(),
		obioptions.CLIParallelWorkers())

	annotated := usable.MakeIWorker(taxonomy.MakeSetSpeciesWorker(),
		obioptions.CLIParallelWorkers(),
	).MakeIWorker(taxonomy.MakeSetGenusWorker(),
		obioptions.CLIParallelWorkers(),
	).MakeIWorker(taxonomy.MakeSetFamilyWorker(),
		obioptions.CLIParallelWorkers(),
	)

	//	annotated.MakeIConditionalWorker(obiseq.IsMoreAbundantOrEqualTo(3),1000)

	return annotated
}
