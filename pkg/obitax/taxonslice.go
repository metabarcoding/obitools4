package obitax

type TaxonSlice []*TaxNode

func (set *TaxonSlice) Get(i int) *TaxNode {
	return (*set)[i]
}

func (set *TaxonSlice) Length() int {
	return len(*set)
}
