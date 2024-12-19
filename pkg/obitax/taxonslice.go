/*
Package obitax provides functionality for handling taxonomic data structures,
specifically for representing and manipulating collections of taxon nodes
within a taxonomy.

The primary data structure is the TaxonSlice, which encapsulates a slice of
TaxNode instances and provides methods for accessing, counting, and
formatting these nodes.
*/

package obitax

import (
	"bytes"
	"fmt"
	"log"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

// TaxonSlice represents a slice of TaxNode instances within a taxonomy.
// It encapsulates a collection of taxon nodes and the taxonomy they belong to.
//
// Fields:
//   - slice: A slice of pointers to TaxNode representing the taxon nodes.
//   - taxonomy: A pointer to the Taxonomy instance that these taxon nodes are part of.
type TaxonSlice struct {
	slice    []*TaxNode
	taxonomy *Taxonomy
}

// NewTaxonSlice creates a new TaxonSlice with the specified size and capacity.
// It initializes the slice of TaxNode pointers and associates it with the given taxonomy.
//
// Parameters:
//   - size: The initial size of the slice.
//   - capacity: The capacity of the slice.
//
// Returns:
//   - A pointer to the newly created TaxonSlice.
func (taxonomy *Taxonomy) NewTaxonSlice(size, capacity int) *TaxonSlice {
	return &TaxonSlice{
		slice:    make([]*TaxNode, size, capacity),
		taxonomy: taxonomy.OrDefault(true),
	}
}

// Get retrieves the TaxNode at the specified index from the TaxonSlice.
// It returns the taxon node corresponding to the provided index.
//
// Parameters:
//   - i: An integer representing the index of the taxon node to retrieve.
//
// Returns:
//   - A pointer to the TaxNode at the specified index in the slice.
func (slice *TaxonSlice) Get(i int) *TaxNode {
	if slice == nil {
		return nil
	}
	return slice.slice[i]
}

func (slice *TaxonSlice) Taxon(i int) *Taxon {
	if slice == nil {
		return nil
	}
	return &Taxon{
		Node:     slice.slice[i],
		Taxonomy: slice.taxonomy,
	}
}

// Len returns the number of TaxNode instances in the TaxonSlice.
// It provides the count of taxon nodes contained within the slice.
//
// Returns:
//   - An integer representing the total number of taxon nodes in the TaxonSlice.
func (slice *TaxonSlice) Len() int {
	if slice == nil {
		return 0
	}
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
			*taxon.Id(),
			taxon.ScientificName(),
			taxon.Rank())

		for i := path.Len() - 2; i >= 0; i-- {
			taxon := path.slice[i]
			fmt.Fprintf(&buffer, "|%v@%s@%s",
				*taxon.Id(),
				taxon.ScientificName(),
				taxon.Rank())
		}
	}

	return buffer.String()
}

// Reverse reverses the order of the TaxonSlice.
// If inplace is true, the original slice is modified; otherwise, a new reversed
// TaxonSlice is returned.
//
// Parameters:
//   - inplace: A boolean indicating whether to reverse the slice in place.
//
// Returns:
//   - A pointer to the reversed TaxonSlice. If inplace is true, it returns the original slice.
func (slice *TaxonSlice) Reverse(inplace bool) *TaxonSlice {
	if slice == nil {
		return nil
	}

	rep := obiutils.Reverse(slice.slice, inplace)
	if inplace {
		return slice
	}

	return &TaxonSlice{
		taxonomy: slice.taxonomy,
		slice:    rep,
	}
}

func (slice *TaxonSlice) Set(index int, taxon *Taxon) *TaxonSlice {
	if slice.taxonomy != taxon.Taxonomy {
		log.Panic("Cannot add taxon from a different taxonomy")
	}

	slice.slice[index] = taxon.Node

	return slice
}
