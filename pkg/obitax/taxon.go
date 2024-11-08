package obitax

import (
	"iter"
	"regexp"

	log "github.com/sirupsen/logrus"
)

// Taxon represents a taxon within a taxonomy, encapsulating both the taxonomy
// it belongs to and the specific taxon node information.
//
// Fields:
//   - Taxonomy: A pointer to the Taxonomy[T] instance that this taxon is part of.
//   - Node: A pointer to the TaxNode[T] instance representing the specific taxon.
type Taxon struct {
	Taxonomy *Taxonomy
	Node     *TaxNode
}

// String returns a string representation of the Taxon.
// It formats the output to include the taxonomy code, the taxon ID, and the scientific name.
//
// Returns:
//   - A formatted string representing the Taxon in the form "taxonomy_code:taxon_id [scientific_name]".
func (taxon *Taxon) String() string {
	return taxon.Node.String(taxon.Taxonomy.code)
}

// ScientificName returns the scientific name of the Taxon.
// It retrieves the scientific name from the underlying TaxNode associated with the taxon.
//
// Returns:
//   - The scientific name of the taxon as a string.
func (taxon *Taxon) ScientificName() string {
	return taxon.Node.ScientificName()
}

func (taxon *Taxon) Name(class string) string {
	return taxon.Node.Name(class)
}

func (taxon *Taxon) IsNameEqual(name string) bool {
	return taxon.Node.IsNameEqual(name)
}

func (taxon *Taxon) IsNameMatching(pattern *regexp.Regexp) bool {
	return taxon.Node.IsNameMatching(pattern)
}

func (taxon *Taxon) SetName(name, class string) {
	class = taxon.Taxonomy.nameclasses.Innerize(class)
	taxon.Node.SetName(name, class)
}

// Rank returns the rank of the Taxon.
// It retrieves the rank from the underlying TaxNode associated with the taxon.
//
// Returns:
//   - The rank of the taxon as a string (e.g., species, genus, family).
func (taxon *Taxon) Rank() string {
	return taxon.Node.Rank()
}

// Parent returns a pointer to the parent Taxon of the current Taxon.
// It retrieves the parent identifier from the underlying TaxNode and uses it
// to create a new Taxon instance representing the parent taxon.
//
// Returns:
//   - A pointer to the parent Taxon[T]. If the parent does not exist, it returns
//     a Taxon with a nil Node.
func (taxon *Taxon) Parent() *Taxon {
	pid := taxon.Node.ParentId()
	return &Taxon{taxon.Taxonomy,
		taxon.Taxonomy.nodes.Get(pid)}
}

// IPath returns an iterator that yields the path from the current Taxon to the root Taxon
// in the associated Taxonomy. It traverses up the taxonomy hierarchy until it reaches the root.
//
// Returns:
//   - An iterator function that takes a yield function as an argument. The yield function
//     is called with each Taxon in the path from the current taxon to the root. If the
//     taxonomy has no root node, the method logs a fatal error and terminates the program.
func (taxon *Taxon) IPath() iter.Seq[*Taxon] {
	if taxon.Taxonomy.root == nil {
		log.Fatalf("Taxon[%v].IPath(): Taxonomy has no root node", taxon.Taxonomy.name)
	}

	return func(yield func(*Taxon) bool) {
		for taxon.Node.parent != taxon.Taxonomy.root.id {
			if !yield(taxon) {
				return
			}

			taxon = taxon.Parent()
		}

		yield(taxon)

	}
}

// Path returns a slice of TaxNode[T] representing the path from the current Taxon
// to the root Taxon in the associated Taxonomy. It collects all the nodes in the path
// using the IPath method and returns them as a TaxonSlice.
//
// Returns:
//   - A pointer to a TaxonSlice[T] containing the TaxNode[T] instances in the path
//     from the current taxon to the root.
func (taxon *Taxon) Path() *TaxonSlice {
	s := make([]*TaxNode, 0, 10)

	for t := range taxon.IPath() {
		s = append(s, t.Node)
	}

	return &TaxonSlice{
		slice:    s,
		taxonomy: taxon.Taxonomy,
	}
}

// HasRankDefined checks if any taxon in the path from the current Taxon to the root
// has the specified rank defined. It iterates through the path using the IPath method
// and returns true if a match is found; otherwise, it returns false.
//
// Parameters:
//   - rank: A string representing the rank to check for (e.g., "species", "genus").
//
// Returns:
//   - A boolean indicating whether any taxon in the path has the specified rank defined.
func (taxon *Taxon) HasRankDefined(rank string) bool {
	for t := range taxon.IPath() {
		if t.Node.Rank() == rank {
			return true
		}
	}

	return false
}

// TaxonAtRank returns the first Taxon in the path from the current Taxon to the root
// that has the specified rank defined. It iterates through the path using the IPath method
// and returns the matching Taxon if found; otherwise, it returns nil.
//
// Parameters:
//   - rank: A string representing the rank to search for (e.g., "species", "genus").
//
// Returns:
//   - A pointer to the Taxon[T] that matches the specified rank, or nil if no such taxon exists
//     in the path to the root.
func (taxon *Taxon) TaxonAtRank(rank string) *Taxon {
	for t := range taxon.IPath() {
		if t.Node.Rank() == rank {
			return t
		}
	}

	return nil
}

// Species returns the first Taxon in the path from the current Taxon to the root
// that has the rank "species" defined. It utilizes the TaxonAtRank method to find
// the matching Taxon.
//
// Returns:
//   - A pointer to the Taxon[T] that matches the "species" rank, or nil if no such taxon
//     exists in the path to the root.
func (taxon *Taxon) Species() *Taxon {
	return taxon.TaxonAtRank("species")
}

// Genus returns the first Taxon in the path from the current Taxon to the root
// that has the rank "genus" defined. It utilizes the TaxonAtRank method to find
// the matching Taxon.
//
// Returns:
//   - A pointer to the Taxon[T] that matches the "genus" rank, or nil if no such taxon
//     exists in the path to the root.
func (taxon *Taxon) Genus() *Taxon {
	return taxon.TaxonAtRank("genus")
}

// Family returns the first Taxon in the path from the current Taxon to the root
// that has the rank "family" defined. It utilizes the TaxonAtRank method to find
// the matching Taxon.
//
// Returns:
//   - A pointer to the Taxon[T] that matches the "family" rank, or nil if no such taxon
//     exists in the path to the root.
func (taxon *Taxon) Family() *Taxon {
	return taxon.TaxonAtRank("family")
}
