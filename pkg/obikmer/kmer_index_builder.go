package obikmer

import (
	"math"
	"sync"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
)

// DefaultMinimizerSize returns ceil(k / 2.5) as a reasonable default minimizer size.
func DefaultMinimizerSize(k int) int {
	m := int(math.Ceil(float64(k) / 2.5))
	if m < 1 {
		m = 1
	}
	if m >= k {
		m = k - 1
	}
	return m
}

// MinMinimizerSize returns the minimum m such that 4^m >= nworkers,
// i.e. ceil(log(nworkers) / log(4)).
func MinMinimizerSize(nworkers int) int {
	if nworkers <= 1 {
		return 1
	}
	return int(math.Ceil(math.Log(float64(nworkers)) / math.Log(4)))
}

// ValidateMinimizerSize checks and adjusts the minimizer size to satisfy constraints:
// - m >= ceil(log(nworkers)/log(4))
// - 1 <= m < k
func ValidateMinimizerSize(m, k, nworkers int) int {
	minM := MinMinimizerSize(nworkers)
	if m < minM {
		log.Warnf("Minimizer size %d too small for %d workers (4^%d = %d < %d), adjusting to %d",
			m, nworkers, m, 1<<(2*m), nworkers, minM)
		m = minM
	}
	if m < 1 {
		m = 1
	}
	if m >= k {
		m = k - 1
	}
	return m
}

// BuildKmerIndex builds a KmerSet from an iterator using parallel super-kmer partitioning.
//
// The algorithm:
//  1. Extract super-kmers from each sequence using IterSuperKmers
//  2. Route each super-kmer to a worker based on minimizer % nworkers
//  3. Each worker extracts canonical k-mers and adds them to its local KmerSet
//  4. Merge all KmerSets via Union
//
// Parameters:
//   - iterator: source of BioSequence batches
//   - k: k-mer size (1-31)
//   - m: minimizer size (1 to k-1)
func BuildKmerIndex(iterator obiiter.IBioSequence, k, m int) *KmerSet {
	nproc := obidefault.ParallelWorkers()
	m = ValidateMinimizerSize(m, k, nproc)

	// Channels to route super-kmers to workers
	channels := make([]chan SuperKmer, nproc)
	for i := range channels {
		channels[i] = make(chan SuperKmer, 1024)
	}

	// Workers: each manages a partition of the minimizer space
	sets := make([]*KmerSet, nproc)
	waiter := sync.WaitGroup{}
	waiter.Add(nproc)
	for i := 0; i < nproc; i++ {
		sets[i] = NewKmerSet(k)
		go func(ch chan SuperKmer, ks *KmerSet) {
			defer waiter.Done()
			for sk := range ch {
				for kmer := range IterCanonicalKmers(sk.Sequence, k) {
					ks.AddKmerCode(kmer)
				}
			}
		}(channels[i], sets[i])
	}

	// Reader: extract super-kmers and route them
	seqCount := 0
	for iterator.Next() {
		batch := iterator.Get()
		for _, seq := range batch.Slice() {
			rawSeq := seq.Sequence()
			if len(rawSeq) < k {
				continue
			}
			for sk := range IterSuperKmers(rawSeq, k, m) {
				worker := int(sk.Minimizer % uint64(nproc))
				channels[worker] <- sk
			}
			seqCount++
		}
	}

	// Close channels to signal workers to finish
	for _, ch := range channels {
		close(ch)
	}
	waiter.Wait()

	log.Infof("Processed %d sequences", seqCount)

	// Merge partitions (mostly disjoint -> fast union)
	result := sets[0]
	for i := 1; i < nproc; i++ {
		result.bitmap.Or(sets[i].bitmap)
	}

	log.Infof("Index contains %d k-mers (%.2f MB)",
		result.Len(), float64(result.MemoryUsage())/1024/1024)

	return result
}

// BuildFrequencyFilterIndex builds a FrequencyFilter from an iterator
// using parallel super-kmer partitioning.
//
// Each worker manages its own FrequencyFilter for its partition of the
// minimizer space. Since all k-mers sharing a minimizer go to the same worker,
// the frequency counting is correct per partition.
//
// Parameters:
//   - iterator: source of BioSequence batches
//   - k: k-mer size (1-31)
//   - m: minimizer size (1 to k-1)
//   - minFreq: minimum frequency threshold (>= 1)
func BuildFrequencyFilterIndex(iterator obiiter.IBioSequence, k, m, minFreq int) *FrequencyFilter {
	nproc := obidefault.ParallelWorkers()
	m = ValidateMinimizerSize(m, k, nproc)

	// Channels to route super-kmers to workers
	channels := make([]chan SuperKmer, nproc)
	for i := range channels {
		channels[i] = make(chan SuperKmer, 1024)
	}

	// Workers: each manages a local FrequencyFilter
	filters := make([]*FrequencyFilter, nproc)
	waiter := sync.WaitGroup{}
	waiter.Add(nproc)
	for i := 0; i < nproc; i++ {
		filters[i] = NewFrequencyFilter(k, minFreq)
		go func(ch chan SuperKmer, ff *FrequencyFilter) {
			defer waiter.Done()
			for sk := range ch {
				for kmer := range IterCanonicalKmers(sk.Sequence, k) {
					ff.AddKmerCode(kmer)
				}
			}
		}(channels[i], filters[i])
	}

	// Reader: extract super-kmers and route them
	seqCount := 0
	for iterator.Next() {
		batch := iterator.Get()
		for _, seq := range batch.Slice() {
			rawSeq := seq.Sequence()
			if len(rawSeq) < k {
				continue
			}
			for sk := range IterSuperKmers(rawSeq, k, m) {
				worker := int(sk.Minimizer % uint64(nproc))
				channels[worker] <- sk
			}
			seqCount++
		}
	}

	// Close channels to signal workers to finish
	for _, ch := range channels {
		close(ch)
	}
	waiter.Wait()

	log.Infof("Processed %d sequences", seqCount)

	// Merge FrequencyFilters: union level by level
	result := filters[0]
	for i := 1; i < nproc; i++ {
		for level := 0; level < minFreq; level++ {
			result.Get(level).bitmap.Or(filters[i].Get(level).bitmap)
		}
	}

	stats := result.Stats()
	log.Infof("FrequencyFilter: %d k-mers with freq >= %d (%.2f MB total)",
		stats.FilteredKmers, minFreq, float64(stats.TotalBytes)/1024/1024)

	return result
}
