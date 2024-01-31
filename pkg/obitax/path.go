package obitax

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

// Path generates the lineage path from the current taxon up to the root.
//
// This method does not take parameters as it is called on a TaxNode receiver.
// It returns a pointer to a TaxonSlice containing the path and an error if
// the taxonomy needs reindexing.
func (taxon *TaxNode) Path() (*TaxonSlice, error) {

	path := make(TaxonSlice, 0, 30)
	path = append(path, taxon)

	for taxon != taxon.pparent {
		taxon = taxon.pparent

		if taxon == nil {
			return nil, fmt.Errorf("Taxonomy must be reindexed")
		}

		path = append(path, taxon)
	}

	return &path, nil
}

// TaxonAtRank traverses up the taxonomy tree starting from the current
// node until it finds a node that matches the specified rank.
//
// If a node with the given rank is not found in the path to the root,
// or if the taxonomy tree is not properly indexed (i.e., a node's parent
// is itself), the function will return nil. In case the taxonomy needs
// reindexing, the function will panic.
//
// rank: the taxonomic rank to search for (e.g., "species", "genus").
//
// Returns a pointer to a TaxNode representing the node at the
// specified rank, or nil if no such node exists in the path.
func (taxon *TaxNode) TaxonAtRank(rank string) *TaxNode {
	for taxon.rank != rank && taxon != taxon.pparent {
		taxon = taxon.pparent

		if taxon == nil {
			log.Panicln("Taxonomy must be reindexed")
		}
	}

	if taxon == taxon.pparent && taxon.rank != rank {
		taxon = nil
	}

	return taxon
}

// Species retrieves the TaxNode corresponding to the species rank.
//
// This method does not take any parameters. It is a convenience
// wrapper around the TaxonAtRank method, specifically retrieving
// the species-level taxonomic classification for the calling TaxNode.
//
// Returns a pointer to the TaxNode representing the species.
func (taxon *TaxNode) Species() *TaxNode {
	return taxon.TaxonAtRank("species")
}

func (taxon *TaxNode) Genus() *TaxNode {
	return taxon.TaxonAtRank("genus")
}

func (taxon *TaxNode) Family() *TaxNode {
	return taxon.TaxonAtRank("family")
}

func (taxonomy *Taxonomy) Path(taxid int) (*TaxonSlice, error) {
	taxon, err := taxonomy.Taxon(taxid)

	if err != nil {
		return nil, err
	}

	return taxon.Path()
}
