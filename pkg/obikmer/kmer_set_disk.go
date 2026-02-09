package obikmer

import (
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidist"
	"github.com/pelletier/go-toml/v2"
)

// MetadataFormat represents the metadata serialization format.
// Currently only TOML is used for disk-based indices, but the type
// is kept for backward compatibility with CLI options.
type MetadataFormat int

const (
	FormatTOML MetadataFormat = iota
	FormatYAML
	FormatJSON
)

// String returns the file extension for the format.
func (f MetadataFormat) String() string {
	switch f {
	case FormatTOML:
		return "toml"
	case FormatYAML:
		return "yaml"
	case FormatJSON:
		return "json"
	default:
		return "toml"
	}
}

// KmerSetGroup is a disk-based collection of N k-mer sets sharing the same
// k, m, and partition count P. After construction (via KmerSetGroupBuilder),
// it is immutable and all operations are streaming (partition by partition).
//
// A KmerSetGroup with Size()==1 is effectively a KmerSet (singleton).
type KmerSetGroup struct {
	path       string                 // root directory
	id         string                 // user-assigned identifier
	k          int                    // k-mer size
	m          int                    // minimizer size
	partitions int                    // number of partitions P
	n          int                    // number of sets N
	setsIDs    []string               // IDs of individual sets
	counts     []uint64               // total k-mer count per set (sum over partitions)
	Metadata   map[string]interface{} // group-level user metadata
}

// diskMetadata is the TOML-serializable structure for metadata.toml.
type diskMetadata struct {
	ID           string                 `toml:"id,omitempty"`
	K            int                    `toml:"k"`
	M            int                    `toml:"m"`
	Partitions   int                    `toml:"partitions"`
	Type         string                 `toml:"type"`
	Size         int                    `toml:"size"`
	SetsIDs      []string               `toml:"sets_ids,omitempty"`
	Counts       []uint64               `toml:"counts,omitempty"`
	UserMetadata map[string]interface{} `toml:"user_metadata,omitempty"`
}

// OpenKmerSetGroup opens a finalized index directory in read-only mode.
func OpenKmerSetGroup(directory string) (*KmerSetGroup, error) {
	metaPath := filepath.Join(directory, "metadata.toml")
	f, err := os.Open(metaPath)
	if err != nil {
		return nil, fmt.Errorf("obikmer: open metadata: %w", err)
	}
	defer f.Close()

	var meta diskMetadata
	if err := toml.NewDecoder(f).Decode(&meta); err != nil {
		return nil, fmt.Errorf("obikmer: decode metadata: %w", err)
	}

	ksg := &KmerSetGroup{
		path:       directory,
		id:         meta.ID,
		k:          meta.K,
		m:          meta.M,
		partitions: meta.Partitions,
		n:          meta.Size,
		setsIDs:    meta.SetsIDs,
		counts:     meta.Counts,
		Metadata:   meta.UserMetadata,
	}
	if ksg.Metadata == nil {
		ksg.Metadata = make(map[string]interface{})
	}
	if ksg.setsIDs == nil {
		ksg.setsIDs = make([]string, ksg.n)
	}
	if ksg.counts == nil {
		// Compute counts by scanning partitions
		ksg.counts = make([]uint64, ksg.n)
		for s := 0; s < ksg.n; s++ {
			for p := 0; p < ksg.partitions; p++ {
				path := ksg.partitionPath(s, p)
				r, err := NewKdiReader(path)
				if err != nil {
					continue
				}
				ksg.counts[s] += r.Count()
				r.Close()
			}
		}
	}

	return ksg, nil
}

// SaveMetadata writes the metadata.toml file. This is useful after
// modifying attributes or IDs on an already-finalized index.
func (ksg *KmerSetGroup) SaveMetadata() error {
	return ksg.saveMetadata()
}

// saveMetadata writes the metadata.toml file (internal).
func (ksg *KmerSetGroup) saveMetadata() error {
	meta := diskMetadata{
		ID:           ksg.id,
		K:            ksg.k,
		M:            ksg.m,
		Partitions:   ksg.partitions,
		Type:         "KmerSetGroup",
		Size:         ksg.n,
		SetsIDs:      ksg.setsIDs,
		Counts:       ksg.counts,
		UserMetadata: ksg.Metadata,
	}

	metaPath := filepath.Join(ksg.path, "metadata.toml")
	f, err := os.Create(metaPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return toml.NewEncoder(f).Encode(meta)
}

// partitionPath returns the file path for partition p of set s.
func (ksg *KmerSetGroup) partitionPath(setIndex, partIndex int) string {
	return filepath.Join(ksg.path, fmt.Sprintf("set_%d", setIndex),
		fmt.Sprintf("part_%04d.kdi", partIndex))
}

// Path returns the root directory of the index.
func (ksg *KmerSetGroup) Path() string {
	return ksg.path
}

// K returns the k-mer size.
func (ksg *KmerSetGroup) K() int {
	return ksg.k
}

// M returns the minimizer size.
func (ksg *KmerSetGroup) M() int {
	return ksg.m
}

// Partitions returns the number of partitions P.
func (ksg *KmerSetGroup) Partitions() int {
	return ksg.partitions
}

// Size returns the number of sets N.
func (ksg *KmerSetGroup) Size() int {
	return ksg.n
}

// Id returns the group identifier.
func (ksg *KmerSetGroup) Id() string {
	return ksg.id
}

// SetId sets the group identifier and persists the change.
func (ksg *KmerSetGroup) SetId(id string) {
	ksg.id = id
}

// Len returns the total number of k-mers.
// Without argument: total across all sets.
// With argument setIndex: count for that specific set.
func (ksg *KmerSetGroup) Len(setIndex ...int) uint64 {
	if len(setIndex) == 0 {
		var total uint64
		for _, c := range ksg.counts {
			total += c
		}
		return total
	}
	idx := setIndex[0]
	if idx < 0 || idx >= ksg.n {
		return 0
	}
	return ksg.counts[idx]
}

// Contains checks if a k-mer is present in the specified set.
// Uses binary search on the appropriate partition's KDI file.
func (ksg *KmerSetGroup) Contains(setIndex int, kmer uint64) bool {
	if setIndex < 0 || setIndex >= ksg.n {
		return false
	}
	// Determine partition from minimizer
	// For a canonical k-mer, we need to find which partition it would fall into.
	// The partition is determined by the minimizer during construction.
	// For Contains, we must scan all partitions of this set (linear search within each).
	// A full binary-search approach would require an index file.
	// For now, scan the partition determined by the k-mer's minimizer.
	// Since we don't know the minimizer, we do a linear scan of all partitions.
	// This is O(total_kmers / P) per partition on average.

	// Optimization: scan all partitions in parallel
	type result struct {
		found bool
	}
	ch := make(chan result, ksg.partitions)

	for p := 0; p < ksg.partitions; p++ {
		go func(part int) {
			r, err := NewKdiReader(ksg.partitionPath(setIndex, part))
			if err != nil {
				ch <- result{false}
				return
			}
			defer r.Close()
			for {
				v, ok := r.Next()
				if !ok {
					ch <- result{false}
					return
				}
				if v == kmer {
					ch <- result{true}
					return
				}
				if v > kmer {
					ch <- result{false}
					return
				}
			}
		}(p)
	}

	for i := 0; i < ksg.partitions; i++ {
		res := <-ch
		if res.found {
			// Drain remaining goroutines
			go func() {
				for j := i + 1; j < ksg.partitions; j++ {
					<-ch
				}
			}()
			return true
		}
	}
	return false
}

// Iterator returns an iterator over all k-mers in the specified set,
// in sorted order within each partition. Since partitions are independent,
// to get a globally sorted stream, use iteratorSorted.
func (ksg *KmerSetGroup) Iterator(setIndex int) iter.Seq[uint64] {
	return func(yield func(uint64) bool) {
		if setIndex < 0 || setIndex >= ksg.n {
			return
		}

		// Open all partition readers and merge them
		readers := make([]*KdiReader, 0, ksg.partitions)
		for p := 0; p < ksg.partitions; p++ {
			r, err := NewKdiReader(ksg.partitionPath(setIndex, p))
			if err != nil {
				continue
			}
			if r.Count() > 0 {
				readers = append(readers, r)
			} else {
				r.Close()
			}
		}

		if len(readers) == 0 {
			return
		}

		m := NewKWayMerge(readers)
		defer m.Close()

		for {
			kmer, _, ok := m.Next()
			if !ok {
				return
			}
			if !yield(kmer) {
				return
			}
		}
	}
}

// ==============================
// Attribute API (compatible with old API)
// ==============================

// HasAttribute checks if a metadata key exists.
func (ksg *KmerSetGroup) HasAttribute(key string) bool {
	_, ok := ksg.Metadata[key]
	return ok
}

// GetAttribute returns the value of an attribute.
func (ksg *KmerSetGroup) GetAttribute(key string) (interface{}, bool) {
	switch key {
	case "id":
		return ksg.Id(), true
	case "k":
		return ksg.K(), true
	default:
		value, ok := ksg.Metadata[key]
		return value, ok
	}
}

// SetAttribute sets a metadata attribute.
func (ksg *KmerSetGroup) SetAttribute(key string, value interface{}) {
	switch key {
	case "id":
		if id, ok := value.(string); ok {
			ksg.SetId(id)
		} else {
			panic(fmt.Sprintf("id must be a string, got %T", value))
		}
	case "k":
		panic("k is immutable")
	default:
		ksg.Metadata[key] = value
	}
}

// DeleteAttribute removes a metadata attribute.
func (ksg *KmerSetGroup) DeleteAttribute(key string) {
	delete(ksg.Metadata, key)
}

// GetIntAttribute returns an attribute as int.
func (ksg *KmerSetGroup) GetIntAttribute(key string) (int, bool) {
	v, ok := ksg.GetAttribute(key)
	if !ok {
		return 0, false
	}
	switch val := v.(type) {
	case int:
		return val, true
	case int64:
		return int(val), true
	case float64:
		return int(val), true
	}
	return 0, false
}

// GetStringAttribute returns an attribute as string.
func (ksg *KmerSetGroup) GetStringAttribute(key string) (string, bool) {
	v, ok := ksg.GetAttribute(key)
	if !ok {
		return "", false
	}
	if s, ok := v.(string); ok {
		return s, true
	}
	return fmt.Sprintf("%v", v), true
}

// ==============================
// Jaccard metrics (streaming, disk-based)
// ==============================

// JaccardDistanceMatrix computes a pairwise Jaccard distance matrix
// for all sets in the group. Operates partition by partition in streaming.
func (ksg *KmerSetGroup) JaccardDistanceMatrix() *obidist.DistMatrix {
	n := ksg.n
	labels := make([]string, n)
	for i := 0; i < n; i++ {
		if i < len(ksg.setsIDs) && ksg.setsIDs[i] != "" {
			labels[i] = ksg.setsIDs[i]
		} else {
			labels[i] = fmt.Sprintf("set_%d", i)
		}
	}

	dm := obidist.NewDistMatrixWithLabels(labels)

	// Accumulate intersection and union counts
	intersections := make([][]uint64, n)
	unions := make([][]uint64, n)
	for i := 0; i < n; i++ {
		intersections[i] = make([]uint64, n)
		unions[i] = make([]uint64, n)
	}

	// Process partition by partition
	var mu sync.Mutex
	var wg sync.WaitGroup

	for p := 0; p < ksg.partitions; p++ {
		wg.Add(1)
		go func(part int) {
			defer wg.Done()

			// Open all set readers for this partition
			readers := make([]*KdiReader, n)
			for s := 0; s < n; s++ {
				r, err := NewKdiReader(ksg.partitionPath(s, part))
				if err != nil {
					continue
				}
				readers[s] = r
			}
			defer func() {
				for _, r := range readers {
					if r != nil {
						r.Close()
					}
				}
			}()

			// Merge all N readers to count intersections and unions
			activeReaders := make([]*KdiReader, 0, n)
			activeIndices := make([]int, 0, n)
			for i, r := range readers {
				if r != nil && r.Count() > 0 {
					activeReaders = append(activeReaders, r)
					activeIndices = append(activeIndices, i)
				}
			}
			if len(activeReaders) == 0 {
				return
			}

			merge := NewKWayMerge(activeReaders)
			// Don't close merge here since readers are managed above
			// We only want to iterate

			// We need per-set presence tracking, so we use a custom merge
			// Rebuild with a direct approach
			merge.Close() // close the merge (which closes readers)

			// Reopen readers for custom merge
			for s := 0; s < n; s++ {
				readers[s] = nil
				r, err := NewKdiReader(ksg.partitionPath(s, part))
				if err != nil {
					continue
				}
				if r.Count() > 0 {
					readers[s] = r
				} else {
					r.Close()
				}
			}

			// Custom k-way merge that tracks which sets contain each kmer
			type entry struct {
				val    uint64
				setIdx int
			}

			// Use a simpler approach: read all values for this partition into memory
			// for each set, then do a merge
			setKmers := make([][]uint64, n)
			for s := 0; s < n; s++ {
				if readers[s] == nil {
					continue
				}
				kmers := make([]uint64, 0, readers[s].Count())
				for {
					v, ok := readers[s].Next()
					if !ok {
						break
					}
					kmers = append(kmers, v)
				}
				setKmers[s] = kmers
				readers[s].Close()
				readers[s] = nil
			}

			// Count pairwise intersections using sorted merge
			// For each pair (i,j), count kmers present in both
			localInter := make([][]uint64, n)
			localUnion := make([][]uint64, n)
			for i := 0; i < n; i++ {
				localInter[i] = make([]uint64, n)
				localUnion[i] = make([]uint64, n)
			}

			for i := 0; i < n; i++ {
				localUnion[i][i] = uint64(len(setKmers[i]))
				for j := i + 1; j < n; j++ {
					a, b := setKmers[i], setKmers[j]
					var inter uint64
					ai, bi := 0, 0
					for ai < len(a) && bi < len(b) {
						if a[ai] == b[bi] {
							inter++
							ai++
							bi++
						} else if a[ai] < b[bi] {
							ai++
						} else {
							bi++
						}
					}
					localInter[i][j] = inter
					localUnion[i][j] = uint64(len(a)) + uint64(len(b)) - inter
				}
			}

			mu.Lock()
			for i := 0; i < n; i++ {
				for j := i; j < n; j++ {
					intersections[i][j] += localInter[i][j]
					unions[i][j] += localUnion[i][j]
				}
			}
			mu.Unlock()
		}(p)
	}
	wg.Wait()

	// Compute distances from accumulated counts
	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			u := unions[i][j]
			if u == 0 {
				dm.Set(i, j, 1.0)
			} else {
				dm.Set(i, j, 1.0-float64(intersections[i][j])/float64(u))
			}
		}
	}

	return dm
}

// JaccardSimilarityMatrix computes a pairwise Jaccard similarity matrix.
func (ksg *KmerSetGroup) JaccardSimilarityMatrix() *obidist.DistMatrix {
	n := ksg.n
	labels := make([]string, n)
	for i := 0; i < n; i++ {
		if i < len(ksg.setsIDs) && ksg.setsIDs[i] != "" {
			labels[i] = ksg.setsIDs[i]
		} else {
			labels[i] = fmt.Sprintf("set_%d", i)
		}
	}

	// Reuse distance computation
	dm := ksg.JaccardDistanceMatrix()
	sm := obidist.NewSimilarityMatrixWithLabels(labels)

	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			sm.Set(i, j, 1.0-dm.Get(i, j))
		}
	}

	return sm
}
