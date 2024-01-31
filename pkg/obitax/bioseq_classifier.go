package obitax

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	log "github.com/sirupsen/logrus"
)

// TaxonomyClassifier is a function that creates a new instance of the BioSequenceClassifier
// for taxonomic classification based on a given taxonomic rank, taxonomy, and abort flag.
//
// Parameters:
// - taxonomicRank: the taxonomic rank to classify the sequences at.
// - taxonomy: the taxonomy object used for classification.
// - abortOnMissing: a flag indicating whether to abort if a taxon is missing in the taxonomy.
//
// Return:
// - *obiseq.BioSequenceClassifier: the new instance of the BioSequenceClassifier.
func TaxonomyClassifier(taxonomicRank string,
	taxonomy *Taxonomy,
	abortOnMissing bool) *obiseq.BioSequenceClassifier {

	code := func(sequence *obiseq.BioSequence) int {
		taxid := sequence.Taxid()
		taxon, err := taxonomy.Taxon(taxid)
		if err == nil {
			taxon = taxon.TaxonAtRank(taxonomicRank)
		} else {
			taxon = nil
		}
		if taxon == nil {
			if abortOnMissing {
				if err != nil {
					log.Fatalf("Taxid %d not found in taxonomy", taxid)
				} else {
					log.Fatalf("Taxon at rank %s not found in taxonomy for taxid %d", taxonomicRank, taxid)
				}

			}

			return 0
		}
		return taxon.Taxid()
	}

	value := func(k int) string {
		taxon, _ := taxonomy.Taxon(k)
		return taxon.ScientificName()
	}

	reset := func() {
	}

	clone := func() *obiseq.BioSequenceClassifier {
		return TaxonomyClassifier(taxonomicRank, taxonomy, abortOnMissing)
	}

	c := obiseq.BioSequenceClassifier{
		Code:  code,
		Value: value,
		Reset: reset,
		Clone: clone,
		Type:  "TaxonomyClassifier"}
	return &c
}
