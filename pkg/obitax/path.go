package obitax

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

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

func (taxon *TaxNode) TaxonAtRank(rank string) *TaxNode {
	for taxon.rank != rank && taxon != taxon.pparent {
		taxon = taxon.pparent

		if taxon == nil {
			log.Panicln("Taxonomy must be reindexed")
		}
	}

	if taxon == taxon.pparent {
		taxon = nil
	}

	return taxon
}

func (taxon *TaxNode) Species() *TaxNode {
	return taxon.TaxonAtRank("species")
}

func (taxon *TaxNode) Genus() *TaxNode {
	return taxon.TaxonAtRank("genus")
}

func (taxon *TaxNode) Family() *TaxNode {
	return taxon.TaxonAtRank("family")
}

// Returns a TaxonSet listing the requested taxon and all
// its ancestors in the taxonomy down to the root.
func (taxonomy *Taxonomy) Path(taxid int) (*TaxonSlice, error) {
	taxon, err := taxonomy.Taxon(taxid)

	if err != nil {
		return nil, err
	}

	return taxon.Path()
}
