package obiprimers

import (
	"fmt"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiprimer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	log "github.com/sirupsen/logrus"
)

const noTaxonomyAdvice = "no taxonomy: use the -t option, or give sequences annotated with their " +
	"taxonomic path (obiannotate --taxonomic-path)"

// dropTaxonomicPath deletes the taxonomic path of the sequences. The primer
// search needs the taxa of the sequences, which it reads from their taxid, not
// their path: once the taxonomy is built, the paths (a list of about twenty
// strings per sequence) only take memory.
func dropTaxonomicPath(sequences obiiter.IBioSequence) obiiter.IBioSequence {
	return sequences.MakeIWorker(obiseq.AnnotatorToSeqWorker(func(s *obiseq.BioSequence) {
		s.DeleteAttribute(obiseq.TaxonomicPathAttribute)
	}), false)
}

// embeddedTaxonomyInMemory extracts the taxonomy from the taxonomic paths
// carried by the sequences, which it has to hold in memory since the iterator
// cannot be read twice. It returns an iterator over the same sequences.
func embeddedTaxonomyInMemory(sequences obiiter.IBioSequence) (obiiter.IBioSequence, *obitax.Taxonomy, error) {
	source, slice := sequences.Load()
	taxonomy, err := slice.ExtractTaxonomy(nil, false)
	if err != nil {
		return obiiter.NilIBioSequence, nil, fmt.Errorf("%s: %w", noTaxonomyAdvice, err)
	}
	// every sequence is held in memory here: free the paths right away
	for _, s := range slice {
		s.DeleteAttribute(obiseq.TaxonomicPathAttribute)
	}
	return obiiter.IBatchOver(source, slice, obidefault.BatchSize()), taxonomy, nil
}

// readSequencesAndTaxonomy returns the sequences and the taxonomy to use: the
// given one when there is one (the -t option) and otherwise the taxonomy
// embedded in the sequences. Like obitag, it needs no -t option when the
// sequence files were annotated with obiannotate --taxonomic-path.
//
// For files the embedded taxonomy is collected by a first pass over them that
// keeps nothing but the taxonomic paths, and the sequences are then read a
// second time: the selection of the examples needs the taxonomy, so reading in
// one pass would mean holding every sequence in memory. Standard input cannot
// be read twice, its sequences are held in memory.
func readSequencesAndTaxonomy(args []string, given *obitax.Taxonomy) (obiiter.IBioSequence, *obitax.Taxonomy, error) {
	if given != nil {
		sequences, err := obiconvert.CLIReadBioSequences(args...)
		if err != nil {
			return obiiter.NilIBioSequence, nil, err
		}
		return dropTaxonomicPath(sequences), given, nil
	}

	if len(args) == 0 {
		log.Infof("No taxonomy given: reading the standard input in memory to extract its embedded taxonomy")
		sequences, err := obiconvert.CLIReadBioSequences()
		if err != nil {
			return obiiter.NilIBioSequence, nil, err
		}
		return embeddedTaxonomyInMemory(sequences)
	}

	log.Infof("No taxonomy given: extracting the taxonomy embedded in the sequence files (first pass)")
	first, err := obiconvert.CLIReadBioSequences(args...)
	if err != nil {
		return obiiter.NilIBioSequence, nil, err
	}
	taxonomy, err := first.ExtractTaxonomy(false)
	if err != nil || taxonomy == nil {
		if err == nil {
			err = fmt.Errorf("the sequences carry no taxonomic path")
		}
		return obiiter.NilIBioSequence, nil, fmt.Errorf("%s: %w", noTaxonomyAdvice, err)
	}
	log.Infof("Embedded taxonomy extracted: %d taxa", taxonomy.Len())

	sequences, err := obiconvert.CLIReadBioSequences(args...)
	if err != nil {
		return obiiter.NilIBioSequence, nil, err
	}
	return dropTaxonomicPath(sequences), taxonomy, nil
}

// CLIReadSequencesAndTaxonomy reads the sequences of the command line and
// returns them with the taxonomy to use, set as the default taxonomy: the one
// of the -t option, or else the one embedded in the sequences.
func CLIReadSequencesAndTaxonomy(args []string) (obiiter.IBioSequence, *obitax.Taxonomy, error) {
	given := obitax.DefaultTaxonomy()
	sequences, taxonomy, err := readSequencesAndTaxonomy(args, given)
	if err != nil {
		return obiiter.NilIBioSequence, nil, err
	}
	if given == nil {
		taxonomy.SetAsDefault()
	}
	return sequences, taxonomy, nil
}

// CLIPrimerPairs reads the sequences, selects the examples and the
// counterexamples from the taxonomy options, and searches the primer pairs.
func CLIPrimerPairs(sequences obiiter.IBioSequence, taxonomy *obitax.Taxonomy) ([]obiprimer.PairStats, error) {
	if taxonomy == nil {
		return nil, fmt.Errorf(noTaxonomyAdvice)
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
	log.Infof("%d example sequences (%d %s), %d counterexamples (%d %s), %d other sequences",
		counts.Examples, taxa.ExampleTaxa, CLIRank(),
		counts.CounterExamples, taxa.CounterExamples, CLIRank(), counts.Others)

	pairs, info, err := obiprimer.Search(corpus, taxa, params)
	if info != nil {
		log.Infof("%d strict words (found without error in at least %d examples)", info.StrictWords, info.MinStrict)
		if params.NoMultiMatch {
			log.Infof("no multi-match: %d strict words dropped as repeated, %d words dropped for matching several times in a sequence",
				info.Repeated, info.MultiMatch)
		}
		log.Infof("%d approximate occurrences; %d conserved primers, %d good",
			info.Matches, info.Pair.Primers, info.Pair.GoodPrimers)
		log.Infof("%d candidate pairs, %d amplifying enough examples, %d possibly specific enough, %d retained",
			info.Pair.CandidatePairs, info.Pair.RetainedPairs, info.Pair.Candidates, len(pairs))
	}
	return pairs, err
}
