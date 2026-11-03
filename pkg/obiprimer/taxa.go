package obiprimer

import (
	"fmt"
	"slices"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
)

// Taxa gives every corpus sequence its taxon at one rank (the species by
// default in ecoPrimers), as a dense index, and the number of distinct taxa
// among the examples and among the counterexamples.
//
// A sequence whose taxon has no ancestor at that rank (or no taxon at all) has
// the index -1; it is never counted as a taxon.
type Taxa struct {
	Rank            string
	Of              []int32 // Of[i] is the taxon index of Corpus.Seqs[i], or -1
	ExampleTaxa     int
	CounterExamples int // distinct taxa among the counterexamples
	nodes           []*obitax.Taxon
}

// Name returns the taxon with the given index, for reporting.
func (t *Taxa) Name(i int32) *obitax.Taxon { return t.nodes[i] }

// checkRank reports an error if the taxonomy has no such rank.
func checkRank(taxonomy *obitax.Taxonomy, rank string) error {
	if taxonomy == nil {
		return fmt.Errorf("obiprimer: a taxonomy is required to compute taxon statistics")
	}
	if !slices.Contains(taxonomy.RankList(), rank) {
		return fmt.Errorf("obiprimer: unknown taxonomic rank %q", rank)
	}
	return nil
}

// AssignTaxa determines the taxon of every sequence of the corpus at rank. The
// sequences must not be packed (see LoadPackedCorpus).
func (c *Corpus) AssignTaxa(taxonomy *obitax.Taxonomy, rank string) (*Taxa, error) {
	if err := checkRank(taxonomy, rank); err != nil {
		return nil, err
	}
	taxons := make([]*obitax.Taxon, len(c.Seqs))
	for i, seq := range c.Seqs {
		taxons[i] = seq.Taxon(taxonomy).TaxonAtRank(rank)
	}
	return buildTaxa(rank, taxons, c.Roles)
}

// buildTaxa gives the sequences, whose taxon at rank is taxons[i] (nil when
// there is none), their dense taxon index.
func buildTaxa(rank string, taxons []*obitax.Taxon, roles []Role) (*Taxa, error) {
	t := &Taxa{Rank: rank, Of: make([]int32, len(taxons))}
	index := map[*obitax.TaxNode]int32{}
	exSeen := map[int32]struct{}{}
	coSeen := map[int32]struct{}{}

	for i, taxon := range taxons {
		t.Of[i] = -1
		if taxon == nil {
			continue
		}
		id, ok := index[taxon.Node]
		if !ok {
			if len(t.nodes) >= maxTaxa {
				return nil, fmt.Errorf("obiprimer: more than %d taxa at rank %q", maxTaxa, rank)
			}
			id = int32(len(t.nodes))
			index[taxon.Node] = id
			t.nodes = append(t.nodes, taxon)
		}
		t.Of[i] = id
		switch roles[i] {
		case Example:
			exSeen[id] = struct{}{}
		case CounterExample:
			coSeen[id] = struct{}{}
		}
	}
	t.ExampleTaxa, t.CounterExamples = len(exSeen), len(coSeen)
	return t, nil
}
