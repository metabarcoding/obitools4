package obitax

func (iterator *ITaxonSet) IFilterOnTaxRank(rank string) *ITaxonSet {
	new_iter := NewITaxonSet()

	go func() {
		for iterator.Next() {
			taxon := iterator.Get()
			if taxon.rank == rank {
				new_iter.source <- taxon
			}
		}
		close(new_iter.source)
	}()

	return new_iter
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
