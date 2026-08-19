package obialign

import (
	"math/rand"
	"testing"
)

func TestLocatePatternNormal(t *testing.T) {
	// Pattern fully and exactly present in the middle of a longer sequence.
	start, end, nerr := LocatePattern("id", []byte("ACGT"), []byte("TTTTACGTTTTT"))
	if start != 4 || end != 8 || nerr != 0 {
		t.Errorf("got start=%d end=%d nerr=%d, want start=4 end=8 nerr=0", start, end, nerr)
	}
}

func TestLocatePatternOneMismatch(t *testing.T) {
	start, end, nerr := LocatePattern("id", []byte("ACGT"), []byte("TTTTACTTTTTT"))
	if nerr != 1 {
		t.Errorf("got nerr=%d, want 1 (start=%d end=%d)", nerr, start, end)
	}
}

// The real-world case that used to panic: pattern longer than the sequence
// fragment extracted for indel relocation.
func TestLocatePatternPatternLongerThanSequence(t *testing.T) {
	start, end, nerr := LocatePattern("id",
		[]byte("GGGCAATCCTGAGCCAAATC"),
		[]byte("tcctgagccaaatcacgtt"))

	if start != -1 || end != -1 || nerr != -1 {
		t.Errorf("got start=%d end=%d nerr=%d, want the (-1,-1,-1) sentinel", start, end, nerr)
	}
}

func TestLocatePatternSequenceLengthOne(t *testing.T) {
	start, end, nerr := LocatePattern("id", []byte("AB"), []byte("A"))

	if start < 0 || end < 0 || nerr < 0 {
		t.Fatalf("got start=%d end=%d nerr=%d, expected a valid (non-sentinel) result", start, end, nerr)
	}
	if start != 0 || end != 1 || nerr != 1 {
		t.Errorf("got start=%d end=%d nerr=%d, want start=0 end=1 nerr=1", start, end, nerr)
	}
}

// A pattern much longer than the sequence can still yield a mathematically
// valid (in-bounds) alignment: the extra pattern length is absorbed as gaps,
// driving the error count high enough that the caller's maxerr threshold
// rejects it. The function itself must still return consistent bounds.
func TestLocatePatternPatternMuchLongerThanSequence(t *testing.T) {
	start, end, nerr := LocatePattern("id", []byte("ACGTACGTACGTACGTACGT"), []byte("ACG"))

	isSentinel := start == -1 && end == -1 && nerr == -1
	isValid := start >= 0 && end > start && end <= 3 && nerr >= 0

	if !isSentinel && !isValid {
		t.Errorf("got start=%d end=%d nerr=%d, want either the sentinel or consistent in-bounds values", start, end, nerr)
	}
}

func TestLocatePatternNeverReturnsOutOfBounds(t *testing.T) {
	patterns := []string{"A", "AC", "ACG", "ACGT", "ACGTA", "ACGTAC", "ACGTACG", "ACGTACGT"}
	sequences := []string{"A", "AC", "ACG", "ACGT", "ACGTA", "ACGTAC", "ACGTACG", "ACGTACGT"}

	for _, p := range patterns {
		for _, s := range sequences {
			start, end, nerr := LocatePattern("id", []byte(p), []byte(s))

			if start == -1 && end == -1 && nerr == -1 {
				// Explicit "no reliable match" sentinel: always acceptable.
				continue
			}

			if start < 0 || end < 0 || start >= end || end > len(s) || nerr < 0 {
				t.Errorf("pattern=%q sequence=%q -> start=%d end=%d nerr=%d is out of bounds / inconsistent",
					p, s, start, end, nerr)
			}
		}
	}
}

// Randomized property test over a wide range of pattern/sequence length
// combinations, including pattern >= sequence, to make sure the function
// never panics and never returns anything but the sentinel or fully
// consistent, in-bounds coordinates.
func TestLocatePatternRandomizedNeverInvalid(t *testing.T) {
	const bases = "ACGT"
	rng := rand.New(rand.NewSource(42))

	randSeq := func(n int) []byte {
		b := make([]byte, n)
		for i := range b {
			b[i] = bases[rng.Intn(len(bases))]
		}
		return b
	}

	for trial := 0; trial < 5000; trial++ {
		patLen := 1 + rng.Intn(15)
		seqLen := 1 + rng.Intn(15)

		pattern := randSeq(patLen)
		sequence := randSeq(seqLen)

		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panic for pattern=%q sequence=%q: %v", pattern, sequence, r)
				}
			}()

			start, end, nerr := LocatePattern("id", pattern, sequence)

			isSentinel := start == -1 && end == -1 && nerr == -1
			isValid := start >= 0 && end > start && end <= len(sequence) && nerr >= 0

			if !isSentinel && !isValid {
				t.Errorf("pattern=%q sequence=%q -> start=%d end=%d nerr=%d is neither the sentinel nor consistent",
					pattern, sequence, start, end, nerr)
			}
		}()
	}
}
