package obitax

func (taxon *TaxNode) IsSubCladeOf(parent *TaxNode) bool {

	for taxon.taxid != parent.taxid && taxon.parent != taxon.taxid {
		taxon = taxon.pparent
	}

	return taxon.taxid == parent.taxid
}

func (taxon *TaxNode) IsBelongingSubclades(clades *TaxonSet) bool {
	_, ok := (*clades)[taxon.taxid]

	for !ok && taxon.parent != taxon.taxid {
		taxon = taxon.pparent
		_, ok = (*clades)[taxon.taxid]
	}

	return ok
}
