package obiseq

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

func (s *BioSequence) Taxon(taxonomy *obitax.Taxonomy) *obitax.Taxon {

	taxid := s.Taxid()
	if taxid == "NA" {
		return nil
	}

	taxon, _, _ := taxonomy.Taxon(taxid)

	return taxon
}

// SetTaxid sets the taxid for the BioSequence.
//
// Parameters:
//
//	taxid - the taxid to set.
func (s *BioSequence) SetTaxid(taxid string, rank ...string) {
	var err error
	var isAlias bool

	if taxid == "" {
		taxid = "NA"
	} else {
		taxonomy := obitax.DefaultTaxonomy()
		taxon := (*obitax.Taxon)(nil)

		if taxonomy != nil {
			taxon, isAlias, err = taxonomy.Taxon(taxid)

			if err != nil {
				logger := log.Warnf
				if obidefault.FailOnTaxonomy() {
					logger = log.Fatalf
				}
				logger("%s: Taxid: %v is unknown from taxonomy (%v)",
					s.Id(), taxid, err)
			}

			if isAlias {
				if obidefault.FailOnTaxonomy() {
					log.Fatalf("%s: Taxid: %v is an alias from taxonomy (%v) to %s",
						s.Id(), taxid, taxonomy.Name(), taxon.String())
				} else {
					if obidefault.UpdateTaxid() {
						log.Warnf("%s: Taxid: %v is updated to %s",
							s.Id(), taxid, taxon.String())
						taxid = taxon.String()
					} else {
						log.Warnf("%s: Taxid %v has to be updated to %s",
							s.Id(), taxid, taxon.String())
					}
				}

			} else {
				if taxon != nil {
					taxid = taxon.String()
				}
			}

		}
	}

	if len(rank) > 0 {
		r := rank[0]
		s.SetAttribute(r+"_taxid", taxid)
	} else {
		s.SetAttribute("taxid", taxid)
	}
}

func (s *BioSequence) SetTaxon(taxon *obitax.Taxon, rank ...string) {
	taxid := taxon.String()

	if len(rank) > 0 {
		r := rank[0]
		s.SetAttribute(r+"_taxid", taxid)
	} else {
		s.SetAttribute("taxid", taxid)
	}
}

// Taxid returns the taxonomic ID associated with the BioSequence.
//
// It retrieves the "taxid" attribute from the BioSequence's attributes map.
// If the attribute is not found, the function returns 1 as the default taxonomic ID.
// The taxid 1 corresponds to the root taxonomic level.
//
// The function returns an integer representing the taxonomic ID.
func (s *BioSequence) Taxid() (taxid string) {
	var ok bool
	if s.taxon != nil {
		taxid = s.taxon.String()
		ok = true
	} else {
		var ta interface{}
		ta, ok = s.GetAttribute("taxid")
		if ok {
			switch tv := ta.(type) {
			case string:
				taxid = tv
			case int:
				taxid = fmt.Sprintf("%d", tv)
			case float64:
				taxid = fmt.Sprintf("%d", int(tv))
			default:
				log.Fatalf("Taxid: %v is not a string or an integer (%T)", ta, ta)
			}
		}
	}

	if !ok {
		taxid = "NA"
	}

	return taxid
}

// Setting the taxon at a given rank for a given sequence.
//
// Two attributes are added to the sequence. One named by the rank name stores
// the taxid, a second named by the rank name suffixed with '_name' contains the
// Scientific name of the genus.
// If the taxon at the given rank doesn't exist for the taxonomy annotation
// of the sequence, nothing happens.
func (sequence *BioSequence) SetTaxonAtRank(taxonomy *obitax.Taxonomy, rank string) *obitax.Taxon {
	var taxonAtRank *obitax.Taxon

	taxon := sequence.Taxon(taxonomy)
	taxonAtRank = nil
	if taxon != nil {
		taxonAtRank = taxon.TaxonAtRank(rank)
		if taxonAtRank != nil {
			// log.Printf("Taxid: %d  Rank: %s --> proposed : %d (%s)", taxid, rank, taxonAtRank.taxid, *(taxonAtRank.scientificname))
			sequence.SetAttribute(rank+"_taxid", taxonAtRank.String())
			sequence.SetAttribute(rank+"_name", taxonAtRank.ScientificName())
		} else {
			sequence.SetAttribute(rank+"_taxid", "NA")
			sequence.SetAttribute(rank+"_name", "NA")
		}
	}

	return taxonAtRank
}

// Setting the species of a sequence.
func (sequence *BioSequence) SetSpecies(taxonomy *obitax.Taxonomy) *obitax.Taxon {
	return sequence.SetTaxonAtRank(taxonomy, "species")
}

// Setting the genus of a sequence.
func (sequence *BioSequence) SetGenus(taxonomy *obitax.Taxonomy) *obitax.Taxon {
	return sequence.SetTaxonAtRank(taxonomy, "genus")
}

// Setting the family of a sequence.
func (sequence *BioSequence) SetFamily(taxonomy *obitax.Taxonomy) *obitax.Taxon {
	return sequence.SetTaxonAtRank(taxonomy, "family")
}

func (sequence *BioSequence) SetPath(taxonomy *obitax.Taxonomy) []string {
	taxon := sequence.Taxon(taxonomy)
	path := taxon.Path()
	spath := make([]string, path.Len())
	lpath := path.Len() - 1

	for i := lpath; i >= 0; i-- {
		spath[lpath-i] = path.Get(i).String(taxonomy.Code())
	}

	sequence.SetAttribute("taxonomic_path", spath)

	return spath
}

func (sequence *BioSequence) Path() []string {
	path, ok := sequence.GetAttribute("taxonomic_path")

	if !ok {
		return nil
	}

	slice, err := obiutils.InterfaceToStringSlice(path)

	if err != nil {
		log.Fatalf("%s: taxonomic_path has the wrong type (%v)", sequence.Id(), err)
	}

	return slice
}

func (sequence *BioSequence) SetScientificName(taxonomy *obitax.Taxonomy) string {
	taxon := sequence.Taxon(taxonomy)
	name := taxon.ScientificName()

	sequence.SetAttribute("scientific_name", name)

	return name
}

func (sequence *BioSequence) SetTaxonomicRank(taxonomy *obitax.Taxonomy) string {
	taxon := sequence.Taxon(taxonomy)
	rank := taxon.Rank()

	sequence.SetAttribute("taxonomic_rank", rank)

	return rank
}
