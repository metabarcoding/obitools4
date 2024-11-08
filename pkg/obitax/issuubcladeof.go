package obitax

import "log"

func (taxon *Taxon) IsSubCladeOf(parent *Taxon) bool {

	if taxon.Taxonomy != parent.Taxonomy {
		log.Fatalf(
			"Both taxa %s and %s must belong to the same taxonomy",
			taxon.String(),
			parent.String(),
		)
	}

	for t := range taxon.IPath() {
		if t.Node.Id() == parent.Node.Id() {
			return true
		}
	}

	return false
}
