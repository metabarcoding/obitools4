package obitax

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	log "github.com/sirupsen/logrus"
)

// Setting the taxon at a given rank for a given sequence.
//
// Two attributes are added to the sequence. One named by the rank name stores
// the taxid, a second named by the rank name suffixed with '_name' contains the
// Scientific name of the genus.
// If the taxon at the given rank doesn't exist for the taxonomy annotation
// of the sequence, nothing happens.
func (taxonomy *Taxonomy) SetTaxonAtRank(sequence *obiseq.BioSequence, rank string) *TaxNode {
	var taxonAtRank *TaxNode

	taxid := sequence.Taxid()
	taxon, err := taxonomy.Taxon(taxid)
	taxonAtRank = nil
	if err == nil {
		taxonAtRank = taxon.TaxonAtRank(rank)
		if taxonAtRank != nil {
			// log.Printf("Taxid: %d  Rank: %s --> proposed : %d (%s)", taxid, rank, taxonAtRank.taxid, *(taxonAtRank.scientificname))
			sequence.SetAttribute(rank+"_taxid", taxonAtRank.taxid)
			sequence.SetAttribute(rank+"_name", *taxonAtRank.scientificname)
		} else {
			sequence.SetAttribute(rank+"_taxid", -1)
			sequence.SetAttribute(rank+"_name", "NA")
		}
	}

	return taxonAtRank
}

// Setting the species of a sequence.
func (taxonomy *Taxonomy) SetSpecies(sequence *obiseq.BioSequence) *TaxNode {
	return taxonomy.SetTaxonAtRank(sequence, "species")
}

// Setting the genus of a sequence.
func (taxonomy *Taxonomy) SetGenus(sequence *obiseq.BioSequence) *TaxNode {
	return taxonomy.SetTaxonAtRank(sequence, "genus")
}

// Setting the family of a sequence.
func (taxonomy *Taxonomy) SetFamily(sequence *obiseq.BioSequence) *TaxNode {
	return taxonomy.SetTaxonAtRank(sequence, "family")
}

func (taxonomy *Taxonomy) SetPath(sequence *obiseq.BioSequence) string {
	taxid, err := taxonomy.Taxon(sequence.Taxid())

	if err != nil {
		log.Fatalf("Taxid %d not defined in the current taxonomy", sequence.Taxid())
	}

	path, err := taxid.Path()

	if err != nil {
		log.Fatalf("Taxonomy index error: %v", err)
	}

	tpath := path.String()
	sequence.SetAttribute("taxonomic_path", tpath)

	return tpath
}
