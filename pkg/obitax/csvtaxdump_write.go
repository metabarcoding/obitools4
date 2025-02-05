package obitax

import (
	"strings"

	"github.com/TuftsBCB/io/newick"
)

func (taxonomy *Taxonomy) Newick() string {
	if taxonomy == nil {
		return ""
	}

	iterator := taxonomy.AsTaxonSet().Sort().Iterator()

	nodes := make(map[*string]*newick.Tree, taxonomy.Len())
	trees := make([]*newick.Tree, 0)

	for iterator.Next() {
		taxon := iterator.Get()
		tree := &newick.Tree{Label: taxon.String()}
		nodes[taxon.Node.id] = tree
		if parent, ok := nodes[taxon.Parent().Node.id]; ok {
			parent.Children = append(parent.Children, *tree)
		} else {
			trees = append(trees, tree)
		}
	}

	rep := strings.Builder{}

	for _, tree := range trees {
		rep.WriteString(tree.String())
		rep.WriteString("\n")
	}

	return rep.String()
}
