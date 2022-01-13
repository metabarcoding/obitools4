package obitax

type ITaxonSet struct {
	source     chan *TaxNode
	current    *TaxNode
	finished   bool
	p_finished *bool
}

func NewITaxonSet() *ITaxonSet {
	i := ITaxonSet{make(chan *TaxNode), nil, false, nil}
	i.p_finished = &i.finished
	return &i
}

func (set *TaxonSet) Iterator() *ITaxonSet {
	i := NewITaxonSet()

	go func() {
		for _, t := range *set {
			i.source <- t
		}
		close(i.source)
	}()

	return i
}

func (set *TaxonSlice) Iterator() *ITaxonSet {
	i := NewITaxonSet()

	go func() {
		for _, t := range *set {
			i.source <- t
		}
		close(i.source)
	}()

	return i
}

func (taxonmy *Taxonomy) iterator() *ITaxonSet {
	return taxonmy.nodes.Iterator()
}

func (iterator *ITaxonSet) Next() bool {
	if *(iterator.p_finished) {
		return false
	}
	next, ok := (<-iterator.source)

	if ok {
		iterator.current = next
		return true
	}

	iterator.current = nil
	*iterator.p_finished = true
	return false
}

// The 'Get' method returns the instance of *TaxNode
// currently pointed by the iterator. You have to use the
// 'Next' method to move to the next entry before calling
// 'Get' to retreive the following instance.
func (iterator *ITaxonSet) Get() *TaxNode {
	return iterator.current
}

// Finished returns 'true' value if no more data is available
// from the iterator.
func (iterator *ITaxonSet) Finished() bool {
	return *iterator.p_finished
}

func (iterator *ITaxonSet) Split() *ITaxonSet {
	new_iter := ITaxonSet{iterator.source, nil, false, iterator.p_finished}
	return &new_iter
}

func (iterator *ITaxonSet) TaxonSet() *TaxonSet {
	set := make(TaxonSet)

	for iterator.Next() {
		taxon := iterator.Get()
		set[taxon.taxid] = taxon
	}
	return &set
}

func (iterator *ITaxonSet) TaxonSlice() *TaxonSlice {
	slice := make(TaxonSlice, 0)

	for iterator.Next() {
		taxon := iterator.Get()
		slice = append(slice, taxon)
	}
	return &slice
}
