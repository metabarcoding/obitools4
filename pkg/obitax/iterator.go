package obitax

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	log "github.com/sirupsen/logrus"
)

// ITaxon represents an iterator for traversing Taxon instances.
// It provides methods to retrieve the next Taxon and check if the iteration is finished.
type ITaxon struct {
	source     chan *Taxon // Channel to receive Taxon instances
	current    *Taxon      // Current Taxon instance
	finished   bool        // Indicates if the iteration is finished
	p_finished *bool       // Pointer to the finished status
}

// NewITaxon creates a new ITaxon iterator instance and initializes its fields.
func NewITaxon() *ITaxon {
	i := ITaxon{
		source:     make(chan *Taxon),
		current:    nil,
		finished:   false,
		p_finished: nil,
	}
	i.p_finished = &i.finished
	return &i
}

// Iterator creates a new ITaxon iterator for the TaxonSet.
// It starts a goroutine to send Taxon instances from the set to the iterator's source channel.
func (set *TaxonSet) Iterator() *ITaxon {
	i := NewITaxon()

	go func() {
		for _, t := range set.set {
			taxon := &Taxon{
				Taxonomy: set.taxonomy,
				Metadata: nil,
				Node:     t,
			}
			i.Push(taxon)
		}
		close(i.source)
	}()

	return i
}

// Iterator creates a new ITaxon iterator for the TaxonSlice.
// It starts a goroutine to send Taxon instances from the slice to the iterator's source channel.
func (set *TaxonSlice) Iterator() *ITaxon {
	i := NewITaxon()

	go func() {
		for _, t := range set.slice {
			i.Push(&Taxon{
				Taxonomy: set.taxonomy,
				Node:     t,
			})
		}
		i.Close()
	}()

	return i
}

func (iterator *ITaxon) Push(taxon *Taxon) {
	iterator.source <- taxon
}

func (iterator *ITaxon) Close() {
	close(iterator.source)
}

// Iterator creates a new ITaxon iterator for the Taxonomy's nodes.
func (taxonomy *Taxonomy) Iterator() *ITaxon {
	return taxonomy.nodes.Iterator()
}

// Next advances the iterator to the next Taxon instance.
// It returns true if there is a next Taxon, and false if the iteration is finished.
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

// Get returns the current Taxon instance pointed to by the iterator.
// You must call 'Next' before calling 'Get' to retrieve the next instance.
func (iterator *ITaxon) Get() *Taxon {
	if iterator == nil {
		return nil
	}

	return iterator.current
}

// Finished returns true if no more data is available from the iterator.
func (iterator *ITaxon) Finished() bool {
	if iterator == nil {
		return true
	}

	return *iterator.p_finished
}

// Split creates a new ITaxon iterator that shares the same source channel
// and finished status as the original iterator.
func (iterator *ITaxon) Split() *ITaxon {
	if iterator == nil {
		return nil
	}
	return &ITaxon{
		source:     iterator.source,
		current:    nil,
		finished:   false,
		p_finished: iterator.p_finished,
	}
}

func (iterator *ITaxon) AddMetadata(name string, value interface{}) *ITaxon {
	if iterator == nil {
		return nil
	}

	i := NewITaxon()

	go func() {
		for iterator.Next() {
			taxon := iterator.Get()
			taxon.SetMetadata(name, value)
			i.Push(taxon)
		}
		i.Close()
	}()

	return i
}

func (iterator *ITaxon) Concat(iterators ...*ITaxon) *ITaxon {

	newIter := NewITaxon()

	go func() {
		if iterator != nil {
			for iterator.Next() {
				taxon := iterator.Get()
				newIter.Push(taxon)
			}
		}

		for _, iter := range iterators {
			if iter != nil {
				for iter.Next() {
					taxon := iter.Get()
					newIter.Push(taxon)
				}
			}
		}

		newIter.Close()
	}()

	return newIter
}

func (taxon *Taxon) ISubTaxonomy() *ITaxon {

	taxo := taxon.Taxonomy

	path := taxon.Path()
	lpath := path.Len()

	iter := NewITaxon()

	parents := map[*TaxNode]bool{taxon.Node: true}

	obiutils.RegisterAPipe()

	go func() {
		for i := lpath - 1; i >= 0; i-- {
			taxon := path.Taxon(i)
			iter.Push(taxon)
		}

		pushed := true

		log.Warn(parents)
		for pushed {
			itaxo := taxo.Iterator()
			pushed = false
			for itaxo.Next() {
				taxon := itaxo.Get()

				if !parents[taxon.Node] && parents[taxon.Parent().Node] {
					parents[taxon.Node] = true
					iter.Push(taxon)
					pushed = true
				}
			}
		}

		iter.Close()
		obiutils.UnregisterPipe()
	}()

	return iter
}

func (taxonomy *Taxonomy) ISubTaxonomy(taxid string) *ITaxon {
	taxon, err := taxonomy.Taxon(taxid)

	if err != nil {
		return nil
	}

	return taxon.ISubTaxonomy()
}
