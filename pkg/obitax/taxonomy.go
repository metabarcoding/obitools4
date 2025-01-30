/*
Package obitax provides functionality for managing taxonomic data structures,
including hierarchical classifications of taxa. It includes the Taxonomy struct,
which represents a taxonomy and provides methods for working with taxon identifiers
and retrieving information about taxa.
*/

package obitax

import (
	"errors"
	"fmt"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

// Taxonomy represents a hierarchical classification of taxa.
// It holds information about the taxonomy's name, code, ranks, nodes, root node, aliases, and an index.
// The generic type T is used to specify the type of taxon identifiers.
//
// Fields:
//   - name: The name of the taxonomy.
//   - code: A unique code representing the taxonomy.
//   - ids: A pointer to an InnerString instance that holds the taxon identifiers.
//   - ranks: A pointer to an InnerString instance that holds the ranks of the taxa.
//   - nameclasses: A pointer to an InnerString instance that holds the name classes.
//   - names: A pointer to an InnerString instance that holds the names of the taxa.
//   - nodes: A pointer to a TaxonSet containing all the nodes (taxa) in the taxonomy.
//   - root: A pointer to the root TaxNode of the taxonomy.
//   - matcher: A regular expression used for validating taxon identifiers.
//   - index: A map that indexes taxa by their string representation for quick access.
type Taxonomy struct {
	name        string
	code        string
	ids         *TaxidFactory
	ranks       *InnerString
	nameclasses *InnerString
	names       *InnerString
	nodes       *TaxonSet
	root        *TaxNode
	index       map[*string]*TaxonSet
}

// NewTaxonomy creates and initializes a new Taxonomy instance with the specified name and code.
// It sets up the necessary internal structures, including ranks, nodes, aliases, and an index.
//
// Parameters:
//   - name: The name of the taxonomy to be created.
//   - code: A unique code representing the taxonomy.
//   - codeCharacters: A string representing valid characters for the taxon identifiers.
//
// Returns:
//   - A pointer to the newly created Taxonomy instance.
func NewTaxonomy(name, code string, codeCharacters obiutils.AsciiSet) *Taxonomy {
	set := make(map[*string]*TaxNode)

	taxonomy := &Taxonomy{
		name:        name,
		code:        code,
		ids:         NewTaxidFactory(code, codeCharacters),
		ranks:       NewInnerString(),
		nameclasses: NewInnerString(),
		names:       NewInnerString(),
		nodes:       &TaxonSet{set: set},
		root:        nil,
		index:       make(map[*string]*TaxonSet),
	}

	taxonomy.nodes.taxonomy = taxonomy

	return taxonomy
}

// Id converts a given taxid string into the corresponding taxon identifier of type T.
// It uses a regular expression to validate and extract the taxid. If the taxid is invalid,
// the method returns an error along with a zero value of type T.
//
// Parameters:
//   - taxid: A string representation of the taxon identifier to be converted.
//
// Returns:
//   - The taxon identifier as a *string corresponding to the provided taxid.
//   - An error if the taxid is not valid or cannot be converted.
func (taxonomy *Taxonomy) Id(taxid string) (Taxid, error) {
	taxonomy = taxonomy.OrDefault(false)

	if taxonomy == nil {
		return nil, errors.New("Cannot extract Id from nil Taxonomy")
	}

	return taxonomy.ids.FromString(taxid)
}

// TaxidString retrieves the string representation of a taxon node identified by the given ID.
// It looks up the node in the taxonomy and returns its formatted string representation
// along with the taxonomy code. If the node does not exist, it returns an error.
//
// Parameters:
//   - id: The identifier of the taxon node to retrieve.
//
// Returns:
//   - A string representing the taxon node in the format "taxonomyCode:id [scientificName]",
//     or an error if the taxon node with the specified ID does not exist in the taxonomy.
func (taxonomy *Taxonomy) TaxidString(id string) (string, error) {
	taxonomy = taxonomy.OrDefault(false)

	pid, err := taxonomy.Id(id)

	if err != nil {
		return "", err
	}

	taxon := taxonomy.nodes.Get(pid)

	if taxon == nil {
		return "", fmt.Errorf("taxid %s is not part of the taxonomy", id)
	}

	return taxon.String(), nil
}

// Taxon retrieves the Taxon associated with the given taxid string.
// It first converts the taxid to its corresponding identifier using the Id method.
// If the taxon is not found, it logs a fatal error and terminates the program.
//
// Parameters:
//   - taxid: A string representation of the taxon identifier to be retrieved.
//
// Returns:
//   - A pointer to the Taxon instance associated with the provided taxid.
//   - If the taxid is unknown, the method will log a fatal error.
func (taxonomy *Taxonomy) Taxon(taxid string) (*Taxon, error) {
	taxonomy = taxonomy.OrDefault(false)
	if taxonomy == nil {
		return nil, errors.New("cannot extract taxon from nil taxonomy")
	}

	id, err := taxonomy.Id(taxid)

	if err != nil {
		return nil, fmt.Errorf("Taxid %s: %v", taxid, err)
	}

	taxon := taxonomy.nodes.Get(id)

	if taxon == nil {
		return nil,
			fmt.Errorf("Taxid %s is not part of the taxonomy %s",
				taxid,
				taxonomy.name)
	}

	return taxon, nil
}

// AsTaxonSet returns the set of taxon nodes contained within the Taxonomy.
// It provides access to the underlying collection of taxon nodes for further operations.
//
// Returns:
//   - A pointer to the TaxonSet representing the collection of taxon nodes in the taxonomy.
func (taxonomy *Taxonomy) AsTaxonSet() *TaxonSet {
	taxonomy = taxonomy.OrDefault(false)

	if taxonomy == nil {
		return nil
	}

	return taxonomy.nodes
}

// Len returns the number of taxa in the Taxonomy.
// It delegates the call to the Len method of the underlying nodes set.
//
// Returns:
//   - An integer representing the total count of taxa in the taxonomy.
func (taxonomy *Taxonomy) Len() int {
	taxonomy = taxonomy.OrDefault(false)

	if taxonomy == nil {
		return 0
	}

	return taxonomy.nodes.Len()
}

// AddTaxon adds a new taxon to the taxonomy with the specified parameters.
// It checks if the taxon already exists and can replace it if specified.
//
// Parameters:
//   - taxid: The identifier of the taxon to be added.
//   - parent: The identifier of the parent taxon.
//   - rank: The rank of the taxon (e.g., species, genus).
//   - isRoot: A boolean indicating if this taxon is the root of the taxonomy.
//   - replace: A boolean indicating whether to replace an existing taxon with the same taxid.
//
// Returns:
//   - A pointer to the newly created Taxon instance.
//   - An error if the taxon cannot be added (e.g., it already exists and replace is false).
func (taxonomy *Taxonomy) AddTaxon(taxid, parent string, rank string, isRoot bool, replace bool) (*Taxon, error) {
	taxonomy = taxonomy.OrDefault(true)

	parentid, perr := taxonomy.Id(parent)
	if perr != nil {
		return nil, fmt.Errorf("error in parsing parent taxid %s: %v", parent, perr)
	}

	id, err := taxonomy.Id(taxid)
	if err != nil {
		return nil, fmt.Errorf("error in parsing taxid %s: %v", taxid, err)
	}

	if !replace && taxonomy.nodes.Contains(id) {
		return nil, fmt.Errorf("trying to add taxon %s already present in the taxonomy", taxid)
	}

	prank := taxonomy.ranks.Innerize(rank)

	n := &TaxNode{id, parentid, prank, nil, nil}

	taxonomy.nodes.Insert(n)

	if isRoot {
		n.parent = n.id
		taxonomy.root = n
	}

	return &Taxon{
		Taxonomy: taxonomy,
		Node:     n,
	}, nil
}

// AddAlias adds an alias for an existing taxon in the taxonomy.
// It associates a new taxon identifier with an existing taxon identifier,
// allowing for alternative names to be used. If specified, it can replace
// an existing alias.
//
// Parameters:
//   - newtaxid: The new identifier to be added as an alias.
//   - oldtaxid: The existing identifier of the taxon to which the alias is added.
//   - replace: A boolean indicating whether to replace an existing alias with the same newtaxid.
//
// Returns:
//   - A pointer to the Taxon associated with the oldtaxid.
//   - An error if the alias cannot be added (e.g., the old taxon does not exist).
func (taxonomy *Taxonomy) AddAlias(newtaxid, oldtaxid string, replace bool) (*Taxon, error) {
	taxonomy = taxonomy.OrDefault(false)

	newid, err := taxonomy.Id(newtaxid)

	if err != nil {
		return nil, err
	}
	oldid, err := taxonomy.Id(oldtaxid)

	if err != nil {
		return nil, err
	}

	if !replace && taxonomy.nodes.Contains(newid) {
		return nil, fmt.Errorf("trying to add alias %s already present in the taxonomy", newtaxid)
	}

	t := taxonomy.nodes.Get(oldid)

	if t == nil {
		return nil, fmt.Errorf("trying to add alias %s to a taxon that does not exist", oldtaxid)
	}

	taxonomy.nodes.Alias(newid, t)

	return t, nil
}

// RankList returns a slice of strings representing the ranks of the taxa
// in the taxonomy. It retrieves the ranks from the InnerString instance
// associated with the taxonomy.
//
// Returns:
//   - A slice of strings containing the ranks of the taxa.
func (taxonomy *Taxonomy) RankList() []string {
	taxonomy = taxonomy.OrDefault(false)

	if taxonomy == nil {
		return make([]string, 0)
	}

	return taxonomy.ranks.Slice()
}

// Index returns a pointer to the map that indexes taxa by their string representation.
// This allows for quick access to taxon sets based on their identifiers.
//
// Returns:
//   - A pointer to the map that indexes taxa in the taxonomy.
func (taxonomy *Taxonomy) Index() *map[*string]*TaxonSet {
	taxonomy = taxonomy.OrDefault(false)

	if taxonomy == nil {
		return nil
	}

	return &(taxonomy.index)
}

// Name returns the name of the taxonomy.
//
// Returns:
//   - A string representing the name of the taxonomy.
func (taxonomy *Taxonomy) Name() string {
	taxonomy = taxonomy.OrDefault(true)
	return taxonomy.name
}

// Code returns the unique code representing the taxonomy.
//
// Returns:
//   - A string representing the unique code of the taxonomy.
func (taxonomy *Taxonomy) Code() string {
	taxonomy = taxonomy.OrDefault(true)
	return taxonomy.code
}

// SetRoot sets the root taxon for the taxonomy.
// It associates the provided Taxon instance as the root of the taxonomy.
//
// Parameters:
//   - root: A pointer to the Taxon instance to be set as the root.
func (taxonomy *Taxonomy) SetRoot(root *Taxon) {
	taxonomy = taxonomy.OrDefault(true)
	taxonomy.root = root.Node
}

// Root returns the root taxon of the taxonomy.
// It returns a pointer to a Taxon instance associated with the root node.
//
// Returns:
//   - A pointer to the Taxon instance representing the root of the taxonomy.
func (taxonomy *Taxonomy) Root() *Taxon {
	taxonomy = taxonomy.OrDefault(true)

	return &Taxon{
		Taxonomy: taxonomy,
		Node:     taxonomy.root,
	}
}

// HasRoot checks if the Taxonomy has a root node defined.
//
// Returns:
//   - A boolean indicating whether the Taxonomy has a root node (true) or not (false).
func (taxonomy *Taxonomy) HasRoot() bool {
	taxonomy = taxonomy.OrDefault(false)
	return taxonomy != nil && taxonomy.root != nil
}

func (taxonomy *Taxonomy) InsertPathString(path []string) (*Taxonomy, error) {
	if len(path) == 0 {
		return nil, errors.New("path is empty")
	}

	code, taxid, scientific_name, rank, err := ParseTaxonString(path[0])

	if taxonomy == nil {
		taxonomy = NewTaxonomy(code, code, obiutils.AsciiAlphaNumSet)
	}

	if err != nil {
		return nil, err
	}

	if taxonomy.Len() == 0 {

		if code != taxonomy.code {
			return nil, fmt.Errorf("cannot insert taxon %s into taxonomy %s with code %s",
				path[0], taxonomy.name, taxonomy.code)
		}

		root, err := taxonomy.AddTaxon(taxid, taxid, rank, true, true)

		if err != nil {
			return nil, err
		}
		root.SetName(scientific_name, "scientificName")
	}

	var current *Taxon
	current, err = taxonomy.Taxon(taxid)

	if err != nil {
		return nil, err
	}

	if !current.IsRoot() {
		return nil, errors.New("path does not start with a root node")
	}

	for _, id := range path[1:] {
		taxon, err := taxonomy.Taxon(id)
		if err == nil {
			if !current.SameAs(taxon.Parent()) {
				return nil, errors.New("path is not consistent with the taxonomy, parent mismatch")
			}
			current = taxon
		} else {
			current, err = current.AddChild(id, false)

			if err != nil {
				return nil, err
			}
		}
	}

	return taxonomy, nil
}
