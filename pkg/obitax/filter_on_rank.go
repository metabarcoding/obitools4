package obitax

// IFilterOnTaxRank filters the iterator to include only those Taxon instances
// that match the specified taxonomic rank. It returns a new ITaxon iterator
// containing the filtered results.
//
// Parameters:
//   - rank: A string representing the taxonomic rank to filter by.
//
// Returns:
//   - A pointer to a new ITaxon iterator containing only the Taxon instances
//     that have the specified taxonomic rank.
func (iterator *ITaxon) IFilterOnTaxRank(rank string) *ITaxon {
	newIter := NewITaxon()
	var prank *string
	var ptax *Taxonomy

	go func() {
		for iterator.Next() {

			taxon := iterator.Get()
			if ptax != taxon.Taxonomy {
				ptax = taxon.Taxonomy
				prank = ptax.ranks.Innerize(rank)
			}

			if taxon.Node.rank == prank {
				newIter.source <- taxon
			}
		}
		close(newIter.source)
	}()

	return newIter
}

// IFilterOnTaxRank filters the TaxonSet to include only those Taxon instances
// that match the specified taxonomic rank. It returns a new ITaxon iterator
// containing the filtered results.
//
// Parameters:
//   - rank: A string representing the taxonomic rank to filter by.
//
// Returns:
//   - A pointer to a new ITaxon iterator containing only the Taxon instances
//     that have the specified taxonomic rank.
func (set *TaxonSet) IFilterOnTaxRank(rank string) *ITaxon {
	return set.Iterator().IFilterOnTaxRank(rank)
}

// IFilterOnTaxRank filters the TaxonSlice to include only those Taxon instances
// that match the specified taxonomic rank. It returns a new ITaxon iterator
// containing the filtered results.
//
// Parameters:
//   - rank: A string representing the taxonomic rank to filter by.
//
// Returns:
//   - A pointer to a new ITaxon iterator containing only the Taxon instances
//     that have the specified taxonomic rank.
func (slice *TaxonSlice) IFilterOnTaxRank(rank string) *ITaxon {
	return slice.Iterator().IFilterOnTaxRank(rank)
}

// IFilterOnTaxRank filters the Taxonomy to include only those Taxon instances
// that match the specified taxonomic rank. It returns a new ITaxon iterator
// containing the filtered results.
//
// Parameters:
//   - rank: A string representing the taxonomic rank to filter by.
//
// Returns:
//   - A pointer to a new ITaxon iterator containing only the Taxon instances
//     that have the specified taxonomic rank.
func (taxonomy *Taxonomy) IFilterOnTaxRank(rank string) *ITaxon {
	return taxonomy.Iterator().IFilterOnTaxRank(rank)
}
