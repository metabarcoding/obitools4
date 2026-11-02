package obigraph

import (
	"context"
	"math/rand"
	"sync"
)

// NeighborFilter defines an implicit graph over the vertices 0..n-1.
// It appends to out[:0] the vertices of candidates that are adjacent to v,
// keeping their order, and returns the resulting slice. It is called
// concurrently by several workers and must not modify candidates.
//
// Describing the adjacency by a filter, instead of a matrix, keeps the memory
// in O(n) and lets the caller evaluate adjacency in bulk (SIMD, for instance).
type NeighborFilter func(v int32, candidates, out []int32) []int32

// CliqueOptions tunes FindClique.
type CliqueOptions struct {
	// Workers is the number of concurrent searches (default 1).
	Workers int
	// Seed seeds the random vertex orders of the restarts.
	Seed int64
	// Progress, if set, is called with the size of the best clique each time a
	// larger one is found. Calls are serialized and made while a lock is held:
	// it must return quickly and must not call FindClique.
	Progress func(best int)
}

// CliqueResult is the outcome of FindClique.
type CliqueResult struct {
	// Clique is the largest clique found (a clique of size >= target if Reached).
	Clique []int32
	// Reached is true if len(Clique) >= target.
	Reached bool
	// Exhausted is true if the whole search space has been explored. If Reached
	// is false, it proves that no clique of the target size exists.
	Exhausted bool
}

// FindClique looks for a clique of at least target vertices in the graph of n
// vertices described by filter. It stops as soon as one is found, when the
// search space is exhausted, or when ctx is done; in the last case, the largest
// clique met so far is returned.
//
// The search is a depth first search with backtracking, pruned when the
// current clique plus the remaining candidates cannot reach target. The first
// pass of the first worker follows the natural order of the vertices (greedy
// "lexicographic" construction); every other pass restarts from a random
// order, with a work budget that doubles at each restart, so that the search
// is complete in the long run.
func FindClique(ctx context.Context, n int, filter NeighborFilter, target int, opts CliqueOptions) CliqueResult {
	if target <= 0 {
		return CliqueResult{Reached: true, Exhausted: true}
	}
	if n <= 0 {
		return CliqueResult{Exhausted: true}
	}

	workers := max(opts.Workers, 1)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		mutex     sync.Mutex
		best      []int32
		exhausted bool
		wg        sync.WaitGroup
	)

	publish := func(clique []int32) {
		mutex.Lock()
		if len(clique) > len(best) {
			best = append(best[:0], clique...)
			if opts.Progress != nil {
				opts.Progress(len(best))
			}
		}
		mutex.Unlock()
	}

	for w := range workers {
		wg.Go(func() {
			s := &cliqueSearch{
				ctx:    ctx,
				filter: filter,
				target: target,
				publish: func(c []int32) {
					publish(c)
					if len(c) >= target {
						cancel()
					}
				},
			}
			rng := rand.New(rand.NewSource(opts.Seed + int64(w)))
			candidates := make([]int32, n)
			for i := range candidates {
				candidates[i] = int32(i)
			}

			for restart := 0; ctx.Err() == nil; restart++ {
				if w > 0 || restart > 0 {
					rng.Shuffle(n, func(i, j int) { candidates[i], candidates[j] = candidates[j], candidates[i] })
				}
				s.budget = int64(n) * int64(target) << min(restart, 20)
				s.work = 0
				s.aborted = false
				s.search(0, candidates)

				if !s.aborted && ctx.Err() == nil {
					mutex.Lock()
					exhausted = true
					mutex.Unlock()
					cancel()
				}
			}
		})
	}
	wg.Wait()

	return CliqueResult{
		Clique:    best,
		Reached:   len(best) >= target,
		Exhausted: exhausted,
	}
}

type cliqueSearch struct {
	ctx     context.Context
	filter  NeighborFilter
	target  int
	publish func([]int32)

	current []int32   // the clique under construction
	buffers [][]int32 // candidate list of each depth
	bestLen int

	budget  int64 // maximum work (number of candidates tested) for this restart
	work    int64
	aborted bool // budget or context exhausted before the end of the exploration
}

func (s *cliqueSearch) search(depth int, candidates []int32) {
	for i := range candidates {
		if depth+len(candidates)-i < s.target {
			return
		}
		if s.work > s.budget || s.ctx.Err() != nil {
			s.aborted = true
			return
		}

		v := candidates[i]
		s.current = append(s.current[:depth], v)
		if len(s.current) > s.bestLen {
			s.bestLen = len(s.current)
			s.publish(s.current)
			if s.bestLen >= s.target {
				s.aborted = true // the search is over: do not report an exhaustion
				return
			}
		}

		rest := candidates[i+1:]
		if depth+1+len(rest) < s.target {
			continue
		}
		for len(s.buffers) <= depth+1 {
			s.buffers = append(s.buffers, nil)
		}
		next := s.filter(v, rest, s.buffers[depth+1][:0])
		s.buffers[depth+1] = next
		s.work += int64(len(rest))

		if depth+1+len(next) < s.target {
			continue
		}
		s.search(depth+1, next)
		if s.aborted {
			return
		}
	}
}
