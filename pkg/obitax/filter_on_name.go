package obitax

import (
	"regexp"
)

func (taxonomy *Taxonomy) IFilterOnName(name string, strict bool) *ITaxonSet {
	if strict {
		nodes, ok := taxonomy.index[name]
		if ok {
			return nodes.Iterator()
		} else {
			empty := make(TaxonSet)
			return (&empty).Iterator()
		}
	}

	return taxonomy.Iterator().IFilterOnName(name, strict)
}

func (iterator *ITaxonSet) IFilterOnName(name string, strict bool) *ITaxonSet {
	new_iterator := NewITaxonSet()
	sentTaxa := make(map[int]bool)

	if strict {
		go func() {
			for iterator.Next() {
				taxon := iterator.Get()
				if _, ok := sentTaxa[taxon.taxid]; !ok {
					if taxon.IsNameEqual(name) {
						sentTaxa[taxon.taxid] = true
						new_iterator.source <- taxon
					}
				}
			}
			close(new_iterator.source)
		}()
	} else {
		pattern := regexp.MustCompile(name)

		go func() {
			for iterator.Next() {
				taxon := iterator.Get()
				if _, ok := sentTaxa[taxon.taxid]; !ok {
					if taxon.IsNameMatching(pattern) {
						sentTaxa[taxon.taxid] = true
						new_iterator.source <- taxon
					}
				}
			}
			close(new_iterator.source)
		}()
	}

	return new_iterator
}
