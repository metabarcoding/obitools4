/*
Package obitax provides functionality for managing taxonomic data structures,
specifically for representing and manipulating collections of taxa within a taxonomy.
It includes the TaxonSet structure, which holds mappings of taxon identifiers to their
corresponding TaxNode instances, along with methods for managing and querying these taxa.
*/

package obitax

import log "github.com/sirupsen/logrus"

// TaxonSet represents a collection of taxa within a taxonomy.
// It holds a mapping of taxon identifiers to their corresponding TaxNode instances,
// as well as a reference to the associated Taxonomy.
//
// Fields:
//   - set: A map that associates taxon identifiers of type *string with their corresponding TaxNode instances.
//   - nalias: The number of aliases in the TaxonSet.
//   - taxonomy: A pointer to the Taxonomy instance that this TaxonSet belongs to.
type TaxonSet struct {
	set      map[*string]*TaxNode
	nalias   int
	taxonomy *Taxonomy
}

// NewTaxonSet creates a new TaxonSet associated with the given Taxonomy.
// It initializes the set as an empty map and sets the alias count to zero.
//
// Returns:
//   - A pointer to the newly created TaxonSet.
func (taxonomy *Taxonomy) NewTaxonSet() *TaxonSet {
	return &TaxonSet{
		set:      make(map[*string]*TaxNode),
		nalias:   0,
		taxonomy: taxonomy.OrDefault(true),
	}
}

// Get retrieves the TaxNode associated with the specified taxon identifier.
// It returns the TaxNode if it exists in the TaxonSet; otherwise, it returns nil.
//
// Parameters:
//   - id: A pointer to the taxon identifier for which the TaxNode is to be retrieved.
//
// Returns:
//   - A pointer to the TaxNode associated with the provided identifier, or nil
//     if no such taxon exists in the set.
func (set *TaxonSet) Get(id *string) *Taxon {
	if set == nil {
		return nil
	}

	node := set.set[id]
	if node == nil {
		return nil
	}

	return &Taxon{
		Taxonomy: set.taxonomy,
		Node:     set.set[id],
	}
}

// Len returns the number of unique taxa in the TaxonSet.
// It calculates the count by subtracting the number of aliases from the total
// number of entries in the set.
//
// Returns:
//   - An integer representing the count of unique taxa in the TaxonSet.
func (set *TaxonSet) Len() int {
	if set == nil {
		return 0
	}
	return len(set.set) - set.nalias
}

// Insert adds a TaxNode to the TaxonSet. If a taxon with the same identifier
// already exists in the set, it updates the reference. If the existing taxon was
// an alias, its alias count is decremented.
//
// Parameters:
//   - taxon: A pointer to the TaxNode instance to be added to the TaxonSet.
//
// Behavior:
//   - If a taxon with the same identifier already exists and is different from the
//     new taxon, the alias count is decremented.
func (set *TaxonSet) Insert(node *TaxNode) *TaxonSet {
	if set == nil {
		log.Panic("Cannot insert node into nil TaxonSet")
	}

	if old := set.set[node.id]; old != nil && old.id != node.id {
		set.nalias--
	}
	set.set[node.id] = node

	return set
}

// InsertTaxon adds a Taxon to the TaxonSet. It verifies that the Taxon belongs
// to the same Taxonomy as the TaxonSet before insertion. If they do not match,
// it logs a fatal error and terminates the program.
//
// Parameters:
//   - taxon: A pointer to the Taxon instance to be added to the TaxonSet.
func (set *TaxonSet) InsertTaxon(taxon *Taxon) *TaxonSet {
	if set == nil {
		set = taxon.Taxonomy.NewTaxonSet()
	}

	if set.taxonomy != taxon.Taxonomy {
		log.Fatalf(
			"Cannot insert taxon %s into taxon set belonging to %s taxonomy",
			taxon.String(),
			set.taxonomy.name,
		)
	}

	return set.Insert(taxon.Node)
}

// Taxonomy returns a pointer to the Taxonomy instance that this TaxonSet belongs to.
//
// Returns:
//   - A pointer to the Taxonomy instance that this TaxonSet belongs to.
func (set *TaxonSet) Taxonomy() *Taxonomy {
	if set == nil {
		return nil
	}

	return set.taxonomy
}

// Alias associates a given alias string with a specified TaxNode in the TaxonSet.
// It first converts the alias to its corresponding identifier using the Id method.
// If the original taxon is not part of the taxon set, it logs a fatal error and terminates the program.
//
// Parameters:
//   - alias: A pointer to a string representing the alias to be associated with the taxon node.
//   - node: A pointer to the TaxNode instance that the alias will refer to.
//
// Behavior:
//   - If the original taxon corresponding to the alias is not part of the taxon set,
//     the method will log a fatal error and terminate the program.
func (set *TaxonSet) Alias(id *string, taxon *Taxon) {
	if set == nil {
		log.Panic("Cannot add alias to a nil TaxonSet")
	}

	original := set.Get(taxon.Node.id)
	if original == nil {
		log.Fatalf("Original taxon %v is not part of taxon set", id)
	}
	set.set[id] = taxon.Node
	set.nalias++
}

// IsAlias checks if the given identifier corresponds to an alias in the TaxonSet.
// It retrieves the TaxNode associated with the identifier and returns true if the
// node exists and its identifier is different from the provided identifier; otherwise, it returns false.
//
// Parameters:
//   - id: A pointer to the identifier to be checked for alias status.
//
// Returns:
//   - A boolean indicating whether the identifier corresponds to an alias in the set.
func (set *TaxonSet) IsAlias(id *string) bool {
	taxon := set.Get(id)
	return taxon != nil && taxon.Node.id != id
}

// IsATaxon checks if the given ID corresponds to a valid taxon node in the TaxonSet.
// It returns true if the node exists and its ID matches the provided ID; otherwise, it returns false.
// If the ID corresponds to an alias, it will return false.
//
// Parameters:
//   - id: A pointer to the identifier of the taxon to check.
//
// Returns:
//   - A boolean indicating whether the specified ID corresponds to a valid taxon node.
func (set *TaxonSet) IsATaxon(id *string) bool {
	taxon := set.Get(id)
	return taxon != nil && taxon.Node.id == id
}

// Contains checks if the TaxonSet contains a taxon node with the specified ID.
// It returns true if the node exists in the set; otherwise, it returns false.
// If the ID corresponds to an alias, it will return true if the alias exists.
//
// Parameters:
//   - id: A pointer to the identifier of the taxon to check for presence in the set.
//
// Returns:
//   - A boolean indicating whether the TaxonSet contains a taxon node with the specified ID.
func (set *TaxonSet) Contains(id *string) bool {
	node := set.Get(id)
	return node != nil
}

func (set *TaxonSet) Sort() *TaxonSlice {
	if set == nil {
		return nil
	}

	taxonomy := set.Taxonomy()
	taxa := taxonomy.NewTaxonSlice(0, set.Len())
	parent := make(map[*TaxNode]bool, set.Len())

	pushed := true

	for pushed {
		pushed = false
		for _, node := range set.set {
			if !parent[node] && (parent[set.Get(node.parent).Node] ||
				!set.Contains(node.parent) ||
				node == taxonomy.Root().Node) {
				pushed = true
				taxa.slice = append(taxa.slice, node)
				parent[node] = true
			}
		}
	}

	return taxa
}
