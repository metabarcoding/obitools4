package obitax

import log "github.com/sirupsen/logrus"

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

func (taxon *Taxon) IsBelongingSubclades(clades *TaxonSet) bool {
	ok := clades.Contains(taxon.Node.id)

	for !ok && !taxon.IsRoot() {
		taxon = taxon.Parent()
		ok = clades.Contains(taxon.Node.id)
	}

	if taxon.IsRoot() {
		ok = clades.Contains(taxon.Node.id)
	}

	return ok
}
