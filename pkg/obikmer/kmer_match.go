package obikmer

import (
	"sort"
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// QueryEntry represents a canonical k-mer to look up, together with
// metadata to trace the result back to the originating sequence and position.
type QueryEntry struct {
	Kmer   uint64 // canonical k-mer value
	SeqIdx int    // index within the batch
	Pos    int    // 0-based position in the sequence
}

// MatchResult maps sequence index → sorted slice of matched positions.
type MatchResult map[int][]int

// seqMatchResult collects matched positions for a single sequence.
type seqMatchResult struct {
	mu        sync.Mutex
	positions []int
}

// PrepareQueries extracts all canonical k-mers from a batch of sequences
// and groups them by partition using super-kmer minimizers.
//
// Returns queries[partition] where each slice is sorted by Kmer value.
func (ksg *KmerSetGroup) PrepareQueries(sequences []*obiseq.BioSequence) [][]QueryEntry {
	P := ksg.partitions
	k := ksg.k
	m := ksg.m

	// Pre-allocate partition buckets
	buckets := make([][]QueryEntry, P)
	for i := range buckets {
		buckets[i] = make([]QueryEntry, 0, 64)
	}

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
					Pos:    sk.Start + localPos,
				})
				localPos++
			}
		}
	}

	// Sort each bucket by k-mer value for merge-scan
	for p := range buckets {
		sort.Slice(buckets[p], func(i, j int) bool {
			return buckets[p][i].Kmer < buckets[p][j].Kmer
		})
	}

	return buckets
}

// MatchBatch looks up pre-sorted queries against one set of the index.
// Partitions are processed in parallel. For each partition, a merge-scan
// compares the sorted queries against the sorted KDI stream.
//
// Returns a MatchResult mapping sequence index to sorted matched positions.
func (ksg *KmerSetGroup) MatchBatch(setIndex int, queries [][]QueryEntry) MatchResult {
	P := ksg.partitions

	// Per-sequence result collectors
	var resultMu sync.Mutex
	resultMap := make(map[int]*seqMatchResult)

	getResult := func(seqIdx int) *seqMatchResult {
		resultMu.Lock()
		sr, ok := resultMap[seqIdx]
		if !ok {
			sr = &seqMatchResult{}
			resultMap[seqIdx] = sr
		}
		resultMu.Unlock()
		return sr
	}

	var wg sync.WaitGroup

	for p := 0; p < P; p++ {
		if len(queries[p]) == 0 {
			continue
		}
		wg.Add(1)
		go func(part int) {
			defer wg.Done()
			ksg.matchPartition(setIndex, part, queries[part], getResult)
		}(p)
	}

	wg.Wait()

	// Build final result with sorted positions
	result := make(MatchResult, len(resultMap))
	for seqIdx, sr := range resultMap {
		if len(sr.positions) > 0 {
			sort.Ints(sr.positions)
			result[seqIdx] = sr.positions
		}
	}

	return result
}

// matchPartition processes one partition: opens the KDI reader (with index),
// seeks to the first query, then merge-scans queries against the KDI stream.
func (ksg *KmerSetGroup) matchPartition(
	setIndex int,
	partIndex int,
	queries []QueryEntry, // sorted by Kmer
	getResult func(int) *seqMatchResult,
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
				sr := getResult(queries[qi].SeqIdx)
				sr.mu.Lock()
				sr.positions = append(sr.positions, queries[qi].Pos)
				sr.mu.Unlock()
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
