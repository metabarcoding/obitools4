package obitax

type TaxonSlice []*TaxNode

func (set *TaxonSlice) Get(i int) *TaxNode {
	return (*set)[i]
}

func (set *TaxonSlice) Len() int {
	return len(*set)
}
