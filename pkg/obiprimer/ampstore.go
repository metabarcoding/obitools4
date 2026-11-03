package obiprimer

import "fmt"

// ampRecord is the logical content of the record of one amplicon: its barcode
// (see ampHash), the taxon of its sequence at the chosen rank (-1 when it has
// none) and the role of the sequence. It is stored packed in an ampStore.
type ampRecord struct {
	hash  uint64
	taxon int32
	role  Role
}

const (
	roleBits = 2
	roleMask = 1<<roleBits - 1

	// maxTaxa is the number of taxa a record can carry: the taxon, shifted by
	// one so that "none" is 0, shares a 32 bits word with the role.
	maxTaxa = 1<<(32-roleBits) - 2
)

func packTaxon(taxon int32, role Role) uint32 { return uint32(taxon+1)<<roleBits | uint32(role) }

func unpackTaxon(t uint32) (int32, Role) { return int32(t>>roleBits) - 1, Role(t & roleMask) }

// ampSlice is the amplicons of one pair: 12 bytes per amplicon, kept as two
// parallel arrays (a struct holding a uint64 and a uint32 would take 16).
type ampSlice struct {
	hash []uint64
	tax  []uint32 // packTaxon
}

// ampPlan says how many amplicons each worker finds for each pair, and which
// worker scans which sequence: the write pass replays the counting pass with
// the very same assignment, so that every worker knows where to write its
// records without sharing any counter with the others.
type ampPlan struct {
	owner []uint16   // owner[s]: the worker that scans sequence s
	count [][]uint32 // count[w][pair]: amplicons of the pair in the sequences of worker w
}

// total returns the number of amplicons of every pair.
func (pl *ampPlan) total() []int {
	t := make([]int, len(pl.count[0]))
	for _, row := range pl.count {
		for b, n := range row {
			t[b] += int(n)
		}
	}
	return t
}

// ampStore holds the amplicons of a batch of pairs, grouped by pair: for each
// pair, two slices of the exact size, since their number is known before they
// are written (see planAmplicons). Nothing grows by doubling, nothing is copied
// nor sorted afterwards. The records of a pair are those of worker 0, then those
// of worker 1, and so on.
//
// The pairs have slices of their own, rather than slices of two huge arrays: a
// few kB each find room in the free memory, where blocks of tens of MB make the
// process map new memory (and, on some systems, count the freed one still).
type ampStore struct {
	plan   *ampPlan
	lo     int // first pair of the batch in the plan
	pairs  []ampSlice
	cursor [][]int // cursor[w][b]: next free slot of worker w for pair b
}

// newAmpStore allocates the store of the pairs lo..hi-1 of the plan.
func newAmpStore(plan *ampPlan, lo, hi int) *ampStore {
	n := hi - lo
	s := &ampStore{plan: plan, lo: lo, pairs: make([]ampSlice, n), cursor: make([][]int, len(plan.count))}
	for w := range s.cursor {
		s.cursor[w] = make([]int, n)
	}
	for b := range n {
		pos := 0
		for w, row := range plan.count {
			s.cursor[w][b] = pos
			pos += int(row[lo+b])
		}
		if pos > 0 {
			s.pairs[b] = ampSlice{make([]uint64, pos), make([]uint32, pos)}
		}
	}
	return s
}

// add writes one record of pair b (numbered in the batch), found by worker w.
func (s *ampStore) add(w int, b int32, hash uint64, taxon uint32) {
	i := s.cursor[w][b]
	s.cursor[w][b] = i + 1
	s.pairs[b].hash[i], s.pairs[b].tax[i] = hash, taxon
}

// check panics if a worker did not write exactly the records counted for it:
// the counting and write passes must enumerate the same amplicons.
func (s *ampStore) check() {
	for b := range s.pairs {
		pos := 0
		for w, row := range s.plan.count {
			pos += int(row[s.lo+b])
			if s.cursor[w][b] != pos {
				panic(fmt.Sprintf("obiprimer: pair %d: worker %d wrote %d amplicons, %d were counted",
					s.lo+b, w, s.cursor[w][b]-(pos-int(row[s.lo+b])), row[s.lo+b]))
			}
		}
	}
}

// pair returns the amplicons of pair b (numbered in the batch).
func (s *ampStore) pair(b int) ampSlice { return s.pairs[b] }

// ampBatches cuts the pairs 0..len(count)-1 in consecutive ranges [lo,hi)
// whose amplicons together do not exceed budget records (a pair alone above the
// budget makes a batch of its own). A non positive budget gives one batch.
func ampBatches(count []int, budget int) [][2]int {
	var out [][2]int
	lo, sum := 0, 0
	for i, n := range count {
		if budget > 0 && i > lo && sum+n > budget {
			out = append(out, [2]int{lo, i})
			lo, sum = i, 0
		}
		sum += n
	}
	if lo < len(count) {
		out = append(out, [2]int{lo, len(count)})
	}
	return out
}
