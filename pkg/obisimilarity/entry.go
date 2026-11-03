package obisimilarity

import (
	"sync"
	"sync/atomic"
)

// Entry is a distinct sequence of a dataset, with what the comparisons need.
type Entry struct {
	Seq    []byte
	Length int
	Count  int // count of the sequence (cumulated for identical sequences)
	Table  *Table4
	Sparse []uint16 // words of the table (see SparseWords), nil if unusable
	// Ambiguous sequences contain IUPAC letters other than a, c, g and t: they
	// have no 4-mer table, and the filter does not apply to them.
	Ambiguous bool
}

const (
	notNucleotide = iota
	plainNucleotide
	ambiguousNucleotide
)

var nucleotideClass = func() (t [256]uint8) {
	for _, c := range "acgt" {
		t[c] = plainNucleotide
	}
	for _, c := range "umrwsykvhdbn" {
		t[c] = ambiguousNucleotide
	}
	return t
}()

// Classify tells if a sequence (in lower case) can be compared. It returns
// ok=false for an empty sequence and for a sequence with anything else than
// IUPAC letters; IUPAC letters other than a, c, g and t are only accepted if
// acceptAmbiguous is set, and then ambiguous is true.
func Classify(s []byte, acceptAmbiguous bool) (ok, ambiguous bool) {
	if len(s) == 0 {
		return false, false
	}
	for _, c := range s {
		switch nucleotideClass[c] {
		case notNucleotide:
			return false, false
		case ambiguousNucleotide:
			if !acceptAmbiguous {
				return false, false
			}
			ambiguous = true
		}
	}
	return true, ambiguous
}

// BuildEntries describes n sequences, their 4-mer tables being computed in
// parallel. get returns the sequence i, its count, and whether it is
// ambiguous. The tables are stored contiguously.
func BuildEntries(n int, get func(i int) (seq []byte, count int, ambiguous bool), workers int) []Entry {
	entries := make([]Entry, n)
	tables := make([]Table4, n)

	var next atomic.Int64
	var wg sync.WaitGroup
	for w := 0; w < max(workers, 1); w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				i := int(next.Add(1)) - 1
				if i >= n {
					return
				}
				seq, count, ambiguous := get(i)
				var sparse []uint16
				if !ambiguous {
					tables[i].Fill(seq)
					sparse = SparseWords(&tables[i])
				}
				entries[i] = Entry{
					Seq:       seq,
					Length:    len(seq),
					Count:     count,
					Table:     &tables[i],
					Sparse:    sparse,
					Ambiguous: ambiguous,
				}
			}
		}()
	}
	wg.Wait()
	return entries
}
