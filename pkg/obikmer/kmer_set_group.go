package obikmer

import (
	"fmt"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidist"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// KmerSetGroup represents a vector of KmerSet
// Used to manage multiple k-mer sets (for example, by frequency level)
type KmerSetGroup struct {
	id       string                 // Unique identifier of the KmerSetGroup
	k        int                    // Size of k-mers (immutable)
	sets     []*KmerSet             // Vector of KmerSet
	Metadata map[string]interface{} // Group metadata (not individual sets)
}

// NewKmerSetGroup creates a new group of n KmerSets
func NewKmerSetGroup(k int, n int) *KmerSetGroup {
	if n < 1 {
		panic("KmerSetGroup size must be >= 1")
	}

	sets := make([]*KmerSet, n)
	for i := range sets {
		sets[i] = NewKmerSet(k)
	}

	return &KmerSetGroup{
		k:        k,
		sets:     sets,
		Metadata: make(map[string]interface{}),
	}
}

// K returns the size of k-mers (immutable)
func (ksg *KmerSetGroup) K() int {
	return ksg.k
}

// Size returns the number of KmerSet in the group
func (ksg *KmerSetGroup) Size() int {
	return len(ksg.sets)
}

// Get returns the KmerSet at the given index
// Returns nil if the index is invalid
func (ksg *KmerSetGroup) Get(index int) *KmerSet {
	if index < 0 || index >= len(ksg.sets) {
		return nil
	}
	return ksg.sets[index]
}

// Set replaces the KmerSet at the given index
// Panics if the index is invalid or if k does not match
func (ksg *KmerSetGroup) Set(index int, ks *KmerSet) {
	if index < 0 || index >= len(ksg.sets) {
		panic(fmt.Sprintf("Index out of bounds: %d (size: %d)", index, len(ksg.sets)))
	}
	if ks.k != ksg.k {
		panic(fmt.Sprintf("KmerSet k mismatch: expected %d, got %d", ksg.k, ks.k))
	}
	ksg.sets[index] = ks
}

// Len returns the number of k-mers in a specific KmerSet
// Without argument: returns the number of k-mers in the last KmerSet
// With argument index: returns the number of k-mers in the KmerSet at this index
func (ksg *KmerSetGroup) Len(index ...int) uint64 {
	if len(index) == 0 {
		// Without argument: last KmerSet
		return ksg.sets[len(ksg.sets)-1].Len()
	}

	// With argument: specific KmerSet
	idx := index[0]
	if idx < 0 || idx >= len(ksg.sets) {
		return 0
	}
	return ksg.sets[idx].Len()
}

// MemoryUsage returns the total memory usage in bytes
func (ksg *KmerSetGroup) MemoryUsage() uint64 {
	total := uint64(0)
	for _, ks := range ksg.sets {
		total += ks.MemoryUsage()
	}
	return total
}

// Clear empties all KmerSet in the group
func (ksg *KmerSetGroup) Clear() {
	for _, ks := range ksg.sets {
		ks.Clear()
	}
}

// Copy creates a complete copy of the group (consistent with BioSequence.Copy)
func (ksg *KmerSetGroup) Copy() *KmerSetGroup {
	copiedSets := make([]*KmerSet, len(ksg.sets))
	for i, ks := range ksg.sets {
		copiedSets[i] = ks.Copy() // Copy each KmerSet with its metadata
	}

	// Copy group metadata
	groupMetadata := make(map[string]interface{}, len(ksg.Metadata))
	for k, v := range ksg.Metadata {
		groupMetadata[k] = v
	}

	return &KmerSetGroup{
		id:       ksg.id,
		k:        ksg.k,
		sets:     copiedSets,
		Metadata: groupMetadata,
	}
}

// Id returns the identifier of the KmerSetGroup (consistent with BioSequence.Id)
func (ksg *KmerSetGroup) Id() string {
	return ksg.id
}

// SetId sets the identifier of the KmerSetGroup (consistent with BioSequence.SetId)
func (ksg *KmerSetGroup) SetId(id string) {
	ksg.id = id
}

// AddSequence adds all k-mers from a sequence to a specific KmerSet
func (ksg *KmerSetGroup) AddSequence(seq *obiseq.BioSequence, index int) {
	if index < 0 || index >= len(ksg.sets) {
		panic(fmt.Sprintf("Index out of bounds: %d (size: %d)", index, len(ksg.sets)))
	}
	ksg.sets[index].AddSequence(seq)
}

// AddSequences adds all k-mers from multiple sequences to a specific KmerSet
func (ksg *KmerSetGroup) AddSequences(sequences *obiseq.BioSequenceSlice, index int) {
	if index < 0 || index >= len(ksg.sets) {
		panic(fmt.Sprintf("Index out of bounds: %d (size: %d)", index, len(ksg.sets)))
	}
	ksg.sets[index].AddSequences(sequences)
}

// Union returns the union of all KmerSet in the group
// Optimization: starts from the largest set to minimize operations
func (ksg *KmerSetGroup) Union() *KmerSet {
	if len(ksg.sets) == 0 {
		return NewKmerSet(ksg.k)
	}

	if len(ksg.sets) == 1 {
		return ksg.sets[0].Copy()
	}

	// Find the index of the largest set (the one with the most k-mers)
	maxIdx := 0
	maxCard := ksg.sets[0].Len()
	for i := 1; i < len(ksg.sets); i++ {
		card := ksg.sets[i].Len()
		if card > maxCard {
			maxCard = card
			maxIdx = i
		}
	}

	// Copy the largest set and perform unions in-place
	result := ksg.sets[maxIdx].bitmap.Clone()
	for i := 0; i < len(ksg.sets); i++ {
		if i != maxIdx {
			result.Or(ksg.sets[i].bitmap)
		}
	}

	return NewKmerSetFromBitmap(ksg.k, result)
}

// Intersect returns the intersection of all KmerSet in the group
// Optimization: starts from the smallest set to minimize operations
func (ksg *KmerSetGroup) Intersect() *KmerSet {
	if len(ksg.sets) == 0 {
		return NewKmerSet(ksg.k)
	}

	if len(ksg.sets) == 1 {
		return ksg.sets[0].Copy()
	}

	// Find the index of the smallest set (the one with the fewest k-mers)
	minIdx := 0
	minCard := ksg.sets[0].Len()
	for i := 1; i < len(ksg.sets); i++ {
		card := ksg.sets[i].Len()
		if card < minCard {
			minCard = card
			minIdx = i
		}
	}

	// Copy the smallest set and perform intersections in-place
	result := ksg.sets[minIdx].bitmap.Clone()
	for i := 0; i < len(ksg.sets); i++ {
		if i != minIdx {
			result.And(ksg.sets[i].bitmap)
		}
	}

	return NewKmerSetFromBitmap(ksg.k, result)
}

// Stats returns statistics for each KmerSet in the group
type KmerSetGroupStats struct {
	K          int
	Size       int              // Number of KmerSet
	TotalBytes uint64           // Total memory used
	Sets       []KmerSetStats   // Stats of each KmerSet
}

type KmerSetStats struct {
	Index     int    // Index of the KmerSet in the group
	Len       uint64 // Number of k-mers
	SizeBytes uint64 // Size in bytes
}

func (ksg *KmerSetGroup) Stats() KmerSetGroupStats {
	stats := KmerSetGroupStats{
		K:    ksg.k,
		Size: len(ksg.sets),
		Sets: make([]KmerSetStats, len(ksg.sets)),
	}

	for i, ks := range ksg.sets {
		sizeBytes := ks.MemoryUsage()
		stats.Sets[i] = KmerSetStats{
			Index:     i,
			Len:       ks.Len(),
			SizeBytes: sizeBytes,
		}
		stats.TotalBytes += sizeBytes
	}

	return stats
}

func (ksgs KmerSetGroupStats) String() string {
	result := fmt.Sprintf(`KmerSetGroup Statistics (k=%d, size=%d):
  Total memory: %.2f MB

Set breakdown:
`, ksgs.K, ksgs.Size, float64(ksgs.TotalBytes)/1024/1024)

	for _, set := range ksgs.Sets {
		result += fmt.Sprintf("  Set[%d]: %d k-mers (%.2f MB)\n",
			set.Index,
			set.Len,
			float64(set.SizeBytes)/1024/1024)
	}

	return result
}

// JaccardDistanceMatrix computes a pairwise Jaccard distance matrix for all KmerSets in the group.
// Returns a triangular distance matrix where element (i, j) represents the Jaccard distance
// between set i and set j.
//
// The Jaccard distance is: 1 - (|A ∩ B| / |A ∪ B|)
//
// The matrix labels are set to the IDs of the individual KmerSets if available,
// otherwise they are set to "set_0", "set_1", etc.
//
// Time complexity: O(n² × (|A| + |B|)) where n is the number of sets
// Space complexity: O(n²) for the distance matrix
func (ksg *KmerSetGroup) JaccardDistanceMatrix() *obidist.DistMatrix {
	n := len(ksg.sets)

	// Create labels from set IDs
	labels := make([]string, n)
	for i, ks := range ksg.sets {
		if ks.Id() != "" {
			labels[i] = ks.Id()
		} else {
			labels[i] = fmt.Sprintf("set_%d", i)
		}
	}

	dm := obidist.NewDistMatrixWithLabels(labels)

	// Compute pairwise distances
	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			distance := ksg.sets[i].JaccardDistance(ksg.sets[j])
			dm.Set(i, j, distance)
		}
	}

	return dm
}

// JaccardSimilarityMatrix computes a pairwise Jaccard similarity matrix for all KmerSets in the group.
// Returns a similarity matrix where element (i, j) represents the Jaccard similarity
// between set i and set j.
//
// The Jaccard similarity is: |A ∩ B| / |A ∪ B|
//
// The diagonal is 1.0 (similarity of a set to itself).
//
// The matrix labels are set to the IDs of the individual KmerSets if available,
// otherwise they are set to "set_0", "set_1", etc.
//
// Time complexity: O(n² × (|A| + |B|)) where n is the number of sets
// Space complexity: O(n²) for the similarity matrix
func (ksg *KmerSetGroup) JaccardSimilarityMatrix() *obidist.DistMatrix {
	n := len(ksg.sets)

	// Create labels from set IDs
	labels := make([]string, n)
	for i, ks := range ksg.sets {
		if ks.Id() != "" {
			labels[i] = ks.Id()
		} else {
			labels[i] = fmt.Sprintf("set_%d", i)
		}
	}

	sm := obidist.NewSimilarityMatrixWithLabels(labels)

	// Compute pairwise similarities
	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			similarity := ksg.sets[i].JaccardSimilarity(ksg.sets[j])
			sm.Set(i, j, similarity)
		}
	}

	return sm
}
