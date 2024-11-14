package obitax

import (
	"regexp"
)

func (taxonomy *Taxonomy) IFilterOnName(name string, strict bool) *ITaxon {
	if strict {
		nodes, ok := taxonomy.index[taxonomy.names.Innerize(name)]
		if ok {
			return nodes.Iterator()
		} else {
			empty := taxonomy.NewTaxonSet()
			return empty.Iterator()
		}
	}

	return taxonomy.Iterator().IFilterOnName(name, strict)
}

func (iterator *ITaxon) IFilterOnName(name string, strict bool) *ITaxon {
	newIterator := NewITaxon()
	sentTaxa := make(map[*string]bool)

	if strict {
		go func() {
			for iterator.Next() {
				taxon := iterator.Get()
				node := taxon.Node
				if _, ok := sentTaxa[node.id]; !ok {
					if taxon.IsNameEqual(name) {
						sentTaxa[node.id] = true
						newIterator.source <- taxon
					}
				}
			}
			close(newIterator.source)
		}()
	} else {
		pattern := regexp.MustCompile(name)

		go func() {
			for iterator.Next() {
				taxon := iterator.Get()
				node := taxon.Node
				if _, ok := sentTaxa[node.id]; !ok {
					if taxon.IsNameMatching(pattern) {
						sentTaxa[node.id] = true
						newIterator.source <- taxon
					}
				}
			}
			close(newIterator.source)
		}()
	}

	return newIterator
}
