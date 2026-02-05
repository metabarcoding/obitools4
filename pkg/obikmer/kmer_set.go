package obikmer

import (
	"fmt"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"github.com/RoaringBitmap/roaring/roaring64"
)

// KmerSet wraps a set of k-mers stored in a Roaring Bitmap
// Provides utility methods for manipulating k-mer sets
type KmerSet struct {
	id       string                 // Unique identifier of the KmerSet
	k        int                    // Size of k-mers (immutable)
	bitmap   *roaring64.Bitmap      // Bitmap containing the k-mers
	Metadata map[string]interface{} // User metadata (key=atomic value)
}

// NewKmerSet creates a new empty KmerSet
func NewKmerSet(k int) *KmerSet {
	return &KmerSet{
		k:        k,
		bitmap:   roaring64.New(),
		Metadata: make(map[string]interface{}),
	}
}

// NewKmerSetFromBitmap creates a KmerSet from an existing bitmap
func NewKmerSetFromBitmap(k int, bitmap *roaring64.Bitmap) *KmerSet {
	return &KmerSet{
		k:        k,
		bitmap:   bitmap,
		Metadata: make(map[string]interface{}),
	}
}

// K returns the size of k-mers (immutable)
func (ks *KmerSet) K() int {
	return ks.k
}

// AddKmerCode adds an encoded k-mer to the set
func (ks *KmerSet) AddKmerCode(kmer uint64) {
	ks.bitmap.Add(kmer)
}

// AddCanonicalKmerCode adds an encoded canonical k-mer to the set
func (ks *KmerSet) AddCanonicalKmerCode(kmer uint64) {
	canonical := CanonicalKmer(kmer, ks.k)
	ks.bitmap.Add(canonical)
}

// AddKmer adds a k-mer to the set by encoding the sequence
// The sequence must have exactly k nucleotides
// Zero-allocation: encodes directly without creating an intermediate slice
func (ks *KmerSet) AddKmer(seq []byte) {
	kmer := EncodeKmer(seq, ks.k)
	ks.bitmap.Add(kmer)
}

// AddCanonicalKmer adds a canonical k-mer to the set by encoding the sequence
// The sequence must have exactly k nucleotides
// Zero-allocation: encodes directly in canonical form without creating an intermediate slice
func (ks *KmerSet) AddCanonicalKmer(seq []byte) {
	canonical := EncodeCanonicalKmer(seq, ks.k)
	ks.bitmap.Add(canonical)
}

// AddSequence adds all k-mers from a sequence to the set
// Uses an iterator to avoid allocating an intermediate vector
func (ks *KmerSet) AddSequence(seq *obiseq.BioSequence) {
	rawSeq := seq.Sequence()
	for canonical := range IterCanonicalKmers(rawSeq, ks.k) {
		ks.bitmap.Add(canonical)
	}
}

// AddSequences adds all k-mers from multiple sequences in batch
func (ks *KmerSet) AddSequences(sequences *obiseq.BioSequenceSlice) {
	for _, seq := range *sequences {
		ks.AddSequence(seq)
	}
}

// Contains checks if a k-mer is in the set
func (ks *KmerSet) Contains(kmer uint64) bool {
	return ks.bitmap.Contains(kmer)
}

// Len returns the number of k-mers in the set
func (ks *KmerSet) Len() uint64 {
	return ks.bitmap.GetCardinality()
}

// MemoryUsage returns memory usage in bytes
func (ks *KmerSet) MemoryUsage() uint64 {
	return ks.bitmap.GetSizeInBytes()
}

// Clear empties the set
func (ks *KmerSet) Clear() {
	ks.bitmap.Clear()
}

// Copy creates a copy of the set (consistent with BioSequence.Copy)
func (ks *KmerSet) Copy() *KmerSet {
	// Copy metadata
	metadata := make(map[string]interface{}, len(ks.Metadata))
	for k, v := range ks.Metadata {
		metadata[k] = v
	}

	return &KmerSet{
		id:       ks.id,
		k:        ks.k,
		bitmap:   ks.bitmap.Clone(),
		Metadata: metadata,
	}
}

// Id returns the identifier of the KmerSet (consistent with BioSequence.Id)
func (ks *KmerSet) Id() string {
	return ks.id
}

// SetId sets the identifier of the KmerSet (consistent with BioSequence.SetId)
func (ks *KmerSet) SetId(id string) {
	ks.id = id
}

// Union returns the union of this set with another
func (ks *KmerSet) Union(other *KmerSet) *KmerSet {
	if ks.k != other.k {
		panic(fmt.Sprintf("Cannot union KmerSets with different k values: %d vs %d", ks.k, other.k))
	}
	result := ks.bitmap.Clone()
	result.Or(other.bitmap)
	return NewKmerSetFromBitmap(ks.k, result)
}

// Intersect returns the intersection of this set with another
func (ks *KmerSet) Intersect(other *KmerSet) *KmerSet {
	if ks.k != other.k {
		panic(fmt.Sprintf("Cannot intersect KmerSets with different k values: %d vs %d", ks.k, other.k))
	}
	result := ks.bitmap.Clone()
	result.And(other.bitmap)
	return NewKmerSetFromBitmap(ks.k, result)
}

// Difference returns the difference of this set with another (this - other)
func (ks *KmerSet) Difference(other *KmerSet) *KmerSet {
	if ks.k != other.k {
		panic(fmt.Sprintf("Cannot subtract KmerSets with different k values: %d vs %d", ks.k, other.k))
	}
	result := ks.bitmap.Clone()
	result.AndNot(other.bitmap)
	return NewKmerSetFromBitmap(ks.k, result)
}

// Iterator returns an iterator over all k-mers in the set
func (ks *KmerSet) Iterator() roaring64.IntIterable64 {
	return ks.bitmap.Iterator()
}

// Bitmap returns the underlying bitmap (for compatibility)
func (ks *KmerSet) Bitmap() *roaring64.Bitmap {
	return ks.bitmap
}
