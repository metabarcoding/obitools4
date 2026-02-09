package obisuperkmer

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
)

// CLIExtractSuperKmers extracts super k-mers from an iterator of BioSequences.
//
// This function takes an iterator of BioSequence objects, extracts super k-mers
// from each sequence using the k-mer and minimizer sizes specified by CLI options,
// and returns a new iterator yielding the extracted super k-mers as BioSequence objects.
//
// Each super k-mer is a maximal subsequence where all consecutive k-mers share
// the same minimizer. The resulting BioSequences contain metadata including:
// - minimizer_value: the canonical minimizer value
// - minimizer_seq: the DNA sequence of the minimizer
// - k: the k-mer size used
// - m: the minimizer size used
// - start: starting position in the original sequence
// - end: ending position in the original sequence
// - parent_id: ID of the parent sequence
//
// Parameters:
// - iterator: an iterator yielding BioSequence objects to process.
//
// Returns:
// - An iterator yielding BioSequence objects representing super k-mers.
func CLIExtractSuperKmers(iterator obiiter.IBioSequence) obiiter.IBioSequence {
	// Get k-mer and minimizer sizes from CLI options
	k := CLIKmerSize()
	m := CLIMinimizerSize()

	// Validate parameters
	if m < 1 || m >= k {
		log.Fatalf("Invalid parameters: minimizer size (%d) must be between 1 and k-1 (%d)", m, k-1)
	}

	if k < 2 || k > 31 {
		log.Fatalf("Invalid k-mer size: %d (must be between 2 and 31)", k)
	}

	log.Printf("Extracting super k-mers with k=%d, m=%d", k, m)

	// Create the worker for super k-mer extraction
	worker := obikmer.SuperKmerWorker(k, m)

	// Apply the worker to the iterator with parallel processing
	newIter := iterator.MakeIWorker(
		worker,
		false, // don't merge results
		obidefault.ParallelWorkers(),
	)

	return newIter
}
