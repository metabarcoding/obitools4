package obitax

import (
	"fmt"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

// TaxNode represents a single taxon in a taxonomy.
// It holds information about the taxon's identifier, parent taxon, rank,
// scientific name, and alternate names.
//
// Fields:
//   - id: A pointer to the unique identifier of the taxon of type T.
//   - parent: A pointer to the identifier of the parent taxon of type T.
//   - rank: A pointer to the rank of the taxon (e.g., species, genus).
//   - scientificname: A pointer to a string representing the scientific name of the taxon.
//   - alternatenames: A pointer to a map of alternate names for the taxon, where the key is
//     a pointer to a string representing the class name and the value is a pointer to a string
//     representing the name.
type TaxNode struct {
	id             *string
	parent         *string
	rank           *string
	scientificname *string
	alternatenames *map[*string]*string
}

// String returns a string representation of the TaxNode, including the taxonomy code,
// the node ID, and the scientific name. The output format is "taxonomyCode:id [scientificName]".
//
// Parameters:
//   - taxonomyCode: A string representing the code of the taxonomy to which the node belongs.
//
// Returns:
//   - A formatted string representing the TaxNode in the form "taxonomyCode:id [scientificName]".
func (node *TaxNode) String(taxonomyCode string) string {
	if node.HasScientificName() {
		return fmt.Sprintf("%s:%v [%s]",
			taxonomyCode,
			*node.id,
			node.ScientificName())
	}

	return fmt.Sprintf("%s:%v",
		taxonomyCode,
		*node.id)

}

// Id returns the unique identifier of the TaxNode.
// It retrieves the identifier of type T associated with the taxon node.
//
// Returns:
//   - A pointer to the unique identifier of the taxon node of type T.
func (node *TaxNode) Id() *string {
	return node.id
}

// ParentId returns the identifier of the parent taxon of the TaxNode.
// It retrieves the parent identifier of type T associated with the taxon node.
//
// Returns:
//   - A pointer to the identifier of the parent taxon of type T.
func (node *TaxNode) ParentId() *string {
	return node.parent
}

func (node *TaxNode) HasScientificName() bool {
	return node != nil && node.scientificname != nil
}

// ScientificName returns the scientific name of the TaxNode.
// It dereferences the pointer to the scientific name string associated with the taxon node.
//
// Returns:
//   - The scientific name of the taxon as a string.
//   - If the scientific name is nil, it returns "NA".
//   - Note: This method assumes that the TaxNode itself is not nil; if it may be nil,
//     additional error handling should be implemented.
func (node *TaxNode) ScientificName() string {
	if node == nil {
		return "NA"
	}
	if node.scientificname == nil {
		return "NA"
	}
	return *node.scientificname
}

// Name retrieves the name of the TaxNode based on the specified class.
// If the class is "scientific name", it returns the scientific name of the taxon.
// If the class corresponds to an alternate name, it retrieves that name from the alternatenames map.
// If the class is not recognized or if no alternate names exist, it returns an empty string.
//
// Parameters:
//   - class: A pointer to a string representing the class of name to retrieve
//     (e.g., "scientific name" or an alternate name class).
//
// Returns:
//   - The name of the taxon as a string. If the class is not recognized or if no name is available,
//     an empty string is returned.
func (node *TaxNode) Name(class *string) string {
	if *class == "scientific name" {
		return *node.scientificname
	}

	if node.alternatenames == nil {
		return ""
	}

	if val, ok := (*node.alternatenames)[class]; ok {
		if val != nil {
			return *val
		}
	}

	return ""
}

// SetName sets the name of the TaxNode based on the specified class.
// If the class is "scientific name", it updates the scientific name of the taxon.
// If the class corresponds to an alternate name, it adds or updates that name in the alternatenames map.
//
// Parameters:
//   - name: A pointer to a string representing the name to set.
//   - class: A pointer to a string representing the class of name to set
//     (e.g., "scientific name" or an alternate name class).
func (node *TaxNode) SetName(name, class *string) {
	if node == nil {
		log.Panic("Cannot set name of nil TaxNode")
	}

	if *class == "scientific name" {
		node.scientificname = name
		return
	}

	if node.alternatenames == nil {
		node.alternatenames = &map[*string]*string{}
	}

	(*node.alternatenames)[class] = name
}

// Rank returns the rank of the TaxNode.
// It retrieves the rank associated with the taxon node, which indicates its level in the taxonomy hierarchy.
//
// Returns:
//   - The rank of the taxon as a string (e.g., species, genus, family).
func (node *TaxNode) Rank() string {
	return *node.rank
}

// IsNameEqual checks if the provided name matches the scientific name or any alternate names
// associated with the TaxNode. It returns true if there is a match; otherwise, it returns false.
//
// Parameters:
//   - name: A string representing the name to compare against the scientific name and alternate names.
//
// Returns:
//   - A boolean indicating whether the provided name is equal to the scientific name or exists
//     as an alternate name for the taxon.
func (node *TaxNode) IsNameEqual(name string, ignoreCase bool) bool {
	if node == nil {
		return false
	}

	if *(node.scientificname) == name || (ignoreCase && strings.EqualFold(*(node.scientificname), name)) {
		return true
	}
	if node.alternatenames != nil {
		for _, n := range *node.alternatenames {
			if n != nil && (ignoreCase && strings.EqualFold(*n, name)) {
				return true
			}
		}
	}
	return false
}

// IsNameMatching checks if the scientific name or any alternate names of the TaxNode match
// the provided regular expression pattern. It returns true if there is a match; otherwise, it returns false.
//
// Parameters:
//   - pattern: A pointer to a regexp.Regexp object representing the pattern to match against
//     the scientific name and alternate names.
//
// Returns:
//   - A boolean indicating whether the scientific name or any alternate names match the
//     provided regular expression pattern.
func (node *TaxNode) IsNameMatching(pattern *regexp.Regexp) bool {
	if node == nil {
		return false
	}

	if node.scientificname != nil && pattern.MatchString(*(node.scientificname)) {
		return true
	}

	if node.alternatenames != nil {
		for _, n := range *node.alternatenames {
			if n != nil && pattern.MatchString(*n) {
				return true
			}
		}
	}

	return false
}
