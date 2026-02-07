package obikmer

import (
	"testing"
)

func TestIterSuperKmers(t *testing.T) {
	seq := []byte("ACGTACGTGGGGAAAA")
	k := 5
	m := 3

	count := 0
	for sk := range IterSuperKmers(seq, k, m) {
		count++
		t.Logf("SuperKmer %d: Minimizer=%d, Start=%d, End=%d, Seq=%s",
			count, sk.Minimizer, sk.Start, sk.End, string(sk.Sequence))

		// Verify sequence boundaries
		if sk.Start < 0 || sk.End > len(seq) {
			t.Errorf("Invalid boundaries: Start=%d, End=%d, seqLen=%d",
				sk.Start, sk.End, len(seq))
		}

		// Verify sequence content
		if string(sk.Sequence) != string(seq[sk.Start:sk.End]) {
			t.Errorf("Sequence mismatch: expected %s, got %s",
				string(seq[sk.Start:sk.End]), string(sk.Sequence))
		}
	}

	if count == 0 {
		t.Error("No super k-mers extracted")
	}

	t.Logf("Total super k-mers extracted: %d", count)
}

func TestIterSuperKmersVsSlice(t *testing.T) {
	seq := []byte("ACGTACGTGGGGAAAAACGTACGT")
	k := 7
	m := 4

	// Extract using slice version
	sliceResult := ExtractSuperKmers(seq, k, m, nil)

	// Extract using iterator version
	var iterResult []SuperKmer
	for sk := range IterSuperKmers(seq, k, m) {
		iterResult = append(iterResult, sk)
	}

	// Compare counts
	if len(sliceResult) != len(iterResult) {
		t.Errorf("Different number of super k-mers: slice=%d, iter=%d",
			len(sliceResult), len(iterResult))
	}

	// Compare each super k-mer
	for i := 0; i < len(sliceResult) && i < len(iterResult); i++ {
		slice := sliceResult[i]
		iter := iterResult[i]

		if slice.Minimizer != iter.Minimizer {
			t.Errorf("SuperKmer %d: different minimizers: slice=%d, iter=%d",
				i, slice.Minimizer, iter.Minimizer)
		}

		if slice.Start != iter.Start || slice.End != iter.End {
			t.Errorf("SuperKmer %d: different boundaries: slice=[%d:%d], iter=[%d:%d]",
				i, slice.Start, slice.End, iter.Start, iter.End)
		}

		if string(slice.Sequence) != string(iter.Sequence) {
			t.Errorf("SuperKmer %d: different sequences: slice=%s, iter=%s",
				i, string(slice.Sequence), string(iter.Sequence))
		}
	}
}

// Note: Tests for ToBioSequence and SuperKmerWorker are in a separate
// integration test package to avoid circular dependencies between
// obikmer and obiseq packages.
