package obiiter

import "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"

func (iterator *IBioSequence) ExtractTaxonomy() (taxonomy *obitax.Taxonomy, err error) {

	for iterator.Next() {
		slice := iterator.Get().Slice()

		taxonomy, err = slice.ExtractTaxonomy(taxonomy)

		if err != nil {
			return
		}
	}

	return
}
