package obitax

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

func (taxon *TaxNode) IsSubCladeOf(parent *TaxNode) bool {

	for taxon.taxid != parent.taxid && taxon.parent != taxon.taxid {
		taxon = taxon.pparent
	}

	return taxon.taxid == parent.taxid
}

func (taxon *TaxNode) IsBelongingSubclades(clades *TaxonSet) bool {
	_, ok := (*clades)[taxon.taxid]

	for !ok && taxon.parent != taxon.taxid {
		taxon = taxon.pparent
		_, ok = (*clades)[taxon.taxid]
	}

	return ok
}

func IsSubCladeOf(taxonomy Taxonomy, taxid int) obiseq.SequencePredicate {
	parent, err := taxonomy.Taxon(taxid)

	if err != nil {
		log.Fatalf("Cannot find taxon : %d (%v)", taxid, err)
	}

	f := func(sequence *obiseq.BioSequence) bool {
		taxon, err := taxonomy.Taxon(sequence.Taxid())
		return err == nil && taxon.IsSubCladeOf(parent)
	}

	return f
}
