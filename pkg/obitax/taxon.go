package obitax

import (
	"regexp"
)

type TaxNode struct {
	taxid          int
	parent         int
	pparent        *TaxNode
	rank           string
	scientificname *string
	alternatenames *map[string]*string
}

func NewTaxNode(taxid int, parent int, rank string) *TaxNode {
	n := TaxNode{taxid, parent, nil, rank, nil, nil}
	return &n
}

func (node *TaxNode) ScientificName() string {
	n := node.scientificname
	if n == nil {
		return ""
	}

	return *n
}

func (node *TaxNode) Rank() string {
	return node.rank
}

func (node *TaxNode) Taxid() int {
	return node.taxid
}

func (node *TaxNode) Parent() *TaxNode {
	return node.pparent
}

func (node *TaxNode) IsNameEqual(name string) bool {
	if *(node.scientificname) == name {
		return true
	}
	if node.alternatenames != nil {
		_, ok := (*node.alternatenames)[name]
		return ok
	}
	return false
}

func (node *TaxNode) IsNameMatching(pattern *regexp.Regexp) bool {
	if pattern.MatchString(*(node.scientificname)) {
		return true
	}
	if node.alternatenames != nil {
		for n := range *node.alternatenames {
			if pattern.MatchString(n) {
				return true
			}
		}
	}

	return false
}
