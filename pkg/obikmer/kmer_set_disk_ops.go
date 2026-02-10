package obikmer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// Union computes the union of all sets in the group, producing a new
// singleton KmerSetGroup on disk. A k-mer is in the result if it
// appears in any set.
func (ksg *KmerSetGroup) Union(outputDir string) (*KmerSetGroup, error) {
	return ksg.quorumOp(outputDir, 1, ksg.n)
}

// Intersect computes the intersection of all sets, producing a new
// singleton KmerSetGroup on disk. A k-mer is in the result if it
// appears in every set.
func (ksg *KmerSetGroup) Intersect(outputDir string) (*KmerSetGroup, error) {
	return ksg.quorumOp(outputDir, ksg.n, ksg.n)
}

// Difference computes set_0 minus the union of all other sets.
func (ksg *KmerSetGroup) Difference(outputDir string) (*KmerSetGroup, error) {
	return ksg.differenceOp(outputDir)
}

// QuorumAtLeast returns k-mers present in at least q sets.
func (ksg *KmerSetGroup) QuorumAtLeast(q int, outputDir string) (*KmerSetGroup, error) {
	return ksg.quorumOp(outputDir, q, ksg.n)
}

// QuorumExactly returns k-mers present in exactly q sets.
func (ksg *KmerSetGroup) QuorumExactly(q int, outputDir string) (*KmerSetGroup, error) {
	return ksg.quorumOp(outputDir, q, q)
}

// QuorumAtMost returns k-mers present in at most q sets.
func (ksg *KmerSetGroup) QuorumAtMost(q int, outputDir string) (*KmerSetGroup, error) {
	return ksg.quorumOp(outputDir, 1, q)
}

// UnionWith merges this group with another, producing a new KmerSetGroup
// whose set_i is the union of this.set_i and other.set_i.
// Both groups must have the same k, m, P, and N.
func (ksg *KmerSetGroup) UnionWith(other *KmerSetGroup, outputDir string) (*KmerSetGroup, error) {
	if err := ksg.checkCompatible(other); err != nil {
		return nil, err
	}
	return ksg.pairwiseOp(other, outputDir, mergeUnion)
}

// IntersectWith merges this group with another, producing a new KmerSetGroup
// whose set_i is the intersection of this.set_i and other.set_i.
func (ksg *KmerSetGroup) IntersectWith(other *KmerSetGroup, outputDir string) (*KmerSetGroup, error) {
	if err := ksg.checkCompatible(other); err != nil {
		return nil, err
	}
	return ksg.pairwiseOp(other, outputDir, mergeIntersect)
}

// ==============================
// Internal implementation
// ==============================

func (ksg *KmerSetGroup) checkCompatible(other *KmerSetGroup) error {
	if ksg.k != other.k {
		return fmt.Errorf("obikmer: incompatible k: %d vs %d", ksg.k, other.k)
	}
	if ksg.m != other.m {
		return fmt.Errorf("obikmer: incompatible m: %d vs %d", ksg.m, other.m)
	}
	if ksg.partitions != other.partitions {
		return fmt.Errorf("obikmer: incompatible partitions: %d vs %d", ksg.partitions, other.partitions)
	}
	if ksg.n != other.n {
		return fmt.Errorf("obikmer: incompatible size: %d vs %d", ksg.n, other.n)
	}
	return nil
}

// quorumOp processes all N sets partition by partition.
// For each partition, it opens N KdiReaders and does a k-way merge.
// A kmer is written to the result if minQ <= count <= maxQ.
func (ksg *KmerSetGroup) quorumOp(outputDir string, minQ, maxQ int) (*KmerSetGroup, error) {
	if minQ < 1 {
		minQ = 1
	}
	if maxQ > ksg.n {
		maxQ = ksg.n
	}

	// Create output structure
	setDir := filepath.Join(outputDir, "set_0")
	if err := os.MkdirAll(setDir, 0755); err != nil {
		return nil, err
	}

	counts := make([]uint64, ksg.partitions)

	nWorkers := runtime.NumCPU()
	if nWorkers > ksg.partitions {
		nWorkers = ksg.partitions
	}

	jobs := make(chan int, ksg.partitions)
	var wg sync.WaitGroup
	var errMu sync.Mutex
	var firstErr error

	for w := 0; w < nWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range jobs {
				c, err := ksg.quorumPartition(p, setDir, minQ, maxQ)
				if err != nil {
					errMu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					errMu.Unlock()
					return
				}
				counts[p] = c
			}
		}()
	}

	for p := 0; p < ksg.partitions; p++ {
		jobs <- p
	}
	close(jobs)
	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	var totalCount uint64
	for _, c := range counts {
		totalCount += c
	}

	result := &KmerSetGroup{
		path:       outputDir,
		k:          ksg.k,
		m:          ksg.m,
		partitions: ksg.partitions,
		n:          1,
		setsIDs:    []string{""},
		counts:     []uint64{totalCount},
		Metadata:   make(map[string]interface{}),
	}

	if err := result.saveMetadata(); err != nil {
		return nil, err
	}

	return result, nil
}

// quorumPartition processes a single partition for quorum filtering.
func (ksg *KmerSetGroup) quorumPartition(partIdx int, outSetDir string, minQ, maxQ int) (uint64, error) {
	// Open readers for all sets
	readers := make([]*KdiReader, 0, ksg.n)
	for s := 0; s < ksg.n; s++ {
		r, err := NewKdiReader(ksg.partitionPath(s, partIdx))
		if err != nil {
			// Close already-opened readers
			for _, rr := range readers {
				rr.Close()
			}
			return 0, err
		}
		if r.Count() > 0 {
			readers = append(readers, r)
		} else {
			r.Close()
		}
	}

	outPath := filepath.Join(outSetDir, fmt.Sprintf("part_%04d.kdi", partIdx))

	if len(readers) == 0 {
		// Write empty KDI
		w, err := NewKdiWriter(outPath)
		if err != nil {
			return 0, err
		}
		return 0, w.Close()
	}

	merge := NewKWayMerge(readers)
	// merge.Close() will close readers

	w, err := NewKdiWriter(outPath)
	if err != nil {
		merge.Close()
		return 0, err
	}

	for {
		kmer, count, ok := merge.Next()
		if !ok {
			break
		}
		if count >= minQ && count <= maxQ {
			if err := w.Write(kmer); err != nil {
				merge.Close()
				w.Close()
				return 0, err
			}
		}
	}

	merge.Close()
	cnt := w.Count()
	return cnt, w.Close()
}

// differenceOp computes set_0 minus the union of all other sets.
func (ksg *KmerSetGroup) differenceOp(outputDir string) (*KmerSetGroup, error) {
	if ksg.n < 1 {
		return nil, fmt.Errorf("obikmer: difference requires at least 1 set")
	}

	setDir := filepath.Join(outputDir, "set_0")
	if err := os.MkdirAll(setDir, 0755); err != nil {
		return nil, err
	}

	counts := make([]uint64, ksg.partitions)

	nWorkers := runtime.NumCPU()
	if nWorkers > ksg.partitions {
		nWorkers = ksg.partitions
	}

	jobs := make(chan int, ksg.partitions)
	var wg sync.WaitGroup
	var errMu sync.Mutex
	var firstErr error

	for w := 0; w < nWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range jobs {
				c, err := ksg.differencePartition(p, setDir)
				if err != nil {
					errMu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					errMu.Unlock()
					return
				}
				counts[p] = c
			}
		}()
	}

	for p := 0; p < ksg.partitions; p++ {
		jobs <- p
	}
	close(jobs)
	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	var totalCount uint64
	for _, c := range counts {
		totalCount += c
	}

	result := &KmerSetGroup{
		path:       outputDir,
		k:          ksg.k,
		m:          ksg.m,
		partitions: ksg.partitions,
		n:          1,
		setsIDs:    []string{""},
		counts:     []uint64{totalCount},
		Metadata:   make(map[string]interface{}),
	}

	if err := result.saveMetadata(); err != nil {
		return nil, err
	}

	return result, nil
}

// differencePartition computes set_0 - union(set_1..set_{n-1}) for one partition.
func (ksg *KmerSetGroup) differencePartition(partIdx int, outSetDir string) (uint64, error) {
	outPath := filepath.Join(outSetDir, fmt.Sprintf("part_%04d.kdi", partIdx))

	// Open set_0 reader
	r0, err := NewKdiReader(ksg.partitionPath(0, partIdx))
	if err != nil {
		return 0, err
	}

	if r0.Count() == 0 {
		r0.Close()
		w, err := NewKdiWriter(outPath)
		if err != nil {
			return 0, err
		}
		return 0, w.Close()
	}

	// Open readers for the other sets and merge them
	var otherReaders []*KdiReader
	for s := 1; s < ksg.n; s++ {
		r, err := NewKdiReader(ksg.partitionPath(s, partIdx))
		if err != nil {
			r0.Close()
			for _, rr := range otherReaders {
				rr.Close()
			}
			return 0, err
		}
		if r.Count() > 0 {
			otherReaders = append(otherReaders, r)
		} else {
			r.Close()
		}
	}

	w, err := NewKdiWriter(outPath)
	if err != nil {
		r0.Close()
		for _, rr := range otherReaders {
			rr.Close()
		}
		return 0, err
	}

	if len(otherReaders) == 0 {
		// No other sets — copy set_0
		for {
			v, ok := r0.Next()
			if !ok {
				break
			}
			if err := w.Write(v); err != nil {
				r0.Close()
				w.Close()
				return 0, err
			}
		}
		r0.Close()
		cnt := w.Count()
		return cnt, w.Close()
	}

	// Merge other sets to get the "subtraction" stream
	otherMerge := NewKWayMerge(otherReaders)

	// Streaming difference: advance both streams
	v0, ok0 := r0.Next()
	vo, _, oko := otherMerge.Next()

	for ok0 {
		if !oko || v0 < vo {
			// v0 not in others → emit
			if err := w.Write(v0); err != nil {
				r0.Close()
				otherMerge.Close()
				w.Close()
				return 0, err
			}
			v0, ok0 = r0.Next()
		} else if v0 == vo {
			// v0 in others → skip
			v0, ok0 = r0.Next()
			vo, _, oko = otherMerge.Next()
		} else {
			// vo < v0 → advance others
			vo, _, oko = otherMerge.Next()
		}
	}

	r0.Close()
	otherMerge.Close()
	cnt := w.Count()
	return cnt, w.Close()
}

// mergeMode defines how to combine two values during pairwise operations.
type mergeMode int

const (
	mergeUnion     mergeMode = iota // emit if in either
	mergeIntersect                  // emit if in both
)

// pairwiseOp applies a merge operation between corresponding sets of two groups.
func (ksg *KmerSetGroup) pairwiseOp(other *KmerSetGroup, outputDir string, mode mergeMode) (*KmerSetGroup, error) {
	for s := 0; s < ksg.n; s++ {
		setDir := filepath.Join(outputDir, fmt.Sprintf("set_%d", s))
		if err := os.MkdirAll(setDir, 0755); err != nil {
			return nil, err
		}
	}

	counts := make([][]uint64, ksg.n)
	for s := 0; s < ksg.n; s++ {
		counts[s] = make([]uint64, ksg.partitions)
	}

	nWorkers := runtime.NumCPU()
	if nWorkers > ksg.partitions {
		nWorkers = ksg.partitions
	}

	type job struct {
		setIdx  int
		partIdx int
	}
	jobs := make(chan job, ksg.n*ksg.partitions)
	var wg sync.WaitGroup
	var errMu sync.Mutex
	var firstErr error

	for w := 0; w < nWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				c, err := pairwiseMergePartition(
					ksg.partitionPath(j.setIdx, j.partIdx),
					other.partitionPath(j.setIdx, j.partIdx),
					filepath.Join(outputDir, fmt.Sprintf("set_%d", j.setIdx),
						fmt.Sprintf("part_%04d.kdi", j.partIdx)),
					mode,
				)
				if err != nil {
					errMu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					errMu.Unlock()
					return
				}
				counts[j.setIdx][j.partIdx] = c
			}
		}()
	}

	for s := 0; s < ksg.n; s++ {
		for p := 0; p < ksg.partitions; p++ {
			jobs <- job{s, p}
		}
	}
	close(jobs)
	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	totalCounts := make([]uint64, ksg.n)
	setsIDs := make([]string, ksg.n)
	for s := 0; s < ksg.n; s++ {
		for p := 0; p < ksg.partitions; p++ {
			totalCounts[s] += counts[s][p]
		}
	}

	result := &KmerSetGroup{
		path:       outputDir,
		k:          ksg.k,
		m:          ksg.m,
		partitions: ksg.partitions,
		n:          ksg.n,
		setsIDs:    setsIDs,
		counts:     totalCounts,
		Metadata:   make(map[string]interface{}),
	}

	if err := result.saveMetadata(); err != nil {
		return nil, err
	}

	return result, nil
}

// pairwiseMergePartition merges two KDI files (sorted streams) with the given mode.
func pairwiseMergePartition(pathA, pathB, outPath string, mode mergeMode) (uint64, error) {
	rA, err := NewKdiReader(pathA)
	if err != nil {
		return 0, err
	}
	rB, err := NewKdiReader(pathB)
	if err != nil {
		rA.Close()
		return 0, err
	}

	w, err := NewKdiWriter(outPath)
	if err != nil {
		rA.Close()
		rB.Close()
		return 0, err
	}

	cnt, mergeErr := doPairwiseMerge(rA, rB, w, mode)
	rA.Close()
	rB.Close()
	closeErr := w.Close()
	if mergeErr != nil {
		return 0, mergeErr
	}
	return cnt, closeErr
}

func doPairwiseMerge(rA, rB *KdiReader, w *KdiWriter, mode mergeMode) (uint64, error) {
	vA, okA := rA.Next()
	vB, okB := rB.Next()

	for okA && okB {
		if vA == vB {
			if err := w.Write(vA); err != nil {
				return 0, err
			}
			vA, okA = rA.Next()
			vB, okB = rB.Next()
		} else if vA < vB {
			if mode == mergeUnion {
				if err := w.Write(vA); err != nil {
					return 0, err
				}
			}
			vA, okA = rA.Next()
		} else {
			if mode == mergeUnion {
				if err := w.Write(vB); err != nil {
					return 0, err
				}
			}
			vB, okB = rB.Next()
		}
	}

	if mode == mergeUnion {
		for okA {
			if err := w.Write(vA); err != nil {
				return 0, err
			}
			vA, okA = rA.Next()
		}
		for okB {
			if err := w.Write(vB); err != nil {
				return 0, err
			}
			vB, okB = rB.Next()
		}
	}

	return w.Count(), nil
}
