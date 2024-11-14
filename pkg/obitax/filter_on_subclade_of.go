package obitax

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

func (set *TaxonSet) IFilterOnSubcladeOf(taxon *Taxon) *ITaxon {
	return set.Iterator().IFilterOnSubcladeOf(taxon)
}

func (slice *TaxonSlice) IFilterOnSubcladeOf(taxon *Taxon) *ITaxon {
	return slice.Iterator().IFilterOnSubcladeOf(taxon)
}

func (taxonomy *Taxonomy) IFilterOnSubcladeOf(taxon *Taxon) *ITaxon {
	return taxonomy.Iterator().IFilterOnSubcladeOf(taxon)
}

func (iterator *ITaxon) IFilterBelongingSubclades(clades *TaxonSet) *ITaxon {

	if clades.Len() == 0 {
		return iterator
	}

	// Considers the second simplest case when only
	// a single subclase is provided
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
