package obiclust

import (
	"math/rand"
	"testing"
)

// commonWordsRef is the definition of commonWords, spelled out.
func commonWordsRef(t1, t2 *table4) int {
	total := 0
	for i := range t1.count {
		total += int(min(t1.count[i], t2.count[i]))
	}
	return total + min(t1.overflow, t2.overflow)
}

func TestCommonWords(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	seq := func(n int, alpha string) []byte {
		s := make([]byte, n)
		for i := range s {
			s[i] = alpha[rng.Intn(len(alpha))]
		}
		return s
	}
	for it := 0; it < 3000; it++ {
		alpha := []string{"acgt", "ac", "a"}[it%3]
		t1 := buildTable4(seq(rng.Intn(1200), alpha))
		t2 := buildTable4(seq(rng.Intn(1200), alpha))
		if got, want := commonWords(t1, t2), commonWordsRef(t1, t2); got != want {
			t.Fatalf("commonWords = %d, want %d", got, want)
		}
	}
}

// TestBuildTable4 checks the table against the 4-mers counted the long way,
// including the saturation of the counters.
func TestBuildTable4(t *testing.T) {
	code := map[byte]int{'a': 0, 'c': 1, 't': 2, 'g': 3}
	rng := rand.New(rand.NewSource(4))
	for it := 0; it < 500; it++ {
		alpha := []string{"acgt", "a", "ac"}[it%3]
		n := rng.Intn(900)
		s := make([]byte, n)
		for i := range s {
			s[i] = alpha[rng.Intn(len(alpha))]
		}
		var counts [256]int
		for i := 0; i+4 <= n; i++ {
			w := 0
			for _, c := range s[i : i+4] {
				w = w<<2 | code[c]
			}
			counts[w]++
		}
		tab := buildTable4(s)
		over := 0
		for w, c := range counts {
			if c > 255 {
				over += c - 255
				c = 255
			}
			if int(tab.count[w]) != c {
				t.Fatalf("word %d: got %d, want %d", w, tab.count[w], c)
			}
		}
		if tab.overflow != over {
			t.Fatalf("overflow = %d, want %d", tab.overflow, over)
		}
	}
}
