package obitax

import (
	"bytes"
	"fmt"
)

// TaxonSlice represents a slice of TaxNode[T] instances within a taxonomy.
// It encapsulates a collection of taxon nodes and the taxonomy they belong to.
//
// Fields:
//   - slice: A slice of pointers to TaxNode[T] representing the taxon nodes.
//   - taxonomy: A pointer to the Taxonomy[T] instance that these taxon nodes are part of.
type TaxonSlice struct {
	slice    []*TaxNode
	taxonomy *Taxonomy
}

// Get retrieves the TaxNode[T] at the specified index from the TaxonSlice.
// It returns the taxon node corresponding to the provided index.
//
// Parameters:
//   - i: An integer representing the index of the taxon node to retrieve.
//
// Returns:
//   - A pointer to the TaxNode[T] at the specified index in the slice.
func (slice *TaxonSlice) Get(i int) *TaxNode {
	return slice.slice[i]
}

// Len returns the number of TaxNode[T] instances in the TaxonSlice.
// It provides the count of taxon nodes contained within the slice.
//
// Returns:
//   - An integer representing the total number of taxon nodes in the TaxonSlice.
func (slice *TaxonSlice) Len() int {
	return len(slice.slice)
}

// String returns a string representation of the TaxonSlice.
// It formats the output to include the IDs, scientific names, and ranks of the taxon nodes
// in the slice, concatenated in reverse order, separated by vertical bars.
//
// Returns:
//   - A formatted string representing the TaxonSlice, with each taxon in the format
//     "id@scientific_name@rank". If the slice is empty, it returns an empty string.
func (path *TaxonSlice) String() string {
	var buffer bytes.Buffer

	if path.Len() > 0 {
		taxon := path.slice[path.Len()-1]
		fmt.Fprintf(&buffer, "%v@%s@%s",
			taxon.Id(),
			taxon.ScientificName(),
			taxon.Rank())

		for i := path.Len() - 2; i >= 0; i-- {
			taxon := path.slice[i]
			fmt.Fprintf(&buffer, "|%v@%s@%s",
				taxon.Id(),
				taxon.ScientificName(),
				taxon.Rank())
		}
	}

	return buffer.String()
}
