package obiseq

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
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
	taxonomy *obitax.Taxonomy,
	abortOnMissing bool) *BioSequenceClassifier {

	keys := make(map[*obitax.TaxNode]int)
	codes := make([]*obitax.TaxNode, 1)
	codes[0] = nil
	keys[nil] = 0

	code := func(sequence *BioSequence) int {
		taxon := sequence.Taxon(taxonomy)
		if taxon != nil {
			ttaxon := taxon.TaxonAtRank(taxonomicRank)
			if abortOnMissing && ttaxon == nil {
				log.Fatalf("Taxon at rank %s not found in taxonomy for taxid %d", taxonomicRank, taxon.String())
			}
		} else {
			if abortOnMissing {
				log.Fatalf("Sequence %s: Taxid %s not found in taxonomy",
					sequence.Id(),
					sequence.Taxid())
			}
			taxon = nil
		}

		k, ok := keys[taxon.Node]

		if ok {
			return k
		}

		k = len(codes)
		keys[taxon.Node] = k
		codes = append(codes, taxon.Node)

		return k
	}

	value := func(k int) string {
		taxon := codes[k]
		return taxon.ScientificName()
	}

	reset := func() {
		keys = make(map[*obitax.TaxNode]int)
		codes = make([]*obitax.TaxNode, 1)
		codes[0] = nil
		keys[nil] = 0
	}

	clone := func() *BioSequenceClassifier {
		return TaxonomyClassifier(taxonomicRank, taxonomy, abortOnMissing)
	}

	c := BioSequenceClassifier{
		Code:  code,
		Value: value,
		Reset: reset,
		Clone: clone,
		Type:  "TaxonomyClassifier"}
	return &c
}
