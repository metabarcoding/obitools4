package obitax

import log "github.com/sirupsen/logrus"

// IsSubCladeOf checks if the current Taxon is a subclade of the specified parent Taxon.
// It returns true if the current Taxon is a descendant of the parent Taxon in the taxonomy hierarchy.
//
// Parameters:
//   - parent: A pointer to the parent Taxon to check against.
//
// Returns:
//   - A boolean indicating whether the current Taxon is a subclade of the parent Taxon.
//   - Logs a fatal error if the two taxa do not belong to the same taxonomy.
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

// IsBelongingSubclades checks if the current Taxon belongs to any of the specified subclades.
// It traverses up the taxonomy hierarchy to determine if the current Taxon or any of its ancestors
// belong to the provided TaxonSet.
//
// Parameters:
//   - clades: A pointer to a TaxonSet containing the subclades to check against.
//
// Returns:
//   - A boolean indicating whether the current Taxon or any of its ancestors belong to the specified subclades.
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
