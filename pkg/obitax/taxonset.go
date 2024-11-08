package obitax

import log "github.com/sirupsen/logrus"

// TaxonSet represents a collection of taxa within a taxonomy.
// It holds a mapping of taxon identifiers to their corresponding TaxNode instances,
// as well as a reference to the associated Taxonomy.
//
// Fields:
//   - set: A map that associates taxon identifiers of type T with their corresponding TaxNode[T] instances.
//   - taxonomy: A pointer to the Taxonomy[T] instance that this TaxonSet belongs to.
type TaxonSet struct {
	set      map[string]*TaxNode
	nalias   int
	taxonomy *Taxonomy
}

// Get retrieves the TaxNode[T] associated with the specified taxon identifier.
// It returns the TaxNode if it exists in the TaxonSet; otherwise, it returns nil.
//
// Parameters:
//   - i: The taxon identifier of type T for which the TaxNode is to be retrieved.
//
// Returns:
//   - A pointer to the TaxNode[T] associated with the provided identifier, or nil
//     if no such taxon exists in the set.
func (set *TaxonSet) Get(i string) *TaxNode {
	return set.set[i]
}

// Len returns the number of unique taxa in the TaxonSet.
// It calculates the count by subtracting the number of aliases from the total
// number of entries in the set.
//
// Returns:
//   - An integer representing the count of unique taxa in the TaxonSet.
func (set *TaxonSet) Len() int {
	return len(set.set) - set.nalias
}

// Insert adds a TaxNode[T] to the TaxonSet. If a taxon with the same identifier
// already exists in the set, it updates the reference. If the existing taxon was
// an alias, its alias count is decremented.
//
// Parameters:
//   - taxon: A pointer to the TaxNode[T] instance to be added to the TaxonSet.
//
// Behavior:
//   - If a taxon with the same identifier already exists and is different from the
//     new taxon, the alias count is decremented.
func (set *TaxonSet) Insert(taxon *TaxNode) {
	if old := set.set[taxon.id]; old != nil && old.id != taxon.id {
		set.nalias--
	}
	set.set[taxon.id] = taxon
}

// Taxonomy returns a pointer to the Taxonomy[T] instance that this TaxonSet belongs to.
//
// Returns:
//   - A pointer to the Taxonomy[T] instance that this TaxonSet belongs to
func (set *TaxonSet) Taxonomy() *Taxonomy {
	return set.taxonomy
}

// Alias associates a given alias string with a specified TaxNode in the TaxonSet.
// It first converts the alias to its corresponding identifier using the Id method.
// If the original taxon is not part of the taxon set, it logs a fatal error and terminates the program.
//
// Parameters:
//   - alias: A string representing the alias to be associated with the taxon node.
//   - node: A pointer to the TaxNode[T] instance that the alias will refer to.
//
// Behavior:
//   - If the original taxon corresponding to the alias is not part of the taxon set,
//     the method will log a fatal error and terminate the program.
func (set *TaxonSet) Alias(id string, node *TaxNode) {
	original := set.Get(node.id)
	if original != nil {
		log.Fatalf("Original taxon %v is not part of taxon set", id)
	}
	set.set[id] = node
	set.nalias++
}

// IsAlias checks if the given identifier corresponds to an alias in the TaxonSet.
// It retrieves the TaxNode associated with the identifier and returns true if the
// node exists and its identifier is different from the provided identifier; otherwise, it returns false.
//
// Parameters:
//   - id: The identifier of type T to be checked for alias status.
//
// Returns:
//   - A boolean indicating whether the identifier corresponds to an alias in the set.
func (set *TaxonSet) IsAlias(id string) bool {
	node := set.Get(id)
	return node != nil && node.id != id
}

// IsATaxon checks if the given ID corresponds to a valid taxon node in the TaxonSet.
// It returns true if the node exists and its ID matches the provided ID; otherwise, it returns false.
// id corresponding to alias returns false.
//
// Parameters:
//   - id: The identifier of the taxon to check.
//
// Returns:
//   - A boolean indicating whether the specified ID corresponds to a valid taxon node.
func (set *TaxonSet) IsATaxon(id string) bool {
	node := set.Get(id)
	return node != nil && node.id == id
}

// Contains checks if the TaxonSet contains a taxon node with the specified ID.
// It returns true if the node exists in the set; otherwise, it returns false.
// id corresponding to alias or true taxa returns true.
//
// Parameters:
//   - id: The identifier of the taxon to check for presence in the set.
//
// Returns:
//   - A boolean indicating whether the TaxonSet contains a taxon node with the specified ID.
func (set *TaxonSet) Contains(id string) bool {
	node := set.Get(id)
	return node != nil
}
