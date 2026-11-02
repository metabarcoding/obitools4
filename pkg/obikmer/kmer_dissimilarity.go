package obikmer

import (
	"fmt"
	"strings"
	"sync"
)

// KmerMetric selects the distance used to compare two k-mers.
type KmerMetric int

const (
	// HammingMetric counts the positions at which two k-mers differ.
	HammingMetric KmerMetric = iota
	// LCSMetric is the alignment length minus the LCS length (see KmerLCS).
	LCSMetric
)

// ParseKmerMetric converts "hamming" or "lcs" (case insensitive) to a KmerMetric.
func ParseKmerMetric(name string) (KmerMetric, error) {
	switch strings.ToLower(name) {
	case "hamming":
		return HammingMetric, nil
	case "lcs":
		return LCSMetric, nil
	}
	return HammingMetric, fmt.Errorf("unknown distance type %q (expected hamming or lcs)", name)
}

func (m KmerMetric) String() string {
	if m == LCSMetric {
		return "lcs"
	}
	return "hamming"
}

type lcsScratch struct {
	partners []uint64
	results  []KmerLCSResult
}

// NewKmerDissimilarityFilter describes the graph whose vertices are kmers[i]
// (k-mers of size k, 1 <= k <= 31) and where two vertices are linked when
// their distance is at least dmin. The returned function has the signature of
// obigraph.NeighborFilter: it appends to out[:0] the candidates (indices in
// kmers) linked to v. It is safe for concurrent use.
func NewKmerDissimilarityFilter(kmers []uint64, k int, metric KmerMetric, dmin int) func(v int32, candidates, out []int32) []int32 {
	if metric == HammingMetric {
		return func(v int32, candidates, out []int32) []int32 {
			q := kmers[v]
			out = out[:0]
			for _, c := range candidates {
				if HammingDistance(q, kmers[c], k) >= dmin {
					out = append(out, c)
				}
			}
			return out
		}
	}

	pool := sync.Pool{New: func() any { return new(lcsScratch) }}
	return func(v int32, candidates, out []int32) []int32 {
		s := pool.Get().(*lcsScratch)
		defer pool.Put(s)

		s.partners = s.partners[:0]
		for _, c := range candidates {
			s.partners = append(s.partners, kmers[c])
		}
		if cap(s.results) < len(candidates) {
			s.results = make([]KmerLCSResult, len(candidates))
		}
		s.results = s.results[:len(candidates)]
		KmerLCSBatch(kmers[v], s.partners, k, s.results)

		out = out[:0]
		for i, c := range candidates {
			if s.results[i].Distance() >= dmin {
				out = append(out, c)
			}
		}
		return out
	}
}
