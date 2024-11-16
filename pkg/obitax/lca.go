package obitax

import (
	"fmt"
)

// LCA computes the Lowest Common Ancestor (LCA) of two Taxon instances.
// It traverses the paths from both taxa to the root and finds the deepest
// common node in the taxonomy hierarchy.
//
// Parameters:
//   - t2: A pointer to another Taxon instance to find the LCA with.
//
// Returns:
//   - A pointer to the Taxon representing the LCA of the two taxa, or an error
//     if either of the taxa is nil, if they are not in the same taxonomy, or
//     if the taxonomy is unrooted.
func (t1 *Taxon) LCA(t2 *Taxon) (*Taxon, error) {
	if t1 == nil || t1.Node == nil {
		return nil, fmt.Errorf("try to get LCA of nil taxon")
	}

	if t2 == nil || t2.Node == nil {
		return nil, fmt.Errorf("try to get LCA of nil taxon")
	}

	if t1.Taxonomy != t2.Taxonomy {
		return nil, fmt.Errorf("taxa are not in the same taxonomy")
	}

	if !t1.Taxonomy.HasRoot() {
		return nil, fmt.Errorf("taxa belong to an unrooted taxonomy")
	}

	p1 := t1.Path()
	p2 := t2.Path()

	i1 := p1.Len() - 1
	i2 := p2.Len() - 1

	for i1 >= 0 && i2 >= 0 && p1.slice[i1].id == p2.slice[i2].id {
		i1--
		i2--
	}

	return &Taxon{
		Taxonomy: t1.Taxonomy,
		Node:     p1.slice[i1+1],
	}, nil
}
