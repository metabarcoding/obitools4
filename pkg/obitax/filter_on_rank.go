package obitax

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

func (set *TaxonSet) IFilterOnTaxRank(rank string) *ITaxon {
	return set.Iterator().IFilterOnTaxRank(rank)
}

func (slice *TaxonSlice) IFilterOnTaxRank(rank string) *ITaxon {
	return slice.Iterator().IFilterOnTaxRank(rank)
}

func (taxonomy *Taxonomy) IFilterOnTaxRank(rank string) *ITaxon {
	return taxonomy.Iterator().IFilterOnTaxRank(rank)
}
