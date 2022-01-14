package obitax

import (
	"fmt"
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

// Returns a TaxonSet listing the requested taxon and all
// its ancestors in the taxonomy down to the root.
func (taxonomy *Taxonomy) Path(taxid int) (*TaxonSlice, error) {
	taxon, err := taxonomy.Taxon(taxid)

	if err != nil {
		return nil, err
	}

	return taxon.Path()
}
