package obiiter

import "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"

// ExtractTaxonomy iterates over each slice of the IBioSequence and extracts the taxonomy from it using the ExtractTaxonomy method of the slice.
// If the seqAsTaxa parameter is true, then the sequence itself will be treated as a single taxon. Otherwise, each element in the slice will be treated separately.
// The function returns an error if any of the ExtractTaxonomy calls fail and nil otherwise.
func (iterator *IBioSequence) ExtractTaxonomy(seqAsTaxa bool) (taxonomy *obitax.Taxonomy, err error) {
	// Iterate over each slice in the iterator
	for iterator.Next() {
		// Get the current slice
		slice := iterator.Get().Slice()

		// Try to extract taxonomy from the slice
		taxonomy, err = slice.ExtractTaxonomy(taxonomy, seqAsTaxa)

		// If an error occurred during extraction, return it immediately
		if err != nil {
			return
		}
	}

	// Return the extracted taxonomy and no error if all slices were successfully processed
	return
}
