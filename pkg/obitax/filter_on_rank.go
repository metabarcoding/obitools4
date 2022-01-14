package obitax

func (iterator *ITaxonSet) IFilterOnTaxRank(rank string) *ITaxonSet {
	newIter := NewITaxonSet()

	go func() {
		for iterator.Next() {
			taxon := iterator.Get()
			if taxon.rank == rank {
				newIter.source <- taxon
			}
		}
		close(newIter.source)
	}()

	return newIter
}

func (set *TaxonSet) IFilterOnTaxRank(rank string) *ITaxonSet {
	return set.Iterator().IFilterOnTaxRank(rank)
}

func (slice *TaxonSlice) IFilterOnTaxRank(rank string) *ITaxonSet {
	return slice.Iterator().IFilterOnTaxRank(rank)
}

func (taxonomy *Taxonomy) IFilterOnTaxRank(rank string) *ITaxonSet {
	return taxonomy.Iterator().IFilterOnTaxRank(rank)
}
