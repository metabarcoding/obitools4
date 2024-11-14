package obitax

type ITaxon struct {
	source     chan *Taxon
	current    *Taxon
	finished   bool
	p_finished *bool
}

func NewITaxon() *ITaxon {
	i := ITaxon{
		source:     make(chan *Taxon),
		current:    nil,
		finished:   false,
		p_finished: nil}
	i.p_finished = &i.finished
	return &i
}

func (set *TaxonSet) Iterator() *ITaxon {
	i := NewITaxon()

	go func() {
		for _, t := range set.set {
			i.source <- &Taxon{
				Taxonomy: set.taxonomy,
				Node:     t,
			}
		}
		close(i.source)
	}()

	return i
}

func (set *TaxonSlice) Iterator() *ITaxon {
	i := NewITaxon()

	go func() {
		for _, t := range set.slice {
			i.source <- &Taxon{
				Taxonomy: set.taxonomy,
				Node:     t,
			}
		}
		close(i.source)
	}()

	return i
}

func (taxonmy *Taxonomy) Iterator() *ITaxon {
	return taxonmy.nodes.Iterator()
}

func (iterator *ITaxon) Next() bool {
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
func (iterator *ITaxon) Get() *Taxon {
	return iterator.current
}

// Finished returns 'true' value if no more data is available
// from the iterator.
func (iterator *ITaxon) Finished() bool {
	return *iterator.p_finished
}

func (iterator *ITaxon) Split() *ITaxon {
	return &ITaxon{
		source:     iterator.source,
		current:    nil,
		finished:   false,
		p_finished: iterator.p_finished,
	}
}
