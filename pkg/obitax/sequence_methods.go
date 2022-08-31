package obitax

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

// Setting the taxon at a given rank for a given sequence.
//
// Two attributes are added to the sequence. One named by the rank name stores
// the taxid, a second named by the rank name suffixed with '_name' contains the
// Scientific name of the genus.
// If the taxon at the given rank doesn't exist for the taxonomy annotation
// of the sequence, nothing happens.
func (taxonomy *Taxonomy) SetTaxonAtRank(sequence *obiseq.BioSequence, rank string) *TaxNode {
	taxid := sequence.Taxid()
	taxon, err := taxonomy.Taxon(taxid)
	taxonAtRank := taxon.TaxonAtRank(rank)

	if err == nil && taxonAtRank != nil {
		sequence.SetAttribute(rank, taxonAtRank.taxid)
		sequence.SetAttribute(rank+"_name", taxonAtRank.scientificname)
	}

	return taxonAtRank
}

func (taxonomy *Taxonomy) SetSpecies(sequence *obiseq.BioSequence) *TaxNode {
	return taxonomy.SetTaxonAtRank(sequence, "species")
}

func (taxonomy *Taxonomy) SetGenus(sequence *obiseq.BioSequence) *TaxNode {
	return taxonomy.SetTaxonAtRank(sequence, "genus")
}

func (taxonomy *Taxonomy) SetFamily(sequence *obiseq.BioSequence) *TaxNode {
	return taxonomy.SetTaxonAtRank(sequence, "family")
}
