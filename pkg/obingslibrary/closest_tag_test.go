package obingslibrary

import (
	"math"
	"math/rand"
	"sync"
	"testing"
)

// closestOverPairs is the former search: every pair of tags of the samples, and so
// every tag as many times as it is used.
func closestOverPairs(marker *Marker, tag string, dist func(string, string) int, forward bool) (string, int) {
	mindist, mintag := math.MaxInt, ""
	for ts := range marker.samples {
		t := ts.Reverse
		if forward {
			t = ts.Forward
		}
		d := dist(t, tag)
		if d == mindist && mintag != "" && t != mintag {
			mintag = ""
		}
		if d < mindist {
			mindist, mintag = d, t
		}
	}
	return mintag, mindist
}

func randomWord(rng *rand.Rand, alphabet string, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = alphabet[rng.Intn(len(alphabet))]
	}
	return string(b)
}

// TestClosestTagAgainstAllPairs: searching the distinct tags gives the same tag and
// the same distance as searching every pair, ties (no tag) included, for both
// distances. Few letters make the ties frequent; the tags come in many samples.
func TestClosestTagAgainstAllPairs(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	ties, exact := 0, 0
	for iter := 0; iter < 3000; iter++ {
		marker := &Marker{samples: make(map[TagPair]*PCR)}
		alphabet := []string{"ac", "acg", "acgt"}[rng.Intn(3)]
		length := 3 + rng.Intn(5)
		var tagsF, tagsR []string
		for i, n := 0, 1+rng.Intn(6); i < n; i++ {
			tagsF = append(tagsF, randomWord(rng, alphabet, length))
			tagsR = append(tagsR, randomWord(rng, alphabet, length))
		}
		for i := 0; i < 1+rng.Intn(40); i++ { // many samples, the same tags again and again
			marker.GetPCR(tagsF[rng.Intn(len(tagsF))], tagsR[rng.Intn(len(tagsR))])
		}
		for q := 0; q < 8; q++ {
			observed := randomWord(rng, alphabet, length+rng.Intn(3)-1)
			if rng.Intn(4) == 0 {
				observed = tagsF[rng.Intn(len(tagsF))]
			}
			for _, dist := range []func(string, string) int{Hamming, Levenshtein} {
				gf, gdf := marker.ClosestForwardTag(observed, dist)
				wf, wdf := closestOverPairs(marker, observed, dist, true)
				gr, gdr := marker.ClosestReverseTag(observed, dist)
				wr, wdr := closestOverPairs(marker, observed, dist, false)
				if gf != wf || gdf != wdf || gr != wr || gdr != wdr {
					t.Fatalf("observed %q: forward %q,%d want %q,%d; reverse %q,%d want %q,%d (samples %d)",
						observed, gf, gdf, wf, wdf, gr, gdr, wr, wdr, len(marker.samples))
				}
				if gf == "" && gdf != math.MaxInt {
					ties++
				}
				if gdf == 0 {
					exact++
				}
			}
		}
	}
	if ties < 200 || exact < 500 {
		t.Errorf("the draws are not varied enough: %d ties, %d exact tags", ties, exact)
	}
}

// TestClosestTagAfterNewSample: a sample added once the tags have been listed is seen.
func TestClosestTagAfterNewSample(t *testing.T) {
	marker := &Marker{samples: make(map[TagPair]*PCR)}
	marker.GetPCR("aaaa", "cccc")
	if tag, d := marker.ClosestForwardTag("aaat", Hamming); tag != "aaaa" || d != 1 {
		t.Fatalf("got %q,%d", tag, d)
	}
	marker.GetPCR("aaat", "cccc")
	if tag, d := marker.ClosestForwardTag("aaat", Hamming); tag != "aaat" || d != 0 {
		t.Errorf("a new tag is not found: %q,%d", tag, d)
	}
}

// TestClosestTagConcurrent (with -race): the tags are listed on the first use, by
// whichever goroutine comes first.
func TestClosestTagConcurrent(t *testing.T) {
	marker := &Marker{samples: make(map[TagPair]*PCR)}
	for _, f := range []string{"aacc", "acac", "ccaa"} {
		for _, r := range []string{"ttgg", "tgtg"} {
			marker.GetPCR(f, r)
		}
	}
	var wg sync.WaitGroup
	for g := 0; g < 16; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				if tag, d := marker.ClosestForwardTag("acaa", Hamming); tag != "" && d != 1 {
					t.Errorf("forward %q,%d", tag, d)
				}
				marker.ClosestReverseTag("ttgg", Levenshtein)
			}
		}()
	}
	wg.Wait()
}

// levenshteinRef is the former implementation, with its two rows allocated.
func levenshteinRef(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}
	prev := make([]int, len(s2)+1)
	curr := make([]int, len(s2)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(s1); i++ {
		curr[0] = i
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			curr[j] = min(prev[j]+1, curr[j-1]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(s2)]
}

// TestLevenshteinNoAllocation: same distances as before, for the empty strings, the
// tags, and the strings longer than the rows kept on the stack.
func TestLevenshteinNoAllocation(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	for iter := 0; iter < 5000; iter++ {
		a := randomWord(rng, "acgt", []int{0, 1, 5, 7, 63, 64, 65, 100}[rng.Intn(8)])
		b := randomWord(rng, "acgt", []int{0, 1, 5, 7, 63, 64, 65, 100}[rng.Intn(8)])
		if got, want := Levenshtein(a, b), levenshteinRef(a, b); got != want {
			t.Fatalf("%q %q: %d, want %d", a, b, got, want)
		}
	}
	if n := testing.AllocsPerRun(100, func() { Levenshtein("aacaaca", "aaccaca") }); n != 0 {
		t.Errorf("Levenshtein of two tags allocates %v times", n)
	}
}
