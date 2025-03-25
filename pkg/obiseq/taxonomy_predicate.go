package obiseq

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obilog"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

func IsAValidTaxon(taxonomy *obitax.Taxonomy, withAutoCorrection ...bool) SequencePredicate {
	// deprecatedTaxidsWarning := make(map[string]bool)

	autocorrection := false
	if len(withAutoCorrection) > 0 {
		autocorrection = withAutoCorrection[0]
	}

	f := func(sequence *BioSequence) bool {
		taxon := sequence.Taxon(taxonomy)

		if taxon != nil {
			taxid := sequence.Taxid()
			ttaxid := taxon.String()
			if taxid != ttaxid {
				if autocorrection {
					sequence.SetTaxid(ttaxid)
					log.Printf(
						"Sequence %s : Taxid %s updated with %s",
						sequence.Id(),
						taxid,
						ttaxid,
					)
				} //  else {
				// 	if _, ok := deprecatedTaxidsWarning[taxid]; !ok {
				// 		deprecatedTaxidsWarning[taxid] = true
				// 		log.Printf("Taxid %d is deprecated and must be replaced by %d", taxid, taxon.taxid)
				// 	}
				// }
			}
		}

		return taxon != nil
	}

	return f
}

// A function that takes a taxonomy and a taxid as arguments and returns a function that takes a
// pointer to a BioSequence as an argument and returns a boolean.
func IsSubCladeOf(taxonomy *obitax.Taxonomy, parent *obitax.Taxon) SequencePredicate {

	f := func(sequence *BioSequence) bool {
		taxon := sequence.Taxon(taxonomy)
		return taxon != nil && taxon.IsSubCladeOf(parent)
	}

	return f
}

func IsSubCladeOfSlot(taxonomy *obitax.Taxonomy, key string) SequencePredicate {

	f := func(sequence *BioSequence) bool {
		val, ok := sequence.GetStringAttribute(key)

		if ok {
			parent, _, err := taxonomy.Taxon(val)

			if err != nil {
				obilog.Warnf("%s: %s is unkown from the taxonomy (%v)", sequence.Id(), val, err)
			}

			taxon := sequence.Taxon(taxonomy)
			return parent != nil && taxon != nil && taxon.IsSubCladeOf(parent)
		}

		return false
	}

	return f
}

func HasRequiredRank(taxonomy *obitax.Taxonomy, rank string) SequencePredicate {

	if !obiutils.Contains(taxonomy.RankList(), rank) {
		log.Fatalf("%s is not a valid rank (allowed ranks are %v)",
			rank,
			taxonomy.RankList())
	}

	f := func(sequence *BioSequence) bool {
		taxon := sequence.Taxon(taxonomy)
		return taxon != nil && taxon.HasRankDefined(rank)
	}

	return f
}
