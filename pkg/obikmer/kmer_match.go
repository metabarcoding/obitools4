package obikmer

import (
	"cmp"
	"slices"
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// QueryEntry represents a canonical k-mer to look up, together with
// metadata to trace the result back to the originating sequence and position.
type QueryEntry struct {
	Kmer   uint64 // canonical k-mer value
	SeqIdx int    // index within the batch
	Pos    int    // 1-based position in the sequence
}

// MatchResult holds matched positions for each sequence in a batch.
// results[i] contains the sorted matched positions for sequence i.
type MatchResult [][]int

// PreparedQueries holds pre-computed query buckets along with the number
// of sequences they were built from. This is used by the accumulation
// pipeline to merge queries from multiple batches.
type PreparedQueries struct {
	Buckets [][]QueryEntry // queries[partition], each sorted by Kmer
	NSeqs   int            // number of sequences that produced these queries
	NKmers  int            // total number of k-mer entries across all partitions
}

// MergeQueries merges src into dst, offsetting all SeqIdx values in src
// by dst.NSeqs. Both dst and src must have the same number of partitions.
// After merging, src should not be reused.
//
// Each partition's entries are merged in sorted order (merge-sort of two
// already-sorted slices).
func MergeQueries(dst, src *PreparedQueries) {
	for p := range dst.Buckets {
		if len(src.Buckets[p]) == 0 {
			continue
		}

		offset := dst.NSeqs
		srcB := src.Buckets[p]

		// Offset SeqIdx in src entries
		for i := range srcB {
			srcB[i].SeqIdx += offset
		}

		if len(dst.Buckets[p]) == 0 {
			dst.Buckets[p] = srcB
			continue
		}

		// Merge two sorted slices
		dstB := dst.Buckets[p]
		merged := make([]QueryEntry, 0, len(dstB)+len(srcB))
		i, j := 0, 0
		for i < len(dstB) && j < len(srcB) {
			if dstB[i].Kmer <= srcB[j].Kmer {
				merged = append(merged, dstB[i])
				i++
			} else {
				merged = append(merged, srcB[j])
				j++
			}
		}
		merged = append(merged, dstB[i:]...)
		merged = append(merged, srcB[j:]...)
		dst.Buckets[p] = merged
	}
	dst.NSeqs += src.NSeqs
	dst.NKmers += src.NKmers
}

// PrepareQueries extracts all canonical k-mers from a batch of sequences
// and groups them by partition using super-kmer minimizers.
//
// Returns a PreparedQueries with sorted per-partition buckets.
func (ksg *KmerSetGroup) PrepareQueries(sequences []*obiseq.BioSequence) *PreparedQueries {
	P := ksg.partitions
	k := ksg.k
	m := ksg.m

	// Pre-allocate partition buckets
	buckets := make([][]QueryEntry, P)
	for i := range buckets {
		buckets[i] = make([]QueryEntry, 0, 64)
	}

	totalKmers := 0
	for seqIdx, seq := range sequences {
		bseq := seq.Sequence()
		if len(bseq) < k {
			continue
		}

		// Iterate super-kmers to get minimizer → partition mapping
		for sk := range IterSuperKmers(bseq, k, m) {
			partition := int(sk.Minimizer % uint64(P))

			// Iterate canonical k-mers within this super-kmer
			skSeq := sk.Sequence
			if len(skSeq) < k {
				continue
			}

			localPos := 0
			for kmer := range IterCanonicalKmers(skSeq, k) {
				buckets[partition] = append(buckets[partition], QueryEntry{
					Kmer:   kmer,
					SeqIdx: seqIdx,
					Pos:    sk.Start + localPos + 1,
				})
				localPos++
				totalKmers++
			}
		}
	}

	// Sort each bucket by k-mer value for merge-scan
	for p := range buckets {
		slices.SortFunc(buckets[p], func(a, b QueryEntry) int {
			return cmp.Compare(a.Kmer, b.Kmer)
		})
	}

	return &PreparedQueries{
		Buckets: buckets,
		NSeqs:   len(sequences),
		NKmers:  totalKmers,
	}
}

// MatchBatch looks up pre-sorted queries against one set of the index.
// Partitions are processed in parallel. For each partition, a merge-scan
// compares the sorted queries against the sorted KDI stream.
//
// Returns a MatchResult where result[i] contains sorted matched positions
// for sequence i.
func (ksg *KmerSetGroup) MatchBatch(setIndex int, pq *PreparedQueries) MatchResult {
	P := ksg.partitions

	// Pre-allocated per-sequence results and mutexes.
	// Each partition goroutine appends to results[seqIdx] with mus[seqIdx] held.
	// Contention is low: a sequence's k-mers span many partitions, but each
	// partition processes its queries sequentially and the critical section is tiny.
	results := make([][]int, pq.NSeqs)
	mus := make([]sync.Mutex, pq.NSeqs)

	var wg sync.WaitGroup

	for p := 0; p < P; p++ {
		if len(pq.Buckets[p]) == 0 {
			continue
		}
		wg.Add(1)
		go func(part int) {
			defer wg.Done()
			ksg.matchPartition(setIndex, part, pq.Buckets[part], results, mus)
		}(p)
	}

	wg.Wait()

	// Sort positions within each sequence
	for i := range results {
		if len(results[i]) > 1 {
			slices.Sort(results[i])
		}
	}

	return MatchResult(results)
}

// matchPartition processes one partition: opens the KDI reader (with index),
// seeks to the first query, then merge-scans queries against the KDI stream.
func (ksg *KmerSetGroup) matchPartition(
	setIndex int,
	partIndex int,
	queries []QueryEntry, // sorted by Kmer
	results [][]int,
	mus []sync.Mutex,
) {
	r, err := NewKdiIndexedReader(ksg.partitionPath(setIndex, partIndex))
	if err != nil {
		return
	}
	defer r.Close()

	if r.Count() == 0 || len(queries) == 0 {
		return
	}

	// Seek to the first query's neighborhood
	if err := r.SeekTo(queries[0].Kmer); err != nil {
		return
	}

	// Read first kmer from the stream after seek
	currentKmer, ok := r.Next()
	if !ok {
		return
	}

	qi := 0 // query index

	for qi < len(queries) {
		q := queries[qi]

		// If the next query is far ahead, re-seek instead of linear scan.
		// Only seek if we'd skip more k-mers than the index stride,
		// otherwise linear scan through the buffer is faster than a syscall.
		if r.index != nil && q.Kmer > currentKmer && r.Remaining() > uint64(r.index.stride) {
			_, skipCount, found := r.index.FindOffset(q.Kmer)
			if found && skipCount > r.read+uint64(r.index.stride) {
				if err := r.SeekTo(q.Kmer); err == nil {
					nextKmer, nextOk := r.Next()
					if !nextOk {
						return
					}
					currentKmer = nextKmer
					ok = true
				}
			}
		}

		// Advance KDI stream until >= query kmer
		for currentKmer < q.Kmer {
			currentKmer, ok = r.Next()
			if !ok {
				return // KDI exhausted
			}
		}

		if currentKmer == q.Kmer {
			// Match! Record all queries with this same k-mer value
			matchedKmer := q.Kmer
			for qi < len(queries) && queries[qi].Kmer == matchedKmer {
				idx := queries[qi].SeqIdx
				mus[idx].Lock()
				results[idx] = append(results[idx], queries[qi].Pos)
				mus[idx].Unlock()
				qi++
			}
		} else {
			// currentKmer > q.Kmer: skip all queries with this kmer value
			skippedKmer := q.Kmer
			for qi < len(queries) && queries[qi].Kmer == skippedKmer {
				qi++
			}
		}
	}
}
