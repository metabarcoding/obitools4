package obitax

import (
	"fmt"
	"regexp"

	log "github.com/sirupsen/logrus"
)

// Taxonomy represents a hierarchical classification of taxa.
// It holds information about the taxonomy's name, code, ranks, nodes, root node, aliases, and an index.
// The generic type T is used to specify the type of taxon identifiers.
//
// Fields:
//   - name: The name of the taxonomy.
//   - code: A unique code representing the taxonomy.
//   - ranks: A pointer to an InnerString instance that holds the ranks of the taxa.
//   - nodes: A pointer to a TaxonSet containing all the nodes (taxa) in the taxonomy.
//   - root: A pointer to the root TaxNode of the taxonomy.
//   - index: A map that indexes taxa by their string representation for quick access.
type Taxonomy struct {
	name        string
	code        string
	ranks       *InnerString
	nameclasses *InnerString
	nodes       *TaxonSet
	root        *TaxNode
	matcher     *regexp.Regexp
	index       map[string]*TaxonSet
}

// NewTaxonomy creates and initializes a new Taxonomy instance with the specified name and code.
// It sets up the necessary internal structures, including ranks, nodes, aliases, and an index.
//
// Parameters:
//   - name: The name of the taxonomy to be created.
//   - code: A unique code representing the taxonomy.
//
// Returns:
//   - A pointer to the newly created Taxonomy instance.
func NewTaxonomy(name, code, codeCharacters string) *Taxonomy {
	set := make(map[string]*TaxNode)

	// 	codeCharacters := "[[:alnum:]]" // [[:digit:]]

	matcher := regexp.MustCompile(fmt.Sprintf("^[[:blank:]]*(%s:)?(%s+)", code, codeCharacters))

	taxonomy := &Taxonomy{
		name:        name,
		code:        code,
		ranks:       NewInnerString(),
		nameclasses: NewInnerString(),
		nodes:       &TaxonSet{set: set},
		root:        nil,
		matcher:     matcher,
		index:       make(map[string]*TaxonSet),
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
//   - The taxon identifier of type T corresponding to the provided taxid.
//   - An error if the taxid is not valid or cannot be converted.
func (taxonomy *Taxonomy) Id(taxid string) (string, error) {
	matches := taxonomy.matcher.FindStringSubmatch(taxid)

	if matches == nil {
		return "", fmt.Errorf("Taxid %s is not a valid taxid", taxid)
	}

	return matches[2], nil
}

// TaxidSting retrieves the string representation of a taxon node identified by the given ID.
// It looks up the node in the taxonomy and returns its formatted string representation
// along with the taxonomy code. If the node does not exist, it returns an error.
//
// Parameters:
//   - id: The identifier of the taxon node to retrieve.
//
// Returns:
//   - A string representing the taxon node in the format "taxonomyCode:id [scientificName]",
//     or an error if the taxon node with the specified ID does not exist in the taxonomy.
func (taxonomy *Taxonomy) TaxidSting(id string) (string, error) {
	node := taxonomy.nodes.Get(id)
	if node == nil {
		return "", fmt.Errorf("Taxid %d is part of the taxonomy", id)
	}
	return node.String(taxonomy.code), nil
}

// Taxon retrieves the Taxon associated with the given taxid string.
// It first converts the taxid to its corresponding identifier using the Id method.
// If the taxon is not found, it logs a fatal error and terminates the program.
//
// Parameters:
//   - taxid: A string representation of the taxon identifier to be retrieved.
//
// Returns:
//   - A pointer to the Taxon[T] instance associated with the provided taxid.
//   - If the taxid is unknown, the method will log a fatal error.
func (taxonomy *Taxonomy) Taxon(taxid string) *Taxon {
	id, err := taxonomy.Id(taxid)

	if err != nil {
		log.Fatalf("Taxid %s is not a valid taxid", taxid)
	}

	node := taxonomy.nodes.Get(id)

	if node == nil {
		log.Fatalf("Taxid %s is an unknown taxid", taxid)
	}

	return &Taxon{
		Taxonomy: taxonomy,
		Node:     node,
	}
}

// TaxonSet returns the set of taxon nodes contained within the Taxonomy.
// It provides access to the underlying collection of taxon nodes for further operations.
//
// Returns:
//   - A pointer to the TaxonSet[T] representing the collection of taxon nodes in the taxonomy.
func (taxonomy *Taxonomy) TaxonSet() *TaxonSet {
	return taxonomy.nodes
}

// Len returns the number of taxa in the Taxonomy.
// It delegates the call to the Len method of the underlying nodes set.
//
// Returns:
//   - An integer representing the total count of taxa in the taxonomy.
func (taxonomy *Taxonomy) Len() int {
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
//   - A pointer to the newly created Taxon[T] instance.
//   - An error if the taxon cannot be added (e.g., it already exists and replace is false).
func (taxonomy *Taxonomy) AddTaxon(taxid, parent string, rank string, isRoot bool, replace bool) (*Taxon, error) {
	if !replace && taxonomy.nodes.Contains(taxid) {
		return nil, fmt.Errorf("trying to add taxon %d already present in the taxonomy", taxid)
	}

	rank = taxonomy.ranks.Innerize(rank)

	n := &TaxNode{taxid, parent, rank, nil, nil}

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

func (taxonomy *Taxonomy) AddAlias(newtaxid, oldtaxid string, replace bool) (*Taxon, error) {
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

	n := taxonomy.nodes.Get(oldid)

	if n == nil {
		return nil, fmt.Errorf("trying to add alias %s to a taxon that does not exist", oldtaxid)
	}

	taxonomy.nodes.Alias(newid, n)

	return &Taxon{
		Taxonomy: taxonomy,
		Node:     n,
	}, nil
}

// RankList returns a slice of strings representing the ranks of the taxa
// in the taxonomy. It retrieves the ranks from the InnerString instance
// associated with the taxonomy.
//
// Returns:
//   - A slice of strings containing the ranks of the taxa.
func (taxonomy *Taxonomy) RankList() []string {
	return taxonomy.ranks.Slice()
}

// func (taxonomy *Taxonomy) Taxon(taxid int) (*TaxNode, error) {
// 	t, ok := (*taxonomy.nodes)[taxid]

// 	if !ok {
// 		a, aok := taxonomy.alias[taxid]
// 		if !aok {
// 			return nil, fmt.Errorf("Taxid %d is not part of the taxonomy", taxid)
// 		}
// 		t = a
// 	}
// 	return t, nil
// }

func (taxonomy *Taxonomy) Index() *map[string]*TaxonSet {
	return &(taxonomy.index)
}
