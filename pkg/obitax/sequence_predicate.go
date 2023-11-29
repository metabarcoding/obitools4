package obitax

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

func (taxonomy *Taxonomy) IsAValidTaxon(withAutoCorrection ...bool) obiseq.SequencePredicate {
	deprecatedTaxidsWarning := make(map[int]bool)

	autocorrection := false
	if len(withAutoCorrection) > 0 {
		autocorrection = withAutoCorrection[0]
	}

	f := func(sequence *obiseq.BioSequence) bool {
		taxid := sequence.Taxid()
		taxon, err := taxonomy.Taxon(taxid)

		if err == nil && taxon.taxid != taxid {
			if autocorrection {
				sequence.SetTaxid(taxon.taxid)
				log.Printf("Sequence %s : Taxid %d updated with %d",
					sequence.Id(),
					taxid,
					taxon.taxid)
			} else {
				if _, ok := deprecatedTaxidsWarning[taxid]; !ok {
					deprecatedTaxidsWarning[taxid] = true
					log.Printf("Taxid %d is deprecated and must be replaced by %d", taxid, taxon.taxid)
				}
			}
		}

		return err == nil
	}

	return f
}

// A function that takes a taxonomy and a taxid as arguments and returns a function that takes a
// pointer to a BioSequence as an argument and returns a boolean.
func (taxonomy *Taxonomy) IsSubCladeOf(taxid int) obiseq.SequencePredicate {
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

func (taxonomy *Taxonomy) HasRequiredRank(rank string) obiseq.SequencePredicate {

	if !obiutils.Contains(taxonomy.RankList(), rank) {
		log.Fatalf("%s is not a valid rank (allowed ranks are %v)",
			rank,
			taxonomy.RankList())
	}

	f := func(sequence *obiseq.BioSequence) bool {
		taxon, err := taxonomy.Taxon(sequence.Taxid())
		return err == nil && taxon.HasRankDefined(rank)
	}

	return f
}
