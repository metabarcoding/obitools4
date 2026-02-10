package obik

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

func runIndex(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	outDir := CLIOutputDirectory()
	if outDir == "" || outDir == "-" {
		return fmt.Errorf("--out option is required and must specify a directory path")
	}

	k := CLIKmerSize()
	if k < 2 || k > 31 {
		return fmt.Errorf("invalid k-mer size: %d (must be between 2 and 31)", k)
	}

	m := CLIMinimizerSize()

	minOcc := CLIMinOccurrence()
	if minOcc < 1 {
		return fmt.Errorf("invalid min-occurrence: %d (must be >= 1)", minOcc)
	}

	maxOcc := CLIMaxOccurrence()

	entropyThreshold := CLIIndexEntropyThreshold()
	entropySize := CLIIndexEntropySize()

	// Build options
	var opts []obikmer.BuilderOption
	if minOcc > 1 {
		opts = append(opts, obikmer.WithMinFrequency(minOcc))
	}
	if maxOcc > 0 {
		opts = append(opts, obikmer.WithMaxFrequency(maxOcc))
	}
	if topN := CLISaveFreqKmer(); topN > 0 {
		opts = append(opts, obikmer.WithSaveFreqKmers(topN))
	}
	if entropyThreshold > 0 {
		opts = append(opts, obikmer.WithEntropyFilter(entropyThreshold, entropySize))
	}

	// Determine whether to append to existing group or create new
	var builder *obikmer.KmerSetGroupBuilder
	var err error
	metaPath := filepath.Join(outDir, "metadata.toml")
	if _, statErr := os.Stat(metaPath); statErr == nil {
		// Existing group: append
		log.Infof("Appending to existing kmer index at %s", outDir)
		builder, err = obikmer.AppendKmerSetGroupBuilder(outDir, 1, opts...)
		if err != nil {
			return fmt.Errorf("failed to open existing kmer index for appending: %w", err)
		}
	} else {
		// New group
		if maxOcc > 0 {
			log.Infof("Creating new kmer index: k=%d, m=%d, occurrence=[%d,%d]", k, m, minOcc, maxOcc)
		} else {
			log.Infof("Creating new kmer index: k=%d, m=%d, min-occurrence=%d", k, m, minOcc)
		}
		builder, err = obikmer.NewKmerSetGroupBuilder(outDir, k, m, 1, -1, opts...)
		if err != nil {
			return fmt.Errorf("failed to create kmer index builder: %w", err)
		}
	}

	// Read and process sequences in parallel
	sequences, err := obiconvert.CLIReadBioSequences(args...)
	if err != nil {
		return fmt.Errorf("failed to open sequence files: %w", err)
	}

	nworkers := obidefault.ParallelWorkers()
	var seqCount atomic.Int64
	var wg sync.WaitGroup

	consumer := func(iter obiiter.IBioSequence) {
		defer wg.Done()
		for iter.Next() {
			batch := iter.Get()
			for _, seq := range batch.Slice() {
				builder.AddSequence(0, seq)
				seqCount.Add(1)
			}
		}
	}

	for i := 1; i < nworkers; i++ {
		wg.Add(1)
		go consumer(sequences.Split())
	}
	wg.Add(1)
	go consumer(sequences)
	wg.Wait()

	log.Infof("Processed %d sequences", seqCount.Load())

	// Finalize
	ksg, err := builder.Close()
	if err != nil {
		return fmt.Errorf("failed to finalize kmer index: %w", err)
	}

	// Apply index-id to the new set
	newSetIdx := builder.StartIndex()
	if id := CLIIndexId(); id != "" {
		ksg.SetSetID(newSetIdx, id)
	}

	// Apply group-level tags (-S)
	for key, value := range CLISetTag() {
		ksg.SetAttribute(key, value)
	}

	// Apply per-set tags (-T) to the new set
	for key, value := range _setMetaTags {
		ksg.SetSetMetadata(newSetIdx, key, value)
	}

	if minOcc > 1 {
		ksg.SetAttribute("min_occurrence", minOcc)
	}
	if maxOcc > 0 {
		ksg.SetAttribute("max_occurrence", maxOcc)
	}

	if entropyThreshold > 0 {
		ksg.SetAttribute("entropy_filter", entropyThreshold)
		ksg.SetAttribute("entropy_filter_size", entropySize)
	}

	if err := ksg.SaveMetadata(); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	log.Infof("Index contains %d k-mers for set %d in %s", ksg.Len(newSetIdx), newSetIdx, outDir)
	log.Info("Done.")
	return nil
}
