package obigraph

import (
	"context"
	"math/bits"
	"math/rand"
	"testing"
	"time"
)

// adjacency matrix filter for small graphs
func matrixFilter(adj [][]bool) NeighborFilter {
	return func(v int32, candidates, out []int32) []int32 {
		for _, c := range candidates {
			if adj[v][c] {
				out = append(out, c)
			}
		}
		return out
	}
}

func randomGraph(rng *rand.Rand, n int, p float64) [][]bool {
	adj := make([][]bool, n)
	for i := range adj {
		adj[i] = make([]bool, n)
	}
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if rng.Float64() < p {
				adj[i][j], adj[j][i] = true, true
			}
		}
	}
	return adj
}

// maximum clique size by brute force over all the subsets (n <= 20)
func bruteMaxClique(adj [][]bool) int {
	n := len(adj)
	best := 0
	for mask := 1; mask < 1<<n; mask++ {
		size := bits.OnesCount(uint(mask))
		if size <= best {
			continue
		}
		ok := true
		for i := 0; i < n && ok; i++ {
			if mask>>i&1 == 0 {
				continue
			}
			for j := i + 1; j < n; j++ {
				if mask>>j&1 == 1 && !adj[i][j] {
					ok = false
					break
				}
			}
		}
		if ok {
			best = size
		}
	}
	return best
}

func isClique(adj [][]bool, c []int32) bool {
	seen := map[int32]bool{}
	for i, a := range c {
		if seen[a] {
			return false
		}
		seen[a] = true
		for _, b := range c[i+1:] {
			if !adj[a][b] {
				return false
			}
		}
	}
	return true
}

func TestFindCliqueAgainstBruteForce(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for trial := 0; trial < 40; trial++ {
		n := 8 + rng.Intn(10)
		adj := randomGraph(rng, n, 0.3+0.5*rng.Float64())
		maxSize := bruteMaxClique(adj)
		filter := matrixFilter(adj)

		for _, workers := range []int{1, 3} {
			r := FindClique(context.Background(), n, filter, maxSize, CliqueOptions{Workers: workers})
			if !r.Reached || len(r.Clique) < maxSize || !isClique(adj, r.Clique) {
				t.Fatalf("trial %d: target %d not reached: %+v", trial, maxSize, r)
			}

			r = FindClique(context.Background(), n, filter, maxSize+1, CliqueOptions{Workers: workers})
			if r.Reached || !r.Exhausted {
				t.Fatalf("trial %d: size %d cannot exist, got %+v", trial, maxSize+1, r)
			}
			// best effort: a valid clique, not necessarily a maximum one
			if len(r.Clique) < 1 || len(r.Clique) > maxSize || !isClique(adj, r.Clique) {
				t.Fatalf("trial %d: invalid best clique %v (maximum is %d)", trial, r.Clique, maxSize)
			}
		}
	}
}

func TestFindCliqueTrivialTargets(t *testing.T) {
	adj := randomGraph(rand.New(rand.NewSource(2)), 5, 0.5)
	if r := FindClique(context.Background(), 5, matrixFilter(adj), 0, CliqueOptions{}); !r.Reached {
		t.Fatal("target 0 is always reached")
	}
	if r := FindClique(context.Background(), 0, matrixFilter(adj), 1, CliqueOptions{}); r.Reached {
		t.Fatal("no clique in an empty graph")
	}
	if r := FindClique(context.Background(), 5, matrixFilter(adj), 1, CliqueOptions{}); !r.Reached || len(r.Clique) != 1 {
		t.Fatalf("a single vertex is a clique: %+v", r)
	}
}

// A hopeless search must stop with the deadline and return a valid clique.
// (When the target is far above the clique number of a sparse graph, pruning
// proves the absence at once: see TestFindCliqueAgainstBruteForce.)
func TestFindCliqueTimeout(t *testing.T) {
	// dense graph: its largest clique is far below the target, but pruning only
	// bites after many levels, so the search cannot finish within the deadline
	const n = 300
	adj := randomGraph(rand.New(rand.NewSource(3)), n, 0.9)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	start := time.Now()
	r := FindClique(ctx, n, matrixFilter(adj), 150, CliqueOptions{Workers: 2})
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Fatalf("deadline not honoured: %v", elapsed)
	}
	if r.Reached || r.Exhausted {
		t.Fatalf("size 150 cannot be reached nor proven absent in time: %+v", r)
	}
	if len(r.Clique) < 2 || !isClique(adj, r.Clique) {
		t.Fatalf("invalid best clique %v", r.Clique)
	}
}

func TestFindCliqueProgress(t *testing.T) {
	adj := randomGraph(rand.New(rand.NewSource(4)), 40, 0.7)

	var seen []int
	r := FindClique(context.Background(), 40, matrixFilter(adj), 5, CliqueOptions{
		Progress: func(best int) { seen = append(seen, best) },
	})
	if !r.Reached {
		t.Fatalf("size 5 should be reached: %+v", r)
	}
	if len(seen) == 0 || seen[len(seen)-1] != len(r.Clique) {
		t.Fatalf("progress %v does not end at the clique size %d", seen, len(r.Clique))
	}
	for i := 1; i < len(seen); i++ {
		if seen[i] <= seen[i-1] {
			t.Fatalf("progress must be strictly increasing: %v", seen)
		}
	}
}

// The improving search must end with a proven maximum clique, even when the
// target is out of reach; the focused one only proves the absence of a clique
// of the target size.
func TestFindCliqueModes(t *testing.T) {
	rng := rand.New(rand.NewSource(5))
	for trial := 0; trial < 40; trial++ {
		n := 8 + rng.Intn(10)
		adj := randomGraph(rng, n, 0.3+0.5*rng.Float64())
		maxSize := bruteMaxClique(adj)
		filter := matrixFilter(adj)

		for _, workers := range []int{1, 3} {
			for _, limit := range []int{0, 5, 1 << 20} {
				r := FindClique(context.Background(), n, filter, maxSize+1,
					CliqueOptions{Workers: workers, ColorLimit: limit, mode: modeImproving})
				if r.Reached || !r.Exhausted || !r.Maximum {
					t.Fatalf("trial %d limit %d: improving search should prove the maximum: %+v", trial, limit, r)
				}
				if len(r.Clique) != maxSize || !isClique(adj, r.Clique) {
					t.Fatalf("trial %d limit %d: expected a maximum clique of %d, got %v", trial, limit, maxSize, r.Clique)
				}

				r = FindClique(context.Background(), n, filter, maxSize+1,
					CliqueOptions{Workers: workers, ColorLimit: limit, mode: modeFocused})
				if r.Reached || !r.Exhausted || r.Maximum {
					t.Fatalf("trial %d limit %d: focused search proves absence only: %+v", trial, limit, r)
				}

				r = FindClique(context.Background(), n, filter, maxSize,
					CliqueOptions{Workers: workers, ColorLimit: limit, mode: modeImproving})
				if !r.Reached || len(r.Clique) < maxSize || !isClique(adj, r.Clique) {
					t.Fatalf("trial %d limit %d: target %d not reached by the improving search: %+v", trial, limit, maxSize, r)
				}
			}
		}
	}
}

// When the target is out of reach, the default search must end with a proven
// maximum clique, even if the coloring bound proves the absence of the target
// before any clique has been found.
func TestFindCliqueUnreachableTargetGivesMaximum(t *testing.T) {
	rng := rand.New(rand.NewSource(6))
	for trial := 0; trial < 40; trial++ {
		n := 8 + rng.Intn(10)
		adj := randomGraph(rng, n, 0.3+0.5*rng.Float64())
		maxSize := bruteMaxClique(adj)

		for _, workers := range []int{1, 3} {
			for _, limit := range []int{0, 5, 1 << 20} {
				r := FindClique(context.Background(), n, matrixFilter(adj), maxSize+1,
					CliqueOptions{Workers: workers, ColorLimit: limit})
				if r.Reached || !r.Exhausted || !r.Maximum {
					t.Fatalf("trial %d workers %d limit %d: expected a proven maximum: %+v", trial, workers, limit, r)
				}
				if len(r.Clique) != maxSize || !isClique(adj, r.Clique) {
					t.Fatalf("trial %d workers %d limit %d: expected a maximum clique of %d, got %v",
						trial, workers, limit, maxSize, r.Clique)
				}
			}
		}
	}
}

// A long coloring must not delay the deadline.
func TestFindCliqueColoringHonoursDeadline(t *testing.T) {
	const n = 3000
	adj := randomGraph(rand.New(rand.NewSource(7)), n, 0.5)
	base := matrixFilter(adj)
	slow := func(v int32, candidates, out []int32) []int32 {
		time.Sleep(200 * time.Microsecond)
		return base(v, candidates, out)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	start := time.Now()
	r := FindClique(ctx, n, slow, 100, CliqueOptions{Workers: 2, ColorLimit: 1 << 20})
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("deadline not honoured by the coloring: %v", elapsed)
	}
	if r.Reached || !isClique(adj, r.Clique) {
		t.Fatalf("unexpected result %+v", r)
	}
}
