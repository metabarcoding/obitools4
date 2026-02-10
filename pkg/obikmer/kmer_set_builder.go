package obikmer

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"github.com/schollz/progressbar/v3"
)

// BuilderOption is a functional option for KmerSetGroupBuilder.
type BuilderOption func(*builderConfig)

type builderConfig struct {
	minFreq          int     // 0 means no frequency filtering (simple dedup)
	maxFreq          int     // 0 means no upper bound
	saveFreqTopN     int     // >0 means save the N most frequent k-mers per set to CSV
	entropyThreshold float64 // >0 means filter k-mers with entropy <= threshold
	entropyLevelMax  int     // max sub-word size for entropy (typically 6)
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

// WithEntropyFilter activates entropy-based low-complexity filtering.
// K-mers with entropy <= threshold are discarded during finalization.
// levelMax is the maximum sub-word size for entropy computation (typically 6).
func WithEntropyFilter(threshold float64, levelMax int) BuilderOption {
	return func(c *builderConfig) {
		c.entropyThreshold = threshold
		c.entropyLevelMax = levelMax
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

	// =====================================================================
	// 2-stage pipeline: readers (pure I/O) → workers (CPU + write)
	//
	// - nReaders goroutines read .skm files (pure I/O, fast)
	// - nWorkers goroutines extract k-mers, sort, dedup, filter, write .kdi
	//
	// One unbuffered channel between stages. Readers are truly I/O-bound
	// (small files, buffered reads), workers are CPU-bound and stay busy.
	// =====================================================================
	totalJobs := b.n * b.P

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

	nCPU := obidefault.ParallelWorkers()

	// Stage sizing
	nWorkers := nCPU     // CPU-bound: one per core
	nReaders := nCPU / 4 // pure I/O: few goroutines suffice
	if nReaders < 2 {
		nReaders = 2
	}
	if nReaders > 4 {
		nReaders = 4
	}
	if nWorkers > totalJobs {
		nWorkers = totalJobs
	}
	if nReaders > totalJobs {
		nReaders = totalJobs
	}

	var bar *progressbar.ProgressBar
	if obidefault.ProgressBar() {
		pbopt := []progressbar.Option{
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionSetWidth(15),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetPredictTime(true),
			progressbar.OptionSetDescription("[Finalizing partitions]"),
		}
		bar = progressbar.NewOptions(totalJobs, pbopt...)
	}

	// --- Channel types ---
	type partitionData struct {
		setIdx  int
		partIdx int
		skmers  []SuperKmer // raw super-kmers from I/O stage
	}

	type readJob struct {
		setIdx  int
		partIdx int
	}

	dataCh := make(chan *partitionData) // unbuffered
	readJobs := make(chan readJob, totalJobs)

	var errMu sync.Mutex
	var firstErr error

	// Fill job queue (buffered, all jobs pre-loaded)
	for s := 0; s < b.n; s++ {
		for p := 0; p < b.P; p++ {
			readJobs <- readJob{s, p}
		}
	}
	close(readJobs)

	// --- Stage 1: Readers (pure I/O) ---
	var readWg sync.WaitGroup
	for w := 0; w < nReaders; w++ {
		readWg.Add(1)
		go func() {
			defer readWg.Done()
			for rj := range readJobs {
				skmers, err := b.loadPartitionRaw(rj.setIdx, rj.partIdx)
				if err != nil {
					errMu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					errMu.Unlock()
				}
				dataCh <- &partitionData{rj.setIdx, rj.partIdx, skmers}
			}
		}()
	}

	go func() {
		readWg.Wait()
		close(dataCh)
	}()

	// --- Stage 2: Workers (CPU: extract k-mers + sort/filter + write .kdi) ---
	var workWg sync.WaitGroup
	for w := 0; w < nWorkers; w++ {
		workWg.Add(1)
		go func() {
			defer workWg.Done()
			for pd := range dataCh {
				// CPU: extract canonical k-mers from super-kmers
				kmers := extractCanonicalKmers(pd.skmers, b.k)
				pd.skmers = nil // allow GC of raw super-kmers

				// CPU: sort, dedup, filter
				filtered, spectrum, topN := b.sortFilterPartition(kmers)
				kmers = nil // allow GC of unsorted data

				// I/O: write .kdi file
				globalIdx := b.startIndex + pd.setIdx
				kdiPath := filepath.Join(b.dir,
					fmt.Sprintf("set_%d", globalIdx),
					fmt.Sprintf("part_%04d.kdi", pd.partIdx))

				n, err := b.writePartitionKdi(kdiPath, filtered)
				if err != nil {
					errMu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					errMu.Unlock()
				}
				counts[pd.setIdx][pd.partIdx] = n
				spectra[pd.setIdx][pd.partIdx] = spectrum
				if topKmers != nil {
					topKmers[pd.setIdx][pd.partIdx] = topN
				}
				if bar != nil {
					bar.Add(1)
				}
			}
		}()
	}

	workWg.Wait()

	if bar != nil {
		fmt.Fprintln(os.Stderr)
	}

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

// loadPartitionRaw reads a .skm file and returns raw super-kmers.
// This is pure I/O — no k-mer extraction is done here.
// Returns nil (not an error) if the .skm file is empty or missing.
func (b *KmerSetGroupBuilder) loadPartitionRaw(setIdx, partIdx int) ([]SuperKmer, error) {
	skmPath := filepath.Join(b.dir, ".build",
		fmt.Sprintf("set_%d", setIdx),
		fmt.Sprintf("part_%04d.skm", partIdx))

	fi, err := os.Stat(skmPath)
	if err != nil {
		return nil, nil // empty partition, not an error
	}

	reader, err := NewSkmReader(skmPath)
	if err != nil {
		return nil, nil
	}

	// Estimate capacity from file size. Each super-kmer record is
	// 2 bytes (length) + packed bases (~k/4 bytes), so roughly
	// (2 + k/4) bytes per super-kmer on average.
	avgRecordSize := 2 + b.k/4
	if avgRecordSize < 4 {
		avgRecordSize = 4
	}
	estCount := int(fi.Size()) / avgRecordSize

	skmers := make([]SuperKmer, 0, estCount)
	for {
		sk, ok := reader.Next()
		if !ok {
			break
		}
		skmers = append(skmers, sk)
	}
	reader.Close()

	return skmers, nil
}

// extractCanonicalKmers extracts all canonical k-mers from a slice of super-kmers.
// This is CPU-bound work (sliding-window forward/reverse complement).
func extractCanonicalKmers(skmers []SuperKmer, k int) []uint64 {
	// Pre-compute total capacity to avoid repeated slice growth.
	// Each super-kmer of length L yields L-k+1 canonical k-mers.
	total := 0
	for i := range skmers {
		n := len(skmers[i].Sequence) - k + 1
		if n > 0 {
			total += n
		}
	}

	kmers := make([]uint64, 0, total)
	for _, sk := range skmers {
		for kmer := range IterCanonicalKmers(sk.Sequence, k) {
			kmers = append(kmers, kmer)
		}
	}
	return kmers
}

// sortFilterPartition sorts, deduplicates, and filters k-mers in memory (CPU-bound).
// Returns the filtered sorted slice, frequency spectrum, and optional top-N.
func (b *KmerSetGroupBuilder) sortFilterPartition(kmers []uint64) ([]uint64, map[int]uint64, *TopNKmers) {
	if len(kmers) == 0 {
		return nil, nil, nil
	}

	// Sort (CPU-bound) — slices.Sort avoids reflection overhead of sort.Slice
	slices.Sort(kmers)

	minFreq := b.config.minFreq
	if minFreq <= 0 {
		minFreq = 1 // simple dedup
	}
	maxFreq := b.config.maxFreq

	// Prepare entropy filter if requested
	var entropyFilter *KmerEntropyFilter
	if b.config.entropyThreshold > 0 && b.config.entropyLevelMax > 0 {
		entropyFilter = NewKmerEntropyFilter(b.k, b.config.entropyLevelMax, b.config.entropyThreshold)
	}

	// Prepare top-N collector if requested
	var topN *TopNKmers
	if b.config.saveFreqTopN > 0 {
		topN = NewTopNKmers(b.config.saveFreqTopN)
	}

	// Linear scan: count consecutive identical values, filter, accumulate spectrum
	partSpectrum := make(map[int]uint64)
	filtered := make([]uint64, 0, len(kmers)/2)

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
			if entropyFilter == nil || entropyFilter.Accept(val) {
				filtered = append(filtered, val)
			}
		}
		i += c
	}

	return filtered, partSpectrum, topN
}

// writePartitionKdi writes a sorted slice of k-mers to a .kdi file (I/O-bound).
// Returns the number of k-mers written.
func (b *KmerSetGroupBuilder) writePartitionKdi(kdiPath string, kmers []uint64) (uint64, error) {
	w, err := NewKdiWriter(kdiPath)
	if err != nil {
		return 0, err
	}

	for _, val := range kmers {
		if err := w.Write(val); err != nil {
			w.Close()
			return 0, err
		}
	}

	n := w.Count()
	return n, w.Close()
}

func (b *KmerSetGroupBuilder) writeEmptyKdi(path string, count *uint64) error {
	w, err := NewKdiWriter(path)
	if err != nil {
		return err
	}
	*count = 0
	return w.Close()
}
