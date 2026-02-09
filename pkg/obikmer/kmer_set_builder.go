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
	minFreq int // 0 means no frequency filtering (simple dedup)
}

// WithMinFrequency activates frequency filtering mode.
// Only k-mers seen >= minFreq times are kept in the final index.
func WithMinFrequency(minFreq int) BuilderOption {
	return func(c *builderConfig) {
		c.minFreq = minFreq
	}
}

// KmerSetGroupBuilder constructs a KmerSetGroup on disk.
// During construction, super-kmers are written to temporary .skm files
// partitioned by minimizer. On Close(), each partition is finalized
// (sort, dedup, optional frequency filter) into .kdi files.
type KmerSetGroupBuilder struct {
	dir     string
	k       int
	m       int
	n       int // number of sets
	P       int // number of partitions
	config  builderConfig
	writers [][]*SkmWriter // [setIndex][partIndex]
	mu      [][]sync.Mutex // per-writer mutex for concurrent access
	closed  bool
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
		dir:     directory,
		k:       k,
		m:       m,
		n:       n,
		P:       P,
		config:  config,
		writers: writers,
		mu:      mutexes,
	}, nil
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

	// 2. Create output directory structure
	for s := 0; s < b.n; s++ {
		setDir := filepath.Join(b.dir, fmt.Sprintf("set_%d", s))
		if err := os.MkdirAll(setDir, 0755); err != nil {
			return nil, fmt.Errorf("obikmer: create set dir: %w", err)
		}
	}

	// Process partitions in parallel
	counts := make([][]uint64, b.n)
	for s := 0; s < b.n; s++ {
		counts[s] = make([]uint64, b.P)
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
				if err := b.finalizePartition(j.setIdx, j.partIdx, &counts[j.setIdx][j.partIdx]); err != nil {
					errMu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					errMu.Unlock()
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

	// 3. Build KmerSetGroup and write metadata
	totalCounts := make([]uint64, b.n)
	for s := 0; s < b.n; s++ {
		for p := 0; p < b.P; p++ {
			totalCounts[s] += counts[s][p]
		}
	}

	setsIDs := make([]string, b.n)

	ksg := &KmerSetGroup{
		path:       b.dir,
		k:          b.k,
		m:          b.m,
		partitions: b.P,
		n:          b.n,
		setsIDs:    setsIDs,
		counts:     totalCounts,
		Metadata:   make(map[string]interface{}),
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
// sort, dedup/count, write KDI.
func (b *KmerSetGroupBuilder) finalizePartition(setIdx, partIdx int, count *uint64) error {
	skmPath := filepath.Join(b.dir, ".build",
		fmt.Sprintf("set_%d", setIdx),
		fmt.Sprintf("part_%04d.skm", partIdx))

	kdiPath := filepath.Join(b.dir,
		fmt.Sprintf("set_%d", setIdx),
		fmt.Sprintf("part_%04d.kdi", partIdx))

	// Load super-kmers and extract canonical k-mers
	reader, err := NewSkmReader(skmPath)
	if err != nil {
		// If file doesn't exist or is empty, write empty KDI
		return b.writeEmptyKdi(kdiPath, count)
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
		return b.writeEmptyKdi(kdiPath, count)
	}

	// Sort
	sort.Slice(kmers, func(i, j int) bool { return kmers[i] < kmers[j] })

	// Write KDI based on mode
	w, err := NewKdiWriter(kdiPath)
	if err != nil {
		return err
	}

	minFreq := b.config.minFreq
	if minFreq <= 0 {
		minFreq = 1 // simple dedup
	}

	// Linear scan: count consecutive identical values
	i := 0
	for i < len(kmers) {
		val := kmers[i]
		c := 1
		for i+c < len(kmers) && kmers[i+c] == val {
			c++
		}
		if c >= minFreq {
			if err := w.Write(val); err != nil {
				w.Close()
				return err
			}
		}
		i += c
	}

	*count = w.Count()
	return w.Close()
}

func (b *KmerSetGroupBuilder) writeEmptyKdi(path string, count *uint64) error {
	w, err := NewKdiWriter(path)
	if err != nil {
		return err
	}
	*count = 0
	return w.Close()
}
