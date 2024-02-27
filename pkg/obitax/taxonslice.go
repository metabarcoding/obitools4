package obitax

import (
	"bytes"
	"fmt"
)

type TaxonSlice []*TaxNode

func (set *TaxonSlice) Get(i int) *TaxNode {
	return (*set)[i]
}

func (set *TaxonSlice) Len() int {
	return len(*set)
}

func (path *TaxonSlice) String() string {
	var buffer bytes.Buffer

	if len(*path) > 0 {
		taxon := (*path)[len(*path)-1]
		fmt.Fprintf(&buffer, "%d@%s@%s",
			taxon.Taxid(),
			taxon.ScientificName(),
			taxon.Rank())

		for i := len(*path) - 2; i >= 0; i-- {
			taxon := (*path)[i]
			fmt.Fprintf(&buffer, "|%d@%s@%s",
				taxon.Taxid(),
				taxon.ScientificName(),
				taxon.Rank())
		}
	}

	return buffer.String()
}
