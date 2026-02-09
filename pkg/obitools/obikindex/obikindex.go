package obikindex

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
)

// CLIBuildKmerIndex reads sequences from the iterator and builds a kmer index
// saved as a roaring bitmap directory.
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

	// Resolve metadata format
	format := CLIMetadataFormat()

	log.Infof("Building kmer index: k=%d, m=%d, min-occurrence=%d", k, m, minOcc)

	if minOcc <= 1 {
		// Simple KmerSet mode
		ks := obikmer.BuildKmerIndex(iterator, k, m)

		// Apply metadata
		applyKmerSetMetadata(ks)

		// Save
		log.Infof("Saving KmerSet to %s", outDir)
		if err := ks.Save(outDir, format); err != nil {
			log.Fatalf("Failed to save kmer index: %v", err)
		}
	} else {
		// FrequencyFilter mode
		ff := obikmer.BuildFrequencyFilterIndex(iterator, k, m, minOcc)

		if CLISaveFullFilter() {
			// Save the full filter (all levels)
			applyMetadataGroup(ff.KmerSetGroup)

			log.Infof("Saving full FrequencyFilter to %s", outDir)
			if err := ff.Save(outDir, format); err != nil {
				log.Fatalf("Failed to save frequency filter: %v", err)
			}
		} else {
			// Save only the filtered KmerSet (k-mers with freq >= minOcc)
			ks := ff.GetFilteredSet()
			applyKmerSetMetadata(ks)
			ks.SetAttribute("min_occurrence", minOcc)

			log.Infof("Saving filtered KmerSet (freq >= %d) to %s", minOcc, outDir)
			if err := ks.Save(outDir, format); err != nil {
				log.Fatalf("Failed to save filtered kmer index: %v", err)
			}
		}
	}

	log.Info("Done.")
}

// applyKmerSetMetadata sets index-id and --set-tag metadata on a KmerSet.
func applyKmerSetMetadata(ks *obikmer.KmerSet) {
	if id := CLIIndexId(); id != "" {
		ks.SetId(id)
	}

	for key, value := range CLISetTag() {
		ks.SetAttribute(key, value)
	}
}

// applyMetadataGroup sets index-id and --set-tag metadata on a KmerSetGroup.
func applyMetadataGroup(ksg *obikmer.KmerSetGroup) {
	if id := CLIIndexId(); id != "" {
		ksg.SetId(id)
	}

	for key, value := range CLISetTag() {
		ksg.SetAttribute(key, value)
	}
}
