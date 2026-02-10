package obik

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"github.com/DavidGamba/go-getoptions"
)

// KmerFilter is a predicate applied to individual k-mers during filtering.
// Returns true if the k-mer should be kept.
type KmerFilter func(kmer uint64) bool

// KmerFilterFactory creates a new KmerFilter instance.
// Each goroutine should call the factory to get its own filter,
// since some filters (e.g. KmerEntropyFilter) are not thread-safe.
type KmerFilterFactory func() KmerFilter

// chainFilterFactories combines multiple KmerFilterFactory into one.
// The resulting factory creates a filter that accepts a k-mer only
// if all individual filters accept it.
func chainFilterFactories(factories []KmerFilterFactory) KmerFilterFactory {
	switch len(factories) {
	case 0:
		return func() KmerFilter { return func(uint64) bool { return true } }
	case 1:
		return factories[0]
	default:
		return func() KmerFilter {
			filters := make([]KmerFilter, len(factories))
			for i, f := range factories {
				filters[i] = f()
			}
			return func(kmer uint64) bool {
				for _, f := range filters {
					if !f(kmer) {
						return false
					}
				}
				return true
			}
		}
	}
}

// runFilter implements the "obik filter" subcommand.
// It reads an existing kmer index, applies a chain of filters,
// and writes a new filtered index.
func runFilter(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: obik filter [options] <source_index> --out <dest_index>")
	}

	srcDir := args[0]
	destDir := CLIOutputDirectory()
	if destDir == "" || destDir == "-" {
		return fmt.Errorf("--out option is required and must specify a destination directory")
	}

	// Open source index
	src, err := obikmer.OpenKmerSetGroup(srcDir)
	if err != nil {
		return fmt.Errorf("failed to open source index: %w", err)
	}

	k := src.K()

	// Build filter factory chain from CLI options.
	// Factories are used so each goroutine creates its own filter instance,
	// since some filters (e.g. KmerEntropyFilter) have mutable state.
	var factories []KmerFilterFactory
	var filterDescriptions []string

	// Entropy filter
	entropyThreshold := CLIIndexEntropyThreshold()
	entropySize := CLIIndexEntropySize()
	if entropyThreshold > 0 {
		factories = append(factories, func() KmerFilter {
			ef := obikmer.NewKmerEntropyFilter(k, entropySize, entropyThreshold)
			return ef.Accept
		})
		filterDescriptions = append(filterDescriptions,
			fmt.Sprintf("entropy(threshold=%.4f, level-max=%d)", entropyThreshold, entropySize))
	}

	// Future filters will be added here, e.g.:
	// quorumFilter, frequencyFilter, ...

	if len(factories) == 0 {
		return fmt.Errorf("no filter specified; use --entropy-filter or other filter options")
	}

	filterFactory := chainFilterFactories(factories)

	// Resolve set selection (default: all sets)
	patterns := CLISetPatterns()
	var setIndices []int
	if len(patterns) > 0 {
		setIndices, err = src.MatchSetIDs(patterns)
		if err != nil {
			return fmt.Errorf("failed to match set patterns: %w", err)
		}
		if len(setIndices) == 0 {
			return fmt.Errorf("no sets match the given patterns")
		}
	} else {
		setIndices = make([]int, src.Size())
		for i := range setIndices {
			setIndices[i] = i
		}
	}

	log.Infof("Filtering %d set(s) from %s with: %s",
		len(setIndices), srcDir, strings.Join(filterDescriptions, " + "))

	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}

	P := src.Partitions()

	// Progress bar for partition filtering
	totalPartitions := len(setIndices) * P
	var bar *progressbar.ProgressBar
	if obidefault.ProgressBar() {
		pbopt := []progressbar.Option{
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionSetWidth(15),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetPredictTime(true),
			progressbar.OptionSetDescription("[Filtering partitions]"),
		}
		bar = progressbar.NewOptions(totalPartitions, pbopt...)
	}

	// Process each selected set
	newCounts := make([]uint64, len(setIndices))

	for si, srcIdx := range setIndices {
		setID := src.SetIDOf(srcIdx)
		if setID == "" {
			setID = fmt.Sprintf("set_%d", srcIdx)
		}

		destSetDir := filepath.Join(destDir, fmt.Sprintf("set_%d", si))
		if err := os.MkdirAll(destSetDir, 0755); err != nil {
			return fmt.Errorf("failed to create set directory: %w", err)
		}

		// Process partitions in parallel
		nWorkers := obidefault.ParallelWorkers()
		if nWorkers > P {
			nWorkers = P
		}

		var totalKept atomic.Uint64
		var totalProcessed atomic.Uint64

		type job struct {
			partIdx int
		}

		jobs := make(chan job, P)
		var wg sync.WaitGroup
		var errMu sync.Mutex
		var firstErr error

		for w := 0; w < nWorkers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				// Each goroutine gets its own filter instance
				workerFilter := filterFactory()
				for j := range jobs {
					kept, processed, err := filterPartition(
						src.PartitionPath(srcIdx, j.partIdx),
						filepath.Join(destSetDir, fmt.Sprintf("part_%04d.kdi", j.partIdx)),
						workerFilter,
					)
					if err != nil {
						errMu.Lock()
						if firstErr == nil {
							firstErr = err
						}
						errMu.Unlock()
						return
					}
					totalKept.Add(kept)
					totalProcessed.Add(processed)
					if bar != nil {
						bar.Add(1)
					}
				}
			}()
		}

		for p := 0; p < P; p++ {
			jobs <- job{p}
		}
		close(jobs)
		wg.Wait()

		if firstErr != nil {
			return fmt.Errorf("failed to filter set %q: %w", setID, firstErr)
		}

		kept := totalKept.Load()
		processed := totalProcessed.Load()
		newCounts[si] = kept
		log.Infof("Set %q: %d/%d k-mers kept (%.1f%% removed)",
			setID, kept, processed,
			100.0*float64(processed-kept)/float64(max(processed, 1)))

		// Copy spectrum.bin if it exists
		srcSpecPath := src.SpectrumPath(srcIdx)
		if _, err := os.Stat(srcSpecPath); err == nil {
			destSpecPath := filepath.Join(destSetDir, "spectrum.bin")
			if err := copyFileHelper(srcSpecPath, destSpecPath); err != nil {
				log.Warnf("Could not copy spectrum for set %q: %v", setID, err)
			}
		}
	}

	if bar != nil {
		fmt.Fprintln(os.Stderr)
	}

	// Build destination metadata
	setsIDs := make([]string, len(setIndices))
	setsMetadata := make([]map[string]interface{}, len(setIndices))
	for i, srcIdx := range setIndices {
		setsIDs[i] = src.SetIDOf(srcIdx)
		setsMetadata[i] = src.AllSetMetadata(srcIdx)
		if setsMetadata[i] == nil {
			setsMetadata[i] = make(map[string]interface{})
		}
	}

	// Write metadata for the filtered index
	dest, err := obikmer.NewFilteredKmerSetGroup(
		destDir, k, src.M(), P,
		len(setIndices), setsIDs, newCounts, setsMetadata,
	)
	if err != nil {
		return fmt.Errorf("failed to create filtered metadata: %w", err)
	}

	// Copy group-level metadata and record applied filters
	for key, value := range src.Metadata {
		dest.SetAttribute(key, value)
	}
	if entropyThreshold > 0 {
		dest.SetAttribute("entropy_filter", entropyThreshold)
		dest.SetAttribute("entropy_filter_size", entropySize)
	}
	dest.SetAttribute("filtered_from", srcDir)

	if err := dest.SaveMetadata(); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	log.Info("Done.")
	return nil
}

// filterPartition reads a single .kdi partition, applies the filter predicate,
// and writes the accepted k-mers to a new .kdi file.
// Returns (kept, processed, error).
func filterPartition(srcPath, destPath string, accept KmerFilter) (uint64, uint64, error) {
	reader, err := obikmer.NewKdiReader(srcPath)
	if err != nil {
		// Empty partition — write empty KDI
		w, err2 := obikmer.NewKdiWriter(destPath)
		if err2 != nil {
			return 0, 0, err2
		}
		return 0, 0, w.Close()
	}
	defer reader.Close()

	w, err := obikmer.NewKdiWriter(destPath)
	if err != nil {
		return 0, 0, err
	}

	var kept, processed uint64
	for {
		kmer, ok := reader.Next()
		if !ok {
			break
		}
		processed++
		if accept(kmer) {
			if err := w.Write(kmer); err != nil {
				w.Close()
				return 0, 0, err
			}
			kept++
		}
	}

	return kept, processed, w.Close()
}

// copyFileHelper copies a file (used for spectrum.bin etc.)
func copyFileHelper(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	buf := make([]byte, 32*1024)
	for {
		n, readErr := in.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
		}
		if readErr != nil {
			break
		}
	}
	return out.Close()
}
