package obitax

import (
	"fmt"
)

type TaxName struct {
	name      *string
	nameclass *string
}

type Taxonomy struct {
	nodes *TaxonSet
	alias map[int]*TaxNode
	index map[string]*TaxonSet
}

func NewTaxonomy() *Taxonomy {
	set := make(TaxonSet)
	taxonomy := Taxonomy{
		nodes: &set,
		alias: make(TaxonSet),
		index: make(map[string]*TaxonSet)}
	return &taxonomy
}

func (taxonomy *Taxonomy) TaxonSet() *TaxonSet {
	return taxonomy.nodes
}

func (taxonomy *Taxonomy) Alias() *map[int]*TaxNode {
	return &(taxonomy.alias)
}

func (taxonomy *Taxonomy) Index() *map[string]*TaxonSet {
	return &(taxonomy.index)
}

func (taxonomy *Taxonomy) Length() int {
	return len(*taxonomy.nodes)
}

func (taxonomy *Taxonomy) AddNewTaxa(taxid, parent int, rank string, replace bool, init bool) (*TaxNode, error) {
	if !replace {
		_, ok := (*taxonomy.nodes)[taxid]
		if ok {
			return nil, fmt.Errorf("trying to add taxoon %d already present in the taxonomy", taxid)
		}
	}

	n := NewTaxNode(taxid, parent, rank)
	(*taxonomy.nodes)[taxid] = n

	return n, nil
}

func (taxonomy *Taxonomy) Taxon(taxid int) (*TaxNode, error) {
	t, ok := (*taxonomy.nodes)[taxid]

	if !ok {
		a, aok := taxonomy.alias[taxid]
		if !aok {
			return nil, fmt.Errorf("Taxid %d is not part of the taxonomy", taxid)
		}
		t = a
	}
	return t, nil
}

func (taxonomy *Taxonomy) AddNewName(taxid int, name, nameclass *string) error {
	node, node_err := taxonomy.Taxon(taxid)
	if node_err != nil {
		return node_err
	}

	if *nameclass == "scientific name" {
		node.scientificname = name
	} else {
		names := node.alternatenames
		if names == nil {
			n := make(map[string]*string)
			names = &n
			node.alternatenames = names
		} else {
			(*names)[*name] = nameclass
		}
	}

	i, ok := taxonomy.index[*name]
	if !ok {
		tnm := make(TaxonSet)
		i = &tnm
		taxonomy.index[*name] = i
	}
	(*i)[taxid] = node

	return nil
}

func (taxonomy *Taxonomy) ReindexParent() error {
	var ok bool
	for _, taxon := range *taxonomy.nodes {
		taxon.pparent, ok = (*taxonomy.nodes)[taxon.parent]
		if !ok {
			return fmt.Errorf("Parent %d of taxon %d is not defined in taxonomy",
				taxon.taxid,
				taxon.parent)
		}
	}

	return nil
}

func MakeTaxName(name, nameclass *string) *TaxName {
	tn := TaxName{name, nameclass}
	return &tn
}

func (taxonomy *Taxonomy) AddNewAlias(newtaxid, oldtaxid int) error {
	n, node_err := taxonomy.Taxon(newtaxid)
	if node_err != nil {
		return node_err
	}

	taxonomy.alias[oldtaxid] = n

	return nil
}
