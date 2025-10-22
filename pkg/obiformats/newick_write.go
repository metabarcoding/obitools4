package obiformats

import (
	"fmt"
	"io"
	"os"
	"strings"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	log "github.com/sirupsen/logrus"
)

// Tree corresponds to any value representable in a Newick format. Each
// tree value corresponds to a single node.
type Tree struct {
	// All children of this node, which may be empty.
	Children []*Tree

	// The label of this node. If it's empty, then this node does
	// not have a name.
	TaxNode *obitax.TaxNode

	// The branch length of this node corresponding to the distance between
	// it and its parent node. If it's `nil`, then no distance exists.
	Length *float64
}

func (tree *Tree) Newick(level int, taxid, scientific_name, rank bool) string {
	var buffer strings.Builder

	buffer.WriteString(strings.Repeat(" ", level))

	if len(tree.Children) > 0 {
		buffer.WriteString("(\n")
		for i, c := range tree.Children {
			if i > 0 {
				buffer.WriteString(",\n")
			}
			buffer.WriteString(c.Newick(level+1, taxid, scientific_name, rank))
		}
		buffer.WriteByte('\n')
		buffer.WriteString(strings.Repeat(" ", level))
		buffer.WriteByte(')')
	}
	if scientific_name || taxid || rank {
		buffer.WriteByte('\'')
	}
	if scientific_name {
		sn := strings.ReplaceAll(tree.TaxNode.ScientificName(), ",", "")
		buffer.WriteString(sn)
	}
	if taxid || rank {
		if scientific_name {
			buffer.WriteByte(' ')
		}
		// buffer.WriteByte('-')
		if taxid {
			buffer.WriteString(*tree.TaxNode.Id())
			if rank {
				buffer.WriteByte('@')
			}
		}
		if rank {
			buffer.WriteString(tree.TaxNode.Rank())
		}
		//buffer.WriteByte('-')
	}
	if scientific_name || taxid || rank {
		buffer.WriteByte('\'')
	}

	if tree.Length != nil {
		buffer.WriteString(fmt.Sprintf(":%f", *tree.Length))
	}

	if level == 0 {
		buffer.WriteString(";\n")
	}
	return buffer.String()
}

func Newick(taxa *obitax.TaxonSet, taxid, scientific_name, rank bool) string {
	if taxa == nil {
		return ""
	}

	root := taxa.Sort().Get(0)
	tree, err := taxa.AsPhyloTree(root)

	if err != nil {
		log.Fatalf("Cannot build taxonomy tree: %v", err)
	}

	return tree.Newick(0)
}

func WriteNewick(iterator *obitax.ITaxon,
	file io.WriteCloser,
	options ...WithOption) (*obitax.ITaxon, error) {
	newiterator := obitax.NewITaxon()

	var taxonomy *obitax.Taxonomy
	var taxa *obitax.TaxonSet

	opt := MakeOptions(options)

	file, _ = obiutils.CompressStream(file, opt.CompressedFile(), opt.CloseFile())
	obiutils.RegisterAPipe()

	go func() {
		for iterator.Next() {
			taxon := iterator.Get()
			if taxonomy == nil {
				taxonomy = taxon.Taxonomy
				taxa = taxonomy.NewTaxonSet()
			}
			if taxon.Taxonomy != taxonomy {
				log.Fatal("Newick writer cannot deal with multi-taxonomy iterator")
			}
			taxa.InsertTaxon(taxon)
			newiterator.Push(taxon)
		}

		newick := Newick(taxa, opt.WithTaxid(), opt.WithScientificName(), opt.WithRank())
		file.Write(obiutils.UnsafeBytes(newick))

		newiterator.Close()
		if opt.CloseFile() {
			file.Close()
		}

		obiutils.UnregisterPipe()
		log.Debugf("Writing newick file done")
	}()

	return newiterator, nil
}

func WriteNewickToFile(iterator *obitax.ITaxon,
	filename string,
	options ...WithOption) (*obitax.ITaxon, error) {

	flags := os.O_WRONLY | os.O_CREATE
	flags |= os.O_TRUNC

	file, err := os.OpenFile(filename, flags, 0660)

	if err != nil {
		log.Fatalf("open file error: %v", err)
		return nil, err
	}

	options = append(options, OptionCloseFile())

	iterator, err = WriteNewick(iterator, file, options...)

	return iterator, err
}

func WriteNewickToStdout(iterator *obitax.ITaxon,
	options ...WithOption) (*obitax.ITaxon, error) {
	options = append(options, OptionCloseFile())
	return WriteNewick(iterator, os.Stdout, options...)
}
