package obitax

import "reflect"

func (iterator *ITaxonSet) IFilterOnSubcladeOf(taxon *TaxNode) *ITaxonSet {
	newIter := NewITaxonSet()

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

func (set *TaxonSet) IFilterOnSubcladeOf(taxon *TaxNode) *ITaxonSet {
	return set.Iterator().IFilterOnSubcladeOf(taxon)
}

func (slice *TaxonSlice) IFilterOnSubcladeOf(taxon *TaxNode) *ITaxonSet {
	return slice.Iterator().IFilterOnSubcladeOf(taxon)
}

func (taxonomy *Taxonomy) IFilterOnSubcladeOf(taxon *TaxNode) *ITaxonSet {
	return taxonomy.Iterator().IFilterOnSubcladeOf(taxon)
}

func (iterator *ITaxonSet) IFilterBelongingSubclades(clades *TaxonSet) *ITaxonSet {

	if len(*clades) == 0 {
		return iterator
	}

	// Considers the second simplest case when only
	// a single subclase is provided
	if len(*clades) == 1 {
		keys := reflect.ValueOf(*clades).MapKeys()
		return iterator.IFilterOnSubcladeOf((*clades)[int(keys[0].Int())])
	}

	newIter := NewITaxonSet()

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
