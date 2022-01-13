package obitax

type TaxonSet map[int]*TaxNode

func (set *TaxonSet) Get(i int) *TaxNode {
	return (*set)[i]
}

func (set *TaxonSet) Length() int {
	return len(*set)
}

func (set *TaxonSet) Inserts(taxon *TaxNode) {
	(*set)[taxon.taxid] = taxon
}
