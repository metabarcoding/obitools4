package obigraph

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
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

	// ColorLimit enables the coloring bound for the candidate lists of at most
	// ColorLimit vertices (0 disables it). The vertices of a list are split into
	// classes of pairwise non-adjacent vertices: a clique takes at most one
	// vertex per class, so the number of classes bounds its size. Computing the
	// classes costs up to the square of the list length in calls to the filter,
	// which is why only short lists are colored.
	ColorLimit int

	// mode forces the kind of search (tests only): 0 alternates, modeFocused
	// and modeImproving restrict the search to one of them.
	mode int
}

const (
	modeFocused   = 1
	modeImproving = 2
)

// CliqueResult is the outcome of FindClique.
type CliqueResult struct {
	// Clique is the largest clique found (a clique of size >= target if Reached).
	Clique []int32
	// Reached is true if len(Clique) >= target.
	Reached bool
	// Exhausted is true if the whole search space has been explored. If Reached
	// is false, it proves that no clique of the target size exists.
	Exhausted bool
	// Maximum is true if Clique is proven to be a maximum clique of the graph:
	// no larger clique exists.
	Maximum bool
}

// FindClique looks for a clique of at least target vertices in the graph of n
// vertices described by filter. It stops as soon as one is found, when the
// search space is exhausted, or when ctx is done; in the last case, the largest
// clique met so far is returned.
//
// The search is a depth first search with backtracking, which comes in two
// flavours that the restarts alternate:
//
//   - the focused search prunes every branch that cannot reach target, which
//     is the most efficient way to reach it but gives poor cliques when target
//     is out of reach;
//   - the improving search only prunes the branches that cannot beat the best
//     clique found so far (by any worker), so it keeps improving that clique
//     whatever target is. When it ends, the best clique is a maximum one.
//
// When a focused search proves that no clique of target vertices exists, every
// search turns to the improving flavour, so that the largest clique is returned
// (and proved to be so) instead of an arbitrary one.
//
// On short candidate lists (see CliqueOptions.ColorLimit), a coloring bound
// prunes the branches that cannot be large enough, and orders the vertices so
// that the most constrained ones are tried first.
//
// The first pass of the first worker follows the natural order of the vertices
// (greedy "lexicographic" construction); every other pass restarts from a
// random order, with a work budget that doubles at each restart, so that the
// search is complete in the long run.
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
		bestSize  atomic.Int32
		exhausted bool
		maximum   bool
		wg        sync.WaitGroup
		// set when a focused search has proved that target cannot be reached:
		// every search then turns to finding the largest clique
		unreachable atomic.Bool
	)

	publish := func(clique []int32) {
		mutex.Lock()
		if len(clique) > len(best) {
			best = append(best[:0], clique...)
			bestSize.Store(int32(len(best)))
			if opts.Progress != nil {
				opts.Progress(len(best))
			}
		}
		mutex.Unlock()
	}

	for w := range workers {
		wg.Go(func() {
			s := &cliqueSearch{
				ctx:         ctx,
				filter:      filter,
				target:      target,
				bestSize:    &bestSize,
				colorLimit:  opts.ColorLimit,
				unreachable: &unreachable,
				publish: func(c []int32) {
					publish(c)
					if len(c) >= target {
						cancel()
					}
				},
			}
			rng := rand.New(rand.NewSource(opts.Seed + int64(w)))
			// the search starts from the end of the list: lowest vertex first
			candidates := make([]int32, n)
			for i := range candidates {
				candidates[i] = int32(n - 1 - i)
			}

			for restart := 0; ctx.Err() == nil; restart++ {
				if w > 0 || restart > 0 {
					rng.Shuffle(n, func(i, j int) { candidates[i], candidates[j] = candidates[j], candidates[i] })
				}
				switch opts.mode {
				case modeFocused:
					s.improving = false
				case modeImproving:
					s.improving = true
				default:
					s.improving = unreachable.Load() || (w+restart)%2 == 1
				}
				s.budget = int64(n) * int64(target) << min(restart, 20)
				s.work = 0
				s.aborted = false
				s.search(0, candidates)

				if !s.aborted && ctx.Err() == nil {
					mutex.Lock()
					exhausted = true
					if s.improving {
						maximum = true
					}
					mutex.Unlock()

					if s.improving || opts.mode == modeFocused {
						cancel()
					} else {
						// target is out of reach: look for the largest clique
						unreachable.Store(true)
					}
				}
			}
		})
	}
	wg.Wait()

	return CliqueResult{
		Clique:    best,
		Reached:   len(best) >= target,
		Exhausted: exhausted,
		Maximum:   maximum,
	}
}

type cliqueSearch struct {
	ctx        context.Context
	filter     NeighborFilter
	target     int
	bestSize   *atomic.Int32 // size of the best clique found by any worker
	publish    func([]int32)
	colorLimit int

	unreachable *atomic.Bool // target is proved out of reach

	improving bool // see FindClique: the kind of the current pass

	current []int32   // the clique under construction
	buffers [][]int32 // candidate list of each depth
	sorted  [][]int32 // candidate list of each depth, sorted by color
	colors  [][]int32 // color (from 1) of each vertex of sorted
	classes [][]int32 // scratch space of the coloring
	probe   []int32   // scratch space of the coloring

	budget  int64 // maximum work (number of candidates tested) for this restart
	work    int64
	aborted bool // budget or context exhausted before the end of the exploration
}

// threshold is the size a clique must reach to be worth exploring: the target
// for the focused search, one more than the best clique for the improving one.
func (s *cliqueSearch) threshold() int {
	if s.improving {
		return min(s.target, int(s.bestSize.Load())+1)
	}
	return s.target
}

// mustStop tells whether the current pass has to be interrupted: its work
// budget is spent, or the pass must stop whatever its budget (see mustAbort).
func (s *cliqueSearch) mustStop() bool {
	return s.work > s.budget || s.mustAbort()
}

// mustAbort tells whether the pass has to be interrupted whatever its budget:
// the context is done, or it is a focused pass and its target has been proved
// out of reach.
func (s *cliqueSearch) mustAbort() bool {
	return s.ctx.Err() != nil || (!s.improving && s.unreachable.Load())
}

// colorSort splits candidates into classes of pairwise non-adjacent vertices
// (greedy sequential coloring) and returns them class after class, with the
// color of each vertex. Any clique takes at most one vertex of each class.
//
// The coloring can be long: it stops, and sets s.aborted, as soon as the search
// has to be aborted. Its cost is not charged to the work budget of the pass,
// which limits the exploration of the tree, not the bound computed at each node.
func (s *cliqueSearch) colorSort(depth int, candidates []int32) (sorted, colors []int32) {
	k := 0
	for _, v := range candidates {
		if s.mustAbort() {
			s.aborted = true
			return nil, nil
		}
		placed := false
		for c := 0; c < k; c++ {
			// v joins the class if it is adjacent to none of its members
			s.probe = s.filter(v, s.classes[c], s.probe[:0])
			if len(s.probe) == 0 {
				s.classes[c] = append(s.classes[c], v)
				placed = true
				break
			}
		}
		if !placed {
			if k == len(s.classes) {
				s.classes = append(s.classes, nil)
			}
			s.classes[k] = append(s.classes[k][:0], v)
			k++
		}
	}

	for len(s.sorted) <= depth {
		s.sorted = append(s.sorted, nil)
		s.colors = append(s.colors, nil)
	}
	sorted, colors = s.sorted[depth][:0], s.colors[depth][:0]
	for c := 0; c < k; c++ {
		for _, v := range s.classes[c] {
			sorted = append(sorted, v)
			colors = append(colors, int32(c+1))
		}
	}
	s.sorted[depth], s.colors[depth] = sorted, colors
	return sorted, colors
}

// search extends the current clique of depth vertices with vertices taken in
// candidates, all adjacent to the vertices of the current clique. The list is
// read from its end: the vertex at position i is combined with the vertices
// before it, so that every clique is met once.
func (s *cliqueSearch) search(depth int, candidates []int32) {
	var colors []int32
	if s.colorLimit > 0 && len(candidates) > 1 && len(candidates) <= s.colorLimit {
		candidates, colors = s.colorSort(depth, candidates)
		if s.aborted {
			return
		}
	}

	for i := len(candidates) - 1; i >= 0; i-- {
		// largest clique the vertices 0..i can make
		bound := i + 1
		if colors != nil {
			bound = int(colors[i])
		}
		if depth+bound < s.threshold() {
			return
		}
		if s.mustStop() {
			s.aborted = true
			return
		}

		v := candidates[i]
		s.current = append(s.current[:depth], v)
		if len(s.current) > int(s.bestSize.Load()) {
			s.publish(s.current)
			if len(s.current) >= s.target {
				s.aborted = true // the search is over: do not report an exhaustion
				return
			}
		}

		rest := candidates[:i]
		if depth+1+len(rest) < s.threshold() {
			continue
		}
		for len(s.buffers) <= depth+1 {
			s.buffers = append(s.buffers, nil)
		}
		next := s.filter(v, rest, s.buffers[depth+1][:0])
		s.buffers[depth+1] = next
		s.work += int64(len(rest))

		if depth+1+len(next) < s.threshold() {
			continue
		}
		s.search(depth+1, next)
		if s.aborted {
			return
		}
	}
}
