package obikmer

import "math/bits"

// KmerLCSResult is the result of a k-mer LCS alignment: the length of the
// Longest Common Subsequence and the length of the shortest alignment
// achieving it (same semantics as obialign.FastLCSScore).
type KmerLCSResult struct {
	LCS      int
	AlignLen int
}

// Distance returns AlignLen - LCS, the LCS-mode edit distance.
func (r KmerLCSResult) Distance() int {
	return r.AlignLen - r.LCS
}

// HammingDistance returns the number of positions at which the two k-mers
// differ. Both k-mers are encoded with 2 bits per base (see EncodeKmer);
// bits above the 2*k lowest ones are ignored. k must be between 0 and 32.
func HammingDistance(a, b uint64, k int) int {
	diff := a ^ b
	diff = (diff | diff>>1) & 0x5555555555555555
	diff &= ^uint64(0) >> (64 - 2*k)
	return bits.OnesCount64(diff)
}

// KmerLCS aligns two k-mers of the same size in LCS mode (match +1, mismatch
// and indel 0) and returns the LCS length together with the length of the
// shortest alignment among those reaching it, so that the distance is
// AlignLen - LCS.
//
// It is the scalar reference implementation, straightforward on purpose:
// it is the oracle against which the vectorised versions are tested.
// It returns {-1, -1} if k is not between 1 and 31.
func KmerLCS(a, b uint64, k int) KmerLCSResult {
	if k < 1 || k > 31 {
		return KmerLCSResult{-1, -1}
	}

	var ba, bb [32]uint8
	for i := k - 1; i >= 0; i-- {
		ba[i] = uint8(a & 3)
		bb[i] = uint8(b & 3)
		a >>= 2
		b >>= 2
	}

	// One DP row: best (score, alignment length) per column.
	var score, length [32]int
	for j := 0; j <= k; j++ {
		length[j] = j
	}

	better := func(s1, l1, s2, l2 int) bool {
		return s1 > s2 || (s1 == s2 && l1 < l2)
	}

	for i := 1; i <= k; i++ {
		diagS, diagL := score[0], length[0]
		score[0], length[0] = 0, i

		for j := 1; j <= k; j++ {
			upS, upL := score[j], length[j]

			bestS, bestL := diagS, diagL+1
			if ba[j-1] == bb[i-1] {
				bestS++
			}
			if better(upS, upL+1, bestS, bestL) {
				bestS, bestL = upS, upL+1
			}
			if better(score[j-1], length[j-1]+1, bestS, bestL) {
				bestS, bestL = score[j-1], length[j-1]+1
			}

			diagS, diagL = upS, upL
			score[j], length[j] = bestS, bestL
		}
	}

	return KmerLCSResult{score[k], length[k]}
}

// KmerLCSBatch computes KmerLCS(query, partners[i], k) for every partner and
// stores it in out[i]. out must be at least as long as partners.
//
// When built with GOEXPERIMENT=simd the partners are aligned several at a
// time, one per SIMD lane; otherwise the scalar reference is used.
// It returns false, leaving out untouched, if k is not between 1 and 31 or
// if out is too short.
func KmerLCSBatch(query uint64, partners []uint64, k int, out []KmerLCSResult) bool {
	if k < 1 || k > 31 || len(out) < len(partners) {
		return false
	}
	kmerLCSBatch(query, partners, k, out)
	return true
}

// KmerLCSPartners is a fixed set of k-mers prepared once to be aligned against
// many queries with KmerLCSBatch-like speed, without redoing the per-call
// reorganisation of the partners. It is read-only after creation and can be
// shared between goroutines.
type KmerLCSPartners struct {
	prepared kmerLCSPrepared
	n        int
}

// NewKmerLCSPartners prepares partners for repeated alignment.
// It returns nil if k is not between 1 and 31. The slice is not copied when
// built without GOEXPERIMENT=simd, so it must not be modified afterwards.
func NewKmerLCSPartners(partners []uint64, k int) *KmerLCSPartners {
	if k < 1 || k > 31 {
		return nil
	}
	return &KmerLCSPartners{prepared: prepareKmerLCS(partners, k), n: len(partners)}
}

// Len returns the number of prepared partners.
func (p *KmerLCSPartners) Len() int { return p.n }

// Align computes the alignment of query against every prepared partner and
// stores it in out. It returns false if out is shorter than Len().
func (p *KmerLCSPartners) Align(query uint64, out []KmerLCSResult) bool {
	if len(out) < p.n {
		return false
	}
	p.prepared.align(query, out[:p.n])
	return true
}

func kmerLCSBatchScalar(query uint64, partners []uint64, k int, out []KmerLCSResult) {
	for i, p := range partners {
		out[i] = KmerLCS(query, p, k)
	}
}
