package obiprimer

import (
	"fmt"

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

// AssignTaxa determines the taxon of every sequence of the corpus at rank.
func (c *Corpus) AssignTaxa(taxonomy *obitax.Taxonomy, rank string) (*Taxa, error) {
	if taxonomy == nil {
		return nil, fmt.Errorf("obiprimer: a taxonomy is required to compute taxon statistics")
	}
	ranks := taxonomy.RankList()
	known := false
	for _, r := range ranks {
		if r == rank {
			known = true
			break
		}
	}
	if !known {
		return nil, fmt.Errorf("obiprimer: unknown taxonomic rank %q", rank)
	}

	t := &Taxa{Rank: rank, Of: make([]int32, len(c.Seqs))}
	index := map[*obitax.TaxNode]int32{}
	exSeen := map[int32]struct{}{}
	coSeen := map[int32]struct{}{}

	for i, seq := range c.Seqs {
		t.Of[i] = -1
		taxon := seq.Taxon(taxonomy).TaxonAtRank(rank)
		if taxon == nil {
			continue
		}
		id, ok := index[taxon.Node]
		if !ok {
			id = int32(len(t.nodes))
			index[taxon.Node] = id
			t.nodes = append(t.nodes, taxon)
		}
		t.Of[i] = id
		switch c.Roles[i] {
		case Example:
			exSeen[id] = struct{}{}
		case CounterExample:
			coSeen[id] = struct{}{}
		}
	}
	t.ExampleTaxa, t.CounterExamples = len(exSeen), len(coSeen)
	return t, nil
}
