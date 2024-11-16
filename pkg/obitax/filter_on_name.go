package obitax

import (
	"regexp"
)

// IFilterOnName filters the Taxon instances in the Taxonomy based on the specified name.
// If strict is true, it looks for an exact match of the name. If false, it allows for pattern matching.
// It returns a new ITaxon iterator containing the filtered results.
//
// Parameters:
//   - name: The name to filter Taxon instances by.
//   - strict: A boolean indicating whether to perform strict matching (true) or pattern matching (false).
//
// Returns:
//   - A pointer to a new ITaxon iterator containing only the Taxon instances that match the specified name.
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

// IFilterOnName filters the Taxon instances in the iterator based on the specified name.
// If strict is true, it looks for an exact match of the name. If false, it allows for pattern matching.
// It returns a new ITaxon iterator containing the filtered results.
//
// Parameters:
//   - name: The name to filter Taxon instances by.
//   - strict: A boolean indicating whether to perform strict matching (true) or pattern matching (false).
//
// Returns:
//   - A pointer to a new ITaxon iterator containing only the Taxon instances that match the specified name.
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
