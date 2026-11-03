//go:build goexperiment.simd

package obikmer

import "simd"

// Each lane of a vector holds the state of one (query, partner) alignment as
//
//	score<<6 | (63 - alignment length)
//
// so that taking the maximum of two cells ranks them by score first and, for
// equal scores, by shortest alignment. With k <= 31 the score is <= 31 and the
// alignment length <= 62, hence the value fits in 11 bits and the low field
// is never exhausted. The three DP moves then all reduce to
//
//	cell(i,j) = max(diag + 64*match, up, left) - 1
//
// (every move lengthens the alignment by one; only a diagonal match adds 1
// to the score).
const (
	lcsScoreUnit = 64
	lcsLenBase   = 63
)

func simdLanes() int {
	var probe simd.Uint16s
	return probe.Len()
}

// transposeChunk writes the bases of up to `lanes` partners in dst, laid out
// so that dst[j*lanes+lane] is base j of the partner in `lane`.
func transposeChunk(chunk []uint64, k, lanes int, dst []uint16) {
	for lane, p := range chunk {
		for j := 0; j < k; j++ {
			dst[j*lanes+lane] = uint16(p>>(2*uint(k-1-j))) & 3
		}
	}
}

// alignChunk aligns the query against the partners of one transposed chunk
// and writes one result per lane in out (len(out) <= lanes).
func alignChunk(query uint64, pbase, row []uint16, k, lanes int, out []KmerLCSResult) {
	one := simd.BroadcastUint16s(1)
	bonus := simd.BroadcastUint16s(lcsScoreUnit)

	for j := 0; j <= k; j++ {
		simd.BroadcastUint16s(uint16(lcsLenBase - j)).Store(row[j*lanes:])
	}

	for i := 1; i <= k; i++ {
		qv := simd.BroadcastUint16s(uint16(query>>(2*uint(k-i))) & 3)
		diag := simd.LoadUint16s(row)
		left := simd.BroadcastUint16s(uint16(lcsLenBase - i))
		left.Store(row)

		for j := 1; j <= k; j++ {
			up := simd.LoadUint16s(row[j*lanes:])
			pv := simd.LoadUint16s(pbase[(j-1)*lanes:])
			match := qv.Equal(pv)
			best := diag.Add(bonus.Masked(match)).Max(up).Max(left).Sub(one)
			best.Store(row[j*lanes:])
			diag = up
			left = best
		}
	}

	last := row[k*lanes:]
	for lane := range out {
		v := int(last[lane])
		out[lane] = KmerLCSResult{
			LCS:      v / lcsScoreUnit,
			AlignLen: lcsLenBase - v%lcsScoreUnit,
		}
	}
}

func kmerLCSBatch(query uint64, partners []uint64, k int, out []KmerLCSResult) {
	lanes := simdLanes()
	pbase := make([]uint16, k*lanes)
	row := make([]uint16, (k+1)*lanes)

	for start := 0; start < len(partners); start += lanes {
		end := min(start+lanes, len(partners))
		transposeChunk(partners[start:end], k, lanes, pbase)
		alignChunk(query, pbase, row, k, lanes, out[start:end])
	}
}

// kmerLCSPrepared holds partners already transposed, chunk after chunk.
type kmerLCSPrepared struct {
	k, lanes, n int
	pbase       []uint16 // chunk c at [c*k*lanes : (c+1)*k*lanes]
}

func prepareKmerLCS(partners []uint64, k int) kmerLCSPrepared {
	lanes := simdLanes()
	chunks := (len(partners) + lanes - 1) / lanes
	p := kmerLCSPrepared{k: k, lanes: lanes, n: len(partners),
		pbase: make([]uint16, chunks*k*lanes)}
	for c := 0; c < chunks; c++ {
		end := min((c+1)*lanes, len(partners))
		transposeChunk(partners[c*lanes:end], k, lanes, p.pbase[c*k*lanes:(c+1)*k*lanes])
	}
	return p
}

func (p *kmerLCSPrepared) align(query uint64, out []KmerLCSResult) {
	row := make([]uint16, (p.k+1)*p.lanes)
	size := p.k * p.lanes
	for c, start := 0, 0; start < p.n; c, start = c+1, start+p.lanes {
		end := min(start+p.lanes, p.n)
		alignChunk(query, p.pbase[c*size:(c+1)*size], row, p.k, p.lanes, out[start:end])
	}
}
