package obialign

import (
	"errors"
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// MaxBandedLCSLength is the longest sequence BandedLCS accepts. The score and
// the alignment length are packed on 16 bits each, and the alignment length
// can reach the sum of both sequence lengths.
const MaxBandedLCSLength = 1<<15 - 1

// ErrSequenceTooLong is returned by BandedLCS when a sequence is longer than
// MaxBandedLCSLength.
var ErrSequenceTooLong = errors.New("obialign: sequence too long for BandedLCS")

// A DP cell is packed on 32 bits as
//
//	score<<16 | (lcsLenBase - alignment length)
//
// so that the maximum of two cells ranks them by score first and, for equal
// scores, by shortest alignment. The three moves then reduce to
//
//	cell(i,j) = max(diag + match<<16, up, left) - 1
//
// since every move lengthens the alignment by one and only a diagonal match
// adds one to the score. The value 0 stands for an unavailable cell (outside
// the band): it loses against any real cell. Every cell of the band has at
// least one real neighbour (the band is connected and contains (0,0)), so
// the subtraction never wraps for a cell that is used; the scalar reference
// still clamps the maximum to 1 as a safety net.
const (
	lcsScoreUnit = uint32(1) << 16
	lcsLenBase   = uint32(0xFFFF)
)

// bandedLCSWorkspace holds the scratch memory of one alignment.
type bandedLCSWorkspace struct {
	aRev, bCode []uint32    // encoded bases (aRev is reversed), SIMD kernel only
	rows        [3][]uint32 // antidiagonal (SIMD) or row (scalar) buffers
	widened     int         // number of alignments that needed a second pass
}

var bandedLCSPool = sync.Pool{New: func() any { return new(bandedLCSWorkspace) }}

func growUint32(s []uint32, n int) []uint32 {
	if cap(s) < n {
		return make([]uint32, n)
	}
	return s[:n]
}

// iupacMask gives for each byte its 4-bit nucleotide set (A=1, C=2, G=4,
// T=8), or 0 when the byte is not an IUPAC letter. For two bytes of non-zero
// mask, SameIUPACNuc(a, b) is exactly (mask[a] & mask[b]) != 0.
var iupacMask = buildIUPACMask()

func buildIUPACMask() (m [256]uint8) {
	probes := [4]byte{'a', 'c', 'g', 't'}
	for c := 0; c < 256; c++ {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z') {
			continue
		}
		for bit, p := range probes {
			if obiseq.SameIUPACNuc(byte(c), p) {
				m[c] |= 1 << uint(bit)
			}
		}
	}
	return m
}

// allRegularIUPAC reports whether every base has a non-zero IUPAC mask, in
// which case base compatibility is a plain AND of masks.
func allRegularIUPAC(s []byte) bool {
	for _, c := range s {
		if iupacMask[c] == 0 {
			return false
		}
	}
	return true
}

// bandOffsets returns the diagonal offsets k = j - i covered by a band of
// half-width g around the diagonals 0..delta, with sequence A (columns j) not
// shorter than sequence B (rows i), clipped to the matrix.
func bandOffsets(lA, lB, g int) (kmin, kmax int) {
	return max(-g, -lB), min(lA-lB+g, lA)
}

// BandedLCS computes the score of the longest common subsequence of two
// nucleotide sequences, and the length of the shortest alignment achieving
// it (match: +1, mismatch and gap: 0; bases are compared with the IUPAC
// rules of obiseq.SameIUPACNuc). The distance is alilen - lcs.
//
// If maxError >= 0, the result is the optimal (lcs, alilen) when its distance
// is at most maxError, and (-1, -1) otherwise. This is exact, not a
// heuristic: the DP is restricted to a band, and a bound on what a path
// leaving the band can score proves when the band was wide enough. maxError < 0
// means no limit.
//
// An error is returned when a sequence is longer than MaxBandedLCSLength.
func BandedLCS(seqA, seqB []byte, maxError int) (lcs, alilen int, err error) {
	if len(seqA) > MaxBandedLCSLength || len(seqB) > MaxBandedLCSLength {
		return -1, -1, ErrSequenceTooLong
	}
	if len(seqA) < len(seqB) {
		seqA, seqB = seqB, seqA
	}
	ws := bandedLCSPool.Get().(*bandedLCSWorkspace)
	lcs, alilen = bandedLCS(seqA, seqB, maxError, ws, haveSIMD)
	bandedLCSPool.Put(ws)
	return lcs, alilen, nil
}

// bandedLCS expects len(a) >= len(b), both within MaxBandedLCSLength.
//
// With delta = lA - lB and a band of half-width g (offsets -g .. delta+g):
//   - an alignment with at most E errors needs at least delta+2h gaps to reach
//     the offset delta+h or -h, so it lies in the band of g0 = (E-delta)/2;
//   - a path leaving a band of half-width g makes at least g+1 gaps on the B
//     side, hence scores at most lB-g-1 matches, and has more than E errors
//     as soon as g >= g0.
//
// Let Q be the best path of the band g0. If its distance exceeds E, the
// global optimum cannot be within E either (it would be in the band and be
// Q): the answer is -1. Otherwise, if score(Q) >= lB-g0-1 no outside path can
// beat Q (a tie in score loses on length, the outside path having more errors
// than Q): Q is the global optimum. Else the band is widened once to
// g1 = min(E-delta, lB-score(Q)-1); since score(Q) >= lA-E for any alignment
// within E errors, g1 always certifies the result.
func bandedLCS(a, b []byte, maxError int, ws *bandedLCSWorkspace, useSIMD bool) (int, int) {
	lA, lB := len(a), len(b)
	delta := lA - lB
	if maxError >= 0 && maxError < delta {
		return -1, -1
	}
	if lB == 0 {
		return 0, lA
	}
	simdOK := useSIMD && allRegularIUPAC(a) && allRegularIUPAC(b)
	if simdOK {
		encodeBases(ws, a, b)
	}
	run := func(g int) (int, int) {
		kmin, kmax := bandOffsets(lA, lB, g)
		var key uint32
		if simdOK {
			key = bandedLCSSIMD(lA, lB, kmin, kmax, ws)
		} else {
			key = bandedLCSScalar(a, b, kmin, kmax, ws)
		}
		return int(key >> 16), int(lcsLenBase - key&0xFFFF)
	}

	if maxError < 0 {
		return run(lA)
	}
	g0 := (maxError - delta) / 2
	lcs, alilen := run(g0)
	if alilen-lcs > maxError {
		return -1, -1
	}
	if lcs >= lB-g0-1 {
		return lcs, alilen
	}
	ws.widened++
	lcs, alilen = run(min(maxError-delta, lB-lcs-1))
	if alilen-lcs > maxError {
		return -1, -1
	}
	return lcs, alilen
}
