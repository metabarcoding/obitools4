package obisimilarity

import "sort"

// Member is a copy of the entry of a sequence in an Index, with its own table.
type Member struct {
	Entry
	ID   int // identifier given by the caller
	Rank int // rank in the order of insertion
}

const tabChunk = 1024

// A Bucket holds the members of one length, in the order of insertion, and
// their tables by blocks (see tableBlock).
type Bucket struct {
	Members []Member
	blocks  []tableBlock
	width   int
}

// Commons caches the result of the filter on one block of a bucket.
type Commons struct {
	v   [64]uint8
	blk int
}

// Reset must be called when changing of bucket.
func (c *Commons) Reset() { c.blk = -1 }

// Common returns the sum of the minima of the 4-mer counters of the sequence
// whose Sparse words are given and of the member k of the bucket (without the
// overflow term), or -1 if it is not available, in which case PairScore
// computes it.
func (b *Bucket) Common(k int, sparse []uint16, c *Commons) int {
	if b.width == 0 || sparse == nil {
		return -1
	}
	if blk := k / b.width; blk != c.blk {
		b.blocks[blk].common(b.width, sparse, c.v[:])
		c.blk = blk
	}
	return int(c.v[k%b.width])
}

// Index stores sequences by length. A sequence can only be similar to the ones
// whose length is compatible with its own (see Scoring.LengthCompatible), so
// it only visits the buckets of these lengths, instead of all the sequences,
// and the filter works on blocks of sequences at once.
//
// An Index can grow, but not while it is read.
type Index struct {
	width int // sequences per block of transposed tables, 0 if unused
	byLen map[int]*Bucket
	lens  []int      // lengths that have a bucket, increasing
	tabs  [][]Table4 // chunks holding the tables of the members
	n     int
}

// NewIndex returns an empty index.
func NewIndex() *Index {
	return &Index{width: simdBlockWidth(), byLen: make(map[int]*Bucket)}
}

// Len is the number of sequences in the index.
func (x *Index) Len() int { return x.n }

// Lengths returns the lengths that have a bucket, in increasing order.
func (x *Index) Lengths() []int { return x.lens }

// Bucket returns the bucket of the sequences of the given length.
func (x *Index) Bucket(length int) *Bucket { return x.byLen[length] }

// Add stores a copy of the entry e as the next member of the index, with the
// identifier id, and returns it.
func (x *Index) Add(e Entry, id int) Member {
	if x.n%tabChunk == 0 {
		x.tabs = append(x.tabs, make([]Table4, tabChunk))
	}
	tab := &x.tabs[len(x.tabs)-1][x.n%tabChunk]
	if e.Table != nil {
		*tab = *e.Table
	}
	e.Table = tab

	b := x.byLen[e.Length]
	if b == nil {
		b = &Bucket{width: x.width}
		x.byLen[e.Length] = b
		x.lens = append(x.lens, e.Length)
		sort.Ints(x.lens)
	}
	if w := x.width; w > 0 && !e.Ambiguous {
		if len(b.Members)%w == 0 {
			b.blocks = append(b.blocks, newTableBlock(w))
		}
		b.blocks[len(b.blocks)-1].set(w, len(b.Members)%w, tab)
	} else if w > 0 && len(b.Members)%w == 0 {
		b.blocks = append(b.blocks, newTableBlock(w))
	}

	m := Member{Entry: e, ID: id, Rank: x.n}
	b.Members = append(b.Members, m)
	x.n++
	return m
}
