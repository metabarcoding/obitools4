package obiprimers

import (
	"fmt"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiprimer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
	log "github.com/sirupsen/logrus"
)

// CLIPrimerPairs reads the sequences, selects the examples and the
// counterexamples from the taxonomy options, and searches the primer pairs.
// The default taxonomy must be loaded.
func CLIPrimerPairs(sequences obiiter.IBioSequence) ([]obiprimer.PairStats, error) {
	taxonomy := obitax.DefaultTaxonomy()
	if taxonomy == nil {
		return nil, fmt.Errorf("a taxonomy is required: use the -t option")
	}

	sel := obiprimer.Selection{}
	var err error
	if sel.Restrict, err = obiprimer.BelongingTo(taxonomy, CLIRestrictTaxa()); err != nil {
		return nil, err
	}
	if sel.Ignore, err = obiprimer.BelongingTo(taxonomy, CLIIgnoreTaxa()); err != nil {
		return nil, err
	}
	if sel.Except, err = obiprimer.BelongingTo(taxonomy, CLIExceptionTaxa()); err != nil {
		return nil, err
	}

	params, err := CLISearchParams()
	if err != nil {
		return nil, err
	}

	log.Infof("Reading the sequences")
	corpus, err := obiprimer.LoadCorpus(sequences, sel)
	if err != nil {
		return nil, err
	}
	taxa, err := corpus.AssignTaxa(taxonomy, CLIRank())
	if err != nil {
		return nil, err
	}

	counts := corpus.Counts()
	log.Infof("%d example sequences (%d %s), %d counterexamples (%d %s), %d ignored",
		counts.Examples, taxa.ExampleTaxa, CLIRank(),
		counts.CounterExamples, taxa.CounterExamples, CLIRank(), counts.Ignored)

	pairs, info, err := obiprimer.Search(corpus, taxa, params)
	if info != nil {
		log.Infof("%d strict words (found without error in at least %d examples)", info.StrictWords, info.MinStrict)
		log.Infof("%d approximate occurrences; %d conserved primers, %d good",
			info.Matches, info.Pair.Primers, info.Pair.GoodPrimers)
		log.Infof("%d candidate pairs, %d amplifying enough examples, %d retained",
			info.Pair.CandidatePairs, info.Pair.RetainedPairs, len(pairs))
	}
	return pairs, err
}
