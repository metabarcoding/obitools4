package obitax

// IFilterOnSubcladeOf filters the iterator to include only those Taxon instances
// that are subclades of the specified Taxon. It returns a new ITaxon iterator
// containing the filtered results.
//
// Parameters:
//   - taxon: A pointer to the Taxon to filter subclades against.
//
// Returns:
//   - A pointer to a new ITaxon iterator containing only the Taxon instances
//     that are subclades of the specified Taxon.
func (iterator *ITaxon) IFilterOnSubcladeOf(taxon *Taxon) *ITaxon {
	newIter := NewITaxon()

	go func() {
		for iterator.Next() {
			tx := iterator.Get()
			if tx.IsSubCladeOf(taxon) {
				newIter.source <- tx
			}
		}
		close(newIter.source)
	}()

	return newIter
}

// IFilterOnSubcladeOf filters the TaxonSet to include only those Taxon instances
// that are subclades of the specified Taxon. It returns a new ITaxon iterator
// containing the filtered results.
//
// Parameters:
//   - taxon: A pointer to the Taxon to filter subclades against.
//
// Returns:
//   - A pointer to a new ITaxon iterator containing only the Taxon instances
//     that are subclades of the specified Taxon.
func (set *TaxonSet) IFilterOnSubcladeOf(taxon *Taxon) *ITaxon {
	return set.Iterator().IFilterOnSubcladeOf(taxon)
}

// IFilterOnSubcladeOf filters the TaxonSlice to include only those Taxon instances
// that are subclades of the specified Taxon. It returns a new ITaxon iterator
// containing the filtered results.
//
// Parameters:
//   - taxon: A pointer to the Taxon to filter subclades against.
//
// Returns:
//   - A pointer to a new ITaxon iterator containing only the Taxon instances
//     that are subclades of the specified Taxon.
func (slice *TaxonSlice) IFilterOnSubcladeOf(taxon *Taxon) *ITaxon {
	return slice.Iterator().IFilterOnSubcladeOf(taxon)
}

// IFilterOnSubcladeOf filters the Taxonomy to include only those Taxon instances
// that are subclades of the specified Taxon. It returns a new ITaxon iterator
// containing the filtered results.
//
// Parameters:
//   - taxon: A pointer to the Taxon to filter subclades against.
//
// Returns:
//   - A pointer to a new ITaxon iterator containing only the Taxon instances
//     that are subclades of the specified Taxon.
func (taxonomy *Taxonomy) IFilterOnSubcladeOf(taxon *Taxon) *ITaxon {
	return taxonomy.Iterator().IFilterOnSubcladeOf(taxon)
}

// IFilterBelongingSubclades filters the iterator to include only those Taxon instances
// that belong to any of the specified subclades. It returns a new ITaxon iterator
// containing the filtered results.
//
// Parameters:
//   - clades: A pointer to a TaxonSet containing the subclades to filter against.
//
// Returns:
//   - A pointer to a new ITaxon iterator containing only the Taxon instances
//     that belong to the specified subclades. If the clades set is empty,
//     it returns the original iterator.
func (iterator *ITaxon) IFilterBelongingSubclades(clades *TaxonSet) *ITaxon {
	if clades.Len() == 0 {
		return iterator
	}

	// Considers the second simplest case when only
	// a single subclade is provided
	if clades.Len() == 1 {
		keys := make([]*string, 0, len(clades.set))
		for k := range clades.set {
			keys = append(keys, k)
		}

		return iterator.IFilterOnSubcladeOf(clades.Get(keys[0]))
	}

	newIter := NewITaxon()

	go func() {
		for iterator.Next() {
			tx := iterator.Get()
			if tx.IsBelongingSubclades(clades) {
				newIter.source <- tx
			}
		}
		close(newIter.source)
	}()

	return newIter
}
