package obikmer

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// BuilderOption is a functional option for KmerSetGroupBuilder.
type BuilderOption func(*builderConfig)

type builderConfig struct {
	minFreq      int // 0 means no frequency filtering (simple dedup)
	maxFreq      int // 0 means no upper bound
	saveFreqTopN int // >0 means save the N most frequent k-mers per set to CSV
}

// WithMinFrequency activates frequency filtering mode.
// Only k-mers seen >= minFreq times are kept in the final index.
func WithMinFrequency(minFreq int) BuilderOption {
	return func(c *builderConfig) {
		c.minFreq = minFreq
	}
}

// WithMaxFrequency sets the upper frequency bound.
// Only k-mers seen <= maxFreq times are kept in the final index.
func WithMaxFrequency(maxFreq int) BuilderOption {
	return func(c *builderConfig) {
		c.maxFreq = maxFreq
	}
}

// WithSaveFreqKmers saves the N most frequent k-mers per set to a CSV file
// (top_kmers.csv in each set directory).
func WithSaveFreqKmers(n int) BuilderOption {
	return func(c *builderConfig) {
		c.saveFreqTopN = n
	}
}

// KmerSetGroupBuilder constructs a KmerSetGroup on disk.
// During construction, super-kmers are written to temporary .skm files
// partitioned by minimizer. On Close(), each partition is finalized
// (sort, dedup, optional frequency filter) into .kdi files.
type KmerSetGroupBuilder struct {
	dir        string
	k          int
	m          int
	n          int // number of NEW sets being built
	P          int // number of partitions
	startIndex int // first set index (0 for new groups, existingN for appends)
	config     builderConfig
	existing   *KmerSetGroup  // non-nil when appending to existing group
	writers    [][]*SkmWriter // [setIndex][partIndex] (local index 0..n-1)
	mu         [][]sync.Mutex // per-writer mutex for concurrent access
	closed     bool
}

// NewKmerSetGroupBuilder creates a builder for a new KmerSetGroup.
//
// Parameters:
//   - directory: destination directory (created if necessary)
//   - k: k-mer size (1-31)
//   - m: minimizer size (-1 for auto = ceil(k/2.5))
//   - n: number of sets in the group
//   - P: number of partitions (-1 for auto)
//   - options: optional builder options (e.g. WithMinFrequency)
func NewKmerSetGroupBuilder(directory string, k, m, n, P int,
	options ...BuilderOption) (*KmerSetGroupBuilder, error) {

	if k < 2 || k > 31 {
		return nil, fmt.Errorf("obikmer: k must be between 2 and 31, got %d", k)
	}
	if n < 1 {
		return nil, fmt.Errorf("obikmer: n must be >= 1, got %d", n)
	}

	// Auto minimizer size
	if m < 0 {
		m = int(math.Ceil(float64(k) / 2.5))
	}
	if m < 1 {
		m = 1
	}
	if m >= k {
		m = k - 1
	}

	// Auto partition count
	if P < 0 {
		// Use 4^m as the maximum, capped at a reasonable value
		maxP := 1 << (2 * m) // 4^m
		P = maxP
		if P > 4096 {
			P = 4096
		}
		if P < 64 {
			P = 64
		}
	}

	// Apply options
	var config builderConfig
	for _, opt := range options {
		opt(&config)
	}

	// Create build directory structure
	buildDir := filepath.Join(directory, ".build")
	for s := 0; s < n; s++ {
		setDir := filepath.Join(buildDir, fmt.Sprintf("set_%d", s))
		if err := os.MkdirAll(setDir, 0755); err != nil {
			return nil, fmt.Errorf("obikmer: create build dir: %w", err)
		}
	}

	// Create SKM writers
	writers := make([][]*SkmWriter, n)
	mutexes := make([][]sync.Mutex, n)
	for s := 0; s < n; s++ {
		writers[s] = make([]*SkmWriter, P)
		mutexes[s] = make([]sync.Mutex, P)
		for p := 0; p < P; p++ {
			path := filepath.Join(buildDir, fmt.Sprintf("set_%d", s),
				fmt.Sprintf("part_%04d.skm", p))
			w, err := NewSkmWriter(path)
			if err != nil {
				// Close already-created writers
				for ss := 0; ss <= s; ss++ {
					for pp := 0; pp < P; pp++ {
						if writers[ss][pp] != nil {
							writers[ss][pp].Close()
						}
					}
				}
				return nil, fmt.Errorf("obikmer: create skm writer: %w", err)
			}
			writers[s][p] = w
		}
	}

	return &KmerSetGroupBuilder{
		dir:        directory,
		k:          k,
		m:          m,
		n:          n,
		P:          P,
		startIndex: 0,
		config:     config,
		writers:    writers,
		mu:         mutexes,
	}, nil
}

// AppendKmerSetGroupBuilder opens an existing KmerSetGroup and creates
// a builder that adds n new sets starting from the existing set count.
// The k, m, and partitions are inherited from the existing group.
func AppendKmerSetGroupBuilder(directory string, n int, options ...BuilderOption) (*KmerSetGroupBuilder, error) {
	existing, err := OpenKmerSetGroup(directory)
	if err != nil {
		return nil, fmt.Errorf("obikmer: open existing group: %w", err)
	}

	if n < 1 {
		return nil, fmt.Errorf("obikmer: n must be >= 1, got %d", n)
	}

	k := existing.K()
	m := existing.M()
	P := existing.Partitions()
	startIndex := existing.Size()

	var config builderConfig
	for _, opt := range options {
		opt(&config)
	}

	// Create build directory structure for new sets
	buildDir := filepath.Join(directory, ".build")
	for s := 0; s < n; s++ {
		setDir := filepath.Join(buildDir, fmt.Sprintf("set_%d", s))
		if err := os.MkdirAll(setDir, 0755); err != nil {
			return nil, fmt.Errorf("obikmer: create build dir: %w", err)
		}
	}

	// Create SKM writers for new sets
	writers := make([][]*SkmWriter, n)
	mutexes := make([][]sync.Mutex, n)
	for s := 0; s < n; s++ {
		writers[s] = make([]*SkmWriter, P)
		mutexes[s] = make([]sync.Mutex, P)
		for p := 0; p < P; p++ {
			path := filepath.Join(buildDir, fmt.Sprintf("set_%d", s),
				fmt.Sprintf("part_%04d.skm", p))
			w, err := NewSkmWriter(path)
			if err != nil {
				for ss := 0; ss <= s; ss++ {
					for pp := 0; pp < P; pp++ {
						if writers[ss][pp] != nil {
							writers[ss][pp].Close()
						}
					}
				}
				return nil, fmt.Errorf("obikmer: create skm writer: %w", err)
			}
			writers[s][p] = w
		}
	}

	return &KmerSetGroupBuilder{
		dir:        directory,
		k:          k,
		m:          m,
		n:          n,
		P:          P,
		startIndex: startIndex,
		config:     config,
		existing:   existing,
		writers:    writers,
		mu:         mutexes,
	}, nil
}

// StartIndex returns the first global set index for the new sets being built.
// For new groups this is 0; for appends it is the existing group's Size().
func (b *KmerSetGroupBuilder) StartIndex() int {
	return b.startIndex
}

// AddSequence extracts super-kmers from a sequence and writes them
// to the appropriate partition files for the given set.
func (b *KmerSetGroupBuilder) AddSequence(setIndex int, seq *obiseq.BioSequence) {
	if setIndex < 0 || setIndex >= b.n {
		return
	}
	rawSeq := seq.Sequence()
	if len(rawSeq) < b.k {
		return
	}
	for sk := range IterSuperKmers(rawSeq, b.k, b.m) {
		part := int(sk.Minimizer % uint64(b.P))
		b.mu[setIndex][part].Lock()
		b.writers[setIndex][part].Write(sk)
		b.mu[setIndex][part].Unlock()
	}
}

// AddSuperKmer writes a single super-kmer to the appropriate partition.
func (b *KmerSetGroupBuilder) AddSuperKmer(setIndex int, sk SuperKmer) {
	if setIndex < 0 || setIndex >= b.n {
		return
	}
	part := int(sk.Minimizer % uint64(b.P))
	b.mu[setIndex][part].Lock()
	b.writers[setIndex][part].Write(sk)
	b.mu[setIndex][part].Unlock()
}

// Close finalizes the construction:
//  1. Flush and close all SKM writers
//  2. For each partition of each set (in parallel):
//     - Load super-kmers from .skm
//     - Extract canonical k-mers
//     - Sort and deduplicate (count if frequency filter)
//     - Write .kdi file
//  3. Write metadata.toml
//  4. Remove .build/ directory
//
// Returns the finalized KmerSetGroup in read-only mode.
func (b *KmerSetGroupBuilder) Close() (*KmerSetGroup, error) {
	if b.closed {
		return nil, fmt.Errorf("obikmer: builder already closed")
	}
	b.closed = true

	// 1. Close all SKM writers
	for s := 0; s < b.n; s++ {
		for p := 0; p < b.P; p++ {
			if err := b.writers[s][p].Close(); err != nil {
				return nil, fmt.Errorf("obikmer: close skm writer set=%d part=%d: %w", s, p, err)
			}
		}
	}

	// 2. Create output directory structure for new sets
	for s := 0; s < b.n; s++ {
		globalIdx := b.startIndex + s
		setDir := filepath.Join(b.dir, fmt.Sprintf("set_%d", globalIdx))
		if err := os.MkdirAll(setDir, 0755); err != nil {
			return nil, fmt.Errorf("obikmer: create set dir: %w", err)
		}
	}

	// Process partitions in parallel
	counts := make([][]uint64, b.n)
	spectra := make([][]map[int]uint64, b.n)
	var topKmers [][]*TopNKmers
	for s := 0; s < b.n; s++ {
		counts[s] = make([]uint64, b.P)
		spectra[s] = make([]map[int]uint64, b.P)
	}
	if b.config.saveFreqTopN > 0 {
		topKmers = make([][]*TopNKmers, b.n)
		for s := 0; s < b.n; s++ {
			topKmers[s] = make([]*TopNKmers, b.P)
		}
	}

	nWorkers := runtime.NumCPU()
	if nWorkers > b.P {
		nWorkers = b.P
	}

	type job struct {
		setIdx  int
		partIdx int
	}

	jobs := make(chan job, b.n*b.P)
	var wg sync.WaitGroup
	var errMu sync.Mutex
	var firstErr error

	for w := 0; w < nWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				partSpec, partTop, err := b.finalizePartition(j.setIdx, j.partIdx, &counts[j.setIdx][j.partIdx])
				if err != nil {
					errMu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					errMu.Unlock()
				}
				spectra[j.setIdx][j.partIdx] = partSpec
				if topKmers != nil {
					topKmers[j.setIdx][j.partIdx] = partTop
				}
			}
		}()
	}

	for s := 0; s < b.n; s++ {
		for p := 0; p < b.P; p++ {
			jobs <- job{s, p}
		}
	}
	close(jobs)
	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	// Aggregate per-partition spectra into per-set spectra and write spectrum.bin
	for s := 0; s < b.n; s++ {
		globalIdx := b.startIndex + s
		setSpectrum := make(map[int]uint64)
		for p := 0; p < b.P; p++ {
			if spectra[s][p] != nil {
				MergeSpectraMaps(setSpectrum, spectra[s][p])
			}
		}
		if len(setSpectrum) > 0 {
			specPath := filepath.Join(b.dir, fmt.Sprintf("set_%d", globalIdx), "spectrum.bin")
			if err := WriteSpectrum(specPath, MapToSpectrum(setSpectrum)); err != nil {
				return nil, fmt.Errorf("obikmer: write spectrum set=%d: %w", globalIdx, err)
			}
		}
	}

	// Aggregate per-partition top-N k-mers and write CSV
	if topKmers != nil {
		for s := 0; s < b.n; s++ {
			globalIdx := b.startIndex + s
			merged := NewTopNKmers(b.config.saveFreqTopN)
			for p := 0; p < b.P; p++ {
				merged.MergeTopN(topKmers[s][p])
			}
			results := merged.Results()
			if len(results) > 0 {
				csvPath := filepath.Join(b.dir, fmt.Sprintf("set_%d", globalIdx), "top_kmers.csv")
				if err := WriteTopKmersCSV(csvPath, results, b.k); err != nil {
					return nil, fmt.Errorf("obikmer: write top kmers set=%d: %w", globalIdx, err)
				}
			}
		}
	}

	// 3. Build KmerSetGroup and write metadata
	newCounts := make([]uint64, b.n)
	for s := 0; s < b.n; s++ {
		for p := 0; p < b.P; p++ {
			newCounts[s] += counts[s][p]
		}
	}

	var ksg *KmerSetGroup

	if b.existing != nil {
		// Append mode: extend existing group
		ksg = b.existing
		ksg.n += b.n
		ksg.setsIDs = append(ksg.setsIDs, make([]string, b.n)...)
		ksg.counts = append(ksg.counts, newCounts...)
		newMeta := make([]map[string]interface{}, b.n)
		for i := range newMeta {
			newMeta[i] = make(map[string]interface{})
		}
		ksg.setsMetadata = append(ksg.setsMetadata, newMeta...)
	} else {
		// New group
		setsIDs := make([]string, b.n)
		setsMetadata := make([]map[string]interface{}, b.n)
		for i := range setsMetadata {
			setsMetadata[i] = make(map[string]interface{})
		}
		ksg = &KmerSetGroup{
			path:         b.dir,
			k:            b.k,
			m:            b.m,
			partitions:   b.P,
			n:            b.n,
			setsIDs:      setsIDs,
			counts:       newCounts,
			setsMetadata: setsMetadata,
			Metadata:     make(map[string]interface{}),
		}
	}

	if err := ksg.saveMetadata(); err != nil {
		return nil, fmt.Errorf("obikmer: write metadata: %w", err)
	}

	// 4. Remove .build/ directory
	buildDir := filepath.Join(b.dir, ".build")
	os.RemoveAll(buildDir)

	return ksg, nil
}

// finalizePartition processes a single partition: load SKM, extract k-mers,
// sort, dedup/count, write KDI. Returns a partial frequency spectrum
// (frequency → count of distinct k-mers) computed before filtering,
// and optionally the top-N most frequent k-mers.
func (b *KmerSetGroupBuilder) finalizePartition(setIdx, partIdx int, count *uint64) (map[int]uint64, *TopNKmers, error) {
	// setIdx is local (0..n-1); build dirs use local index, output dirs use global
	skmPath := filepath.Join(b.dir, ".build",
		fmt.Sprintf("set_%d", setIdx),
		fmt.Sprintf("part_%04d.skm", partIdx))

	globalIdx := b.startIndex + setIdx
	kdiPath := filepath.Join(b.dir,
		fmt.Sprintf("set_%d", globalIdx),
		fmt.Sprintf("part_%04d.kdi", partIdx))

	// Load super-kmers and extract canonical k-mers
	reader, err := NewSkmReader(skmPath)
	if err != nil {
		// If file doesn't exist or is empty, write empty KDI
		return nil, nil, b.writeEmptyKdi(kdiPath, count)
	}

	var kmers []uint64
	for {
		sk, ok := reader.Next()
		if !ok {
			break
		}
		for kmer := range IterCanonicalKmers(sk.Sequence, b.k) {
			kmers = append(kmers, kmer)
		}
	}
	reader.Close()

	if len(kmers) == 0 {
		return nil, nil, b.writeEmptyKdi(kdiPath, count)
	}

	// Sort
	sort.Slice(kmers, func(i, j int) bool { return kmers[i] < kmers[j] })

	// Write KDI based on mode
	w, err := NewKdiWriter(kdiPath)
	if err != nil {
		return nil, nil, err
	}

	minFreq := b.config.minFreq
	if minFreq <= 0 {
		minFreq = 1 // simple dedup
	}
	maxFreq := b.config.maxFreq // 0 means no upper bound

	// Prepare top-N collector if requested
	var topN *TopNKmers
	if b.config.saveFreqTopN > 0 {
		topN = NewTopNKmers(b.config.saveFreqTopN)
	}

	// Linear scan: count consecutive identical values and accumulate spectrum
	partSpectrum := make(map[int]uint64)
	i := 0
	for i < len(kmers) {
		val := kmers[i]
		c := 1
		for i+c < len(kmers) && kmers[i+c] == val {
			c++
		}
		partSpectrum[c]++
		if topN != nil {
			topN.Add(val, c)
		}
		if c >= minFreq && (maxFreq <= 0 || c <= maxFreq) {
			if err := w.Write(val); err != nil {
				w.Close()
				return nil, nil, err
			}
		}
		i += c
	}

	*count = w.Count()
	return partSpectrum, topN, w.Close()
}

func (b *KmerSetGroupBuilder) writeEmptyKdi(path string, count *uint64) error {
	w, err := NewKdiWriter(path)
	if err != nil {
		return err
	}
	*count = 0
	return w.Close()
}
