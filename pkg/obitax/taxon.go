package obitax

import (
	"errors"
	"iter"
	"regexp"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	log "github.com/sirupsen/logrus"
)

// Taxon represents a taxon within a taxonomy, encapsulating both the taxonomy
// it belongs to and the specific taxon node information.
//
// Fields:
//   - Taxonomy: A pointer to the Taxonomy instance that this taxon is part of.
//   - Node: A pointer to the TaxNode instance representing the specific taxon.
type Taxon struct {
	Taxonomy *Taxonomy
	Metadata *map[string]*interface{}
	Node     *TaxNode
}

// String returns a string representation of the Taxon.
// It formats the output to include the taxonomy code, the taxon ID, and the scientific name.
//
// Returns:
//   - A formatted string representing the Taxon in the form "taxonomy_code:taxon_id [scientific_name]".
func (taxon *Taxon) String() string {
	if taxon == nil || taxon.Node == nil {
		return "NA"
	}
	return taxon.Node.String(taxon.Taxonomy.code)
}

func (taxon *Taxon) HasScientificName() bool {
	return taxon != nil && taxon.Node.HasScientificName()
}

// ScientificName returns the scientific name of the Taxon.
// It retrieves the scientific name from the underlying TaxNode associated with the taxon.
//
// Returns:
//   - The scientific name of the taxon as a string.
func (taxon *Taxon) ScientificName() string {
	if taxon == nil {
		return "NA"
	}
	return taxon.Node.ScientificName()
}

// Name retrieves the name of the Taxon based on the specified class.
// It uses the taxonomy's name classes to format the name appropriately.
//
// Parameters:
//   - class: A string representing the name class to use for retrieval.
//
// Returns:
//   - The name of the taxon as a string.
func (taxon *Taxon) Name(class string) string {
	if taxon == nil {
		return "NA"
	}
	pclass := taxon.Taxonomy.nameclasses.Innerize(class)
	return taxon.Node.Name(pclass)
}

// IsNameEqual checks if the given name is equal to the name of the Taxon.
// It compares the provided name with the name stored in the TaxNode.
//
// Parameters:
//   - name: A string representing the name to compare against.
//
// Returns:
//   - A boolean indicating whether the names are equal.
func (taxon *Taxon) IsNameEqual(name string, ignoreCase bool) bool {
	if taxon == nil {
		return false
	}

	return taxon.Node.IsNameEqual(name, ignoreCase)
}

// IsNameMatching checks if the name of the Taxon matches the given regular expression pattern.
//
// Parameters:
//   - pattern: A pointer to a compiled regular expression to match against the taxon's name.
//
// Returns:
//   - A boolean indicating whether the taxon's name matches the specified pattern.
func (taxon *Taxon) IsNameMatching(pattern *regexp.Regexp) bool {
	if taxon == nil {
		return false
	}

	return taxon.Node.IsNameMatching(pattern)
}

// SetName sets the name of the Taxon based on the provided name and class.
// It logs a panic if the taxon pointer is nil.
//
// Parameters:
//   - name: A string representing the new name to set for the taxon.
//   - class: A string representing the name class to associate with the taxon.
func (taxon *Taxon) SetName(name, class string) {
	if taxon == nil {
		log.Panicf("nil taxon pointer for name %s [%s]", name, class)
	}

	pclass := taxon.Taxonomy.nameclasses.Innerize(class)
	pname := taxon.Taxonomy.names.Innerize(name)
	taxon.Node.SetName(pname, pclass)
}

// IsRoot checks if the Taxon is the root of the taxonomy.
// It returns true if the taxon is nil or if it matches the root node of the taxonomy.
//
// Returns:
//   - A boolean indicating whether the Taxon is the root of the taxonomy.
func (taxon *Taxon) IsRoot() bool {
	if taxon == nil {
		return true
	}

	return taxon.Taxonomy.root == taxon.Node
}

// Rank returns the rank of the Taxon.
// It retrieves the rank from the underlying TaxNode associated with the taxon.
//
// Returns:
//   - The rank of the taxon as a string (e.g., species, genus, family).
func (taxon *Taxon) Rank() string {
	if taxon == nil {
		return "NA"
	}
	return taxon.Node.Rank()
}

// Parent returns a pointer to the parent Taxon of the current Taxon.
// It retrieves the parent identifier from the underlying TaxNode and uses it
// to create a new Taxon instance representing the parent taxon.
//
// Returns:
//   - A pointer to the parent Taxon. If the parent does not exist, it returns
//     a Taxon with a nil Node.
func (taxon *Taxon) Parent() *Taxon {
	if taxon == nil {
		return nil
	}

	pid := taxon.Node.ParentId()
	return taxon.Taxonomy.nodes.Get(pid)
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
		for !taxon.IsRoot() {
			if !yield(taxon) {
				return
			}

			taxon = taxon.Parent()
		}

		if taxon != nil {
			yield(taxon)
		}
	}
}

// Path returns a slice of TaxNode representing the path from the current Taxon.
// The first element of the slice is the current Taxon, and the last element is the
// to the root Taxon in the associated Taxonomy.
//
// Returns:
//   - A pointer to a TaxonSlice containing the TaxNode instances in the path
//     from the current taxon to the root. If the taxon is nil, it returns nil.
func (taxon *Taxon) Path() *TaxonSlice {
	if taxon == nil {
		return nil
	}

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
	if taxon == nil {
		return false
	}

	prank := taxon.Taxonomy.ranks.Innerize(rank)
	for t := range taxon.IPath() {
		if t.Node.rank == prank {
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
//   - A pointer to the Taxon that matches the specified rank, or nil if no such taxon exists
//     in the path to the root.
func (taxon *Taxon) TaxonAtRank(rank string) *Taxon {
	if taxon == nil {
		return nil
	}

	prank := taxon.Taxonomy.ranks.Innerize(rank)

	for t := range taxon.IPath() {
		if t.Node.rank == prank {
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
//   - A pointer to the Taxon that matches the "species" rank, or nil if no such taxon
//     exists in the path to the root.
func (taxon *Taxon) Species() *Taxon {
	return taxon.TaxonAtRank("species")
}

// Genus returns the first Taxon in the path from the current Taxon to the root
// that has the rank "genus" defined. It utilizes the TaxonAtRank method to find
// the matching Taxon.
//
// Returns:
//   - A pointer to the Taxon that matches the "genus" rank, or nil if no such taxon
//     exists in the path to the root.
func (taxon *Taxon) Genus() *Taxon {
	return taxon.TaxonAtRank("genus")
}

// Family returns the first Taxon in the path from the current Taxon to the root
// that has the rank "family" defined. It utilizes the TaxonAtRank method to find
// the matching Taxon.
//
// Returns:
//   - A pointer to the Taxon that matches the "family" rank, or nil if no such taxon
//     exists in the path to the root.
func (taxon *Taxon) Family() *Taxon {
	return taxon.TaxonAtRank("family")
}

func (taxon *Taxon) SetMetadata(name string, value interface{}) *Taxon {
	if taxon == nil {
		return nil
	}

	if taxon.Metadata == nil {
		m := make(map[string]*interface{})
		taxon.Metadata = &m
	}
	(*taxon.Metadata)[name] = &value
	return taxon
}

func (taxon *Taxon) GetMetadata(name string) *interface{} {
	if taxon == nil || taxon.Metadata == nil {
		return nil
	}
	return (*taxon.Metadata)[name]
}

func (taxon *Taxon) HasMetadata(name string) bool {
	if taxon == nil || taxon.Metadata == nil {
		return false
	}
	_, ok := (*taxon.Metadata)[name]
	return ok
}

func (taxon *Taxon) RemoveMetadata(name string) {
	if taxon == nil || taxon.Metadata == nil {
		return
	}
	delete(*taxon.Metadata, name)
}

func (taxon *Taxon) MetadataAsString(name string) string {
	meta := taxon.GetMetadata(name)
	if meta == nil {
		return ""
	}
	value, err := obiutils.InterfaceToString(*meta)

	if err != nil {
		return ""
	}

	return value
}

func (taxon *Taxon) MetadataKeys() []string {
	if taxon == nil || taxon.Metadata == nil {
		return nil
	}
	keys := make([]string, 0, len(*taxon.Metadata))
	for k := range *taxon.Metadata {
		keys = append(keys, k)
	}
	return keys
}

func (taxon *Taxon) MetadataValues() []interface{} {
	if taxon == nil || taxon.Metadata == nil {
		return nil
	}
	values := make([]interface{}, 0, len(*taxon.Metadata))
	for _, v := range *taxon.Metadata {
		values = append(values, v)
	}
	return values
}

func (taxon *Taxon) MetadataStringValues() []string {
	if taxon == nil || taxon.Metadata == nil {
		return nil
	}
	values := make([]string, 0, len(*taxon.Metadata))
	for _, v := range *taxon.Metadata {
		value, err := obiutils.InterfaceToString(v)
		if err != nil {
			value = ""
		}
		values = append(values, value)
	}
	return values
}

func (taxon *Taxon) SameAs(other *Taxon) bool {
	if taxon == nil || other == nil {
		return false
	}

	return taxon.Taxonomy == other.Taxonomy && taxon.Node.id == other.Node.id
}

func (taxon *Taxon) AddChild(child string, replace bool) (*Taxon, error) {
	if taxon == nil {
		return nil, errors.New("nil taxon")
	}

	code, taxid, scientific_name, rank, err := ParseTaxonString(child)

	if err != nil {
		return nil, err
	}

	if taxon.Taxonomy.code != code {
		return nil, errors.New("taxonomy code mismatch")
	}

	newTaxon, err := taxon.Taxonomy.AddTaxon(taxid, *taxon.Node.id, rank, false, replace)

	if err != nil {
		return nil, err
	}

	newTaxon.SetName(scientific_name, "scientific name")

	return newTaxon, nil
}
