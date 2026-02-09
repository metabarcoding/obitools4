package obikindex

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
)

// CLIBuildKmerIndex reads sequences from the iterator and builds a
// disk-based kmer index using the KmerSetGroupBuilder.
func CLIBuildKmerIndex(iterator obiiter.IBioSequence) {
	// Validate output directory
	outDir := CLIOutputDirectory()
	if outDir == "" || outDir == "-" {
		log.Fatalf("Error: --out option is required and must specify a directory path (not stdout)")
	}

	// Validate k-mer size
	k := CLIKmerSize()
	if k < 2 || k > 31 {
		log.Fatalf("Invalid k-mer size: %d (must be between 2 and 31)", k)
	}

	// Resolve minimizer size
	m := CLIMinimizerSize()

	// Validate min-occurrence
	minOcc := CLIMinOccurrence()
	if minOcc < 1 {
		log.Fatalf("Invalid min-occurrence: %d (must be >= 1)", minOcc)
	}

	log.Infof("Building kmer index: k=%d, m=%d, min-occurrence=%d", k, m, minOcc)

	// Build options
	var opts []obikmer.BuilderOption
	if minOcc > 1 {
		opts = append(opts, obikmer.WithMinFrequency(minOcc))
	}

	// Create builder (1 set, auto partitions)
	builder, err := obikmer.NewKmerSetGroupBuilder(outDir, k, m, 1, -1, opts...)
	if err != nil {
		log.Fatalf("Failed to create kmer index builder: %v", err)
	}

	// Feed sequences
	seqCount := 0
	for iterator.Next() {
		batch := iterator.Get()
		for _, seq := range batch.Slice() {
			builder.AddSequence(0, seq)
			seqCount++
		}
	}

	log.Infof("Processed %d sequences", seqCount)

	// Finalize
	ksg, err := builder.Close()
	if err != nil {
		log.Fatalf("Failed to finalize kmer index: %v", err)
	}

	// Apply metadata
	applyMetadata(ksg)

	if minOcc > 1 {
		ksg.SetAttribute("min_occurrence", minOcc)
	}

	// Persist metadata updates
	if err := ksg.SaveMetadata(); err != nil {
		log.Fatalf("Failed to save metadata: %v", err)
	}

	log.Infof("Index contains %d k-mers in %s", ksg.Len(0), outDir)
	log.Info("Done.")
}

// applyMetadata sets index-id and --set-tag metadata on a KmerSetGroup.
func applyMetadata(ksg *obikmer.KmerSetGroup) {
	if id := CLIIndexId(); id != "" {
		ksg.SetId(id)
	}

	for key, value := range CLISetTag() {
		ksg.SetAttribute(key, value)
	}
}
