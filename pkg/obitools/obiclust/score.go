package obiclust

import (
	"math"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obialign"
	log "github.com/sirupsen/logrus"
)

// entry is a distinct sequence of the dataset with what the clustering needs.
type entry struct {
	seq    []byte
	length int
	count  int // cumulated count of the identical sequences
	table  *table4
	sparse []uint16 // words of the table (see sparseWords), nil if unusable
	// ambiguous sequences contain IUPAC letters other than a, c, g and t:
	// they have no 4-mer table, and cannot be the center of a cluster
	ambiguous bool
}

// scoring holds the way sumaclust scores a pair of sequences and decides
// whether it passes the threshold. It reproduces the arithmetic of the
// reference implementation (integer truncations, floating point
// comparisons) so that the same pairs are accepted with the same scores.
type scoring struct {
	threshold float64 // threshold as used for the comparison (1 - t for a normalized distance)
	normalize bool
	lcsmode   bool // true: score is a similarity, false: it is a distance
	ref       reference
	ratio     float64
	worst     float64 // score of a pair that cannot be attached

	// byLength gives, for the pairs whose longest sequence has a length l,
	// the values that depend on l only. It is built when the reference is not
	// the shortest sequence, otherwise those values depend on both lengths.
	byLength []lengthInfo
}

type lengthInfo struct {
	lcsMin    int32 // see lcsMin
	minCommon int32 // see minCommon
}

// prepareLengths fills byLength for the lengths up to lmax.
func (sc *scoring) prepareLengths(lmax int) {
	if sc.ref == refShortest {
		return
	}
	sc.byLength = make([]lengthInfo, lmax+1)
	for l := range sc.byLength {
		sc.byLength[l] = lengthInfo{int32(sc.lcsMin(l, l)), int32(sc.minCommon(l, l))}
	}
}

// similarityLike tells if the comparison with the threshold is "score >= threshold".
func (sc *scoring) similarityLike() bool { return sc.lcsmode || sc.normalize }

// lcsMin is the smallest LCS score that a pair of lengths l1 >= l2 can have
// and still pass the threshold (calculateLCSmin of sumaclust).
func (sc *scoring) lcsMin(l1, l2 int) int {
	switch {
	case sc.threshold <= 0:
		return 0
	case sc.normalize:
		if sc.ref == refShortest {
			return int(sc.threshold * float64(l2))
		}
		return int(sc.threshold * float64(l1))
	case sc.lcsmode:
		return int(sc.threshold)
	case sc.ref == refShortest:
		return int(float64(l2) - sc.threshold)
	default:
		return int(float64(l1) - sc.threshold)
	}
}

// minCommon is the number of 4-mers that two sequences of lengths l1 >= l2
// must share to be possibly similar enough (isPossibleSumathings).
func (sc *scoring) minCommon(l1, l2 int) int {
	reflen := l2
	if sc.ref == refAlignment || sc.ref == refLongest {
		reflen = l1
	}
	var lcs int
	if sc.normalize {
		lcs = int(math.Ceil(float64(reflen) * sc.threshold))
	} else {
		th := sc.threshold
		if !sc.lcsmode {
			th = float64(reflen) - th
		}
		lcs = int(th)
	}
	return minCommonWords(l1, lcs)
}

// pairScore returns the score of a query sequence against a center, or the
// worst score when the pair cannot be attached.
//
// common, if not negative, is the sum of the minima of the 4-mer counters of
// the two sequences, already computed by a block of tables.
func (sc *scoring) pairScore(q, c *entry, common int, buf *obialign.BandedLCSBuffer) float64 {
	// The quotient of two counts is not above 1 when the first is not above
	// the second, which spares the division in the usual case.
	if !(sc.ratio >= 1 && q.count <= c.count) && float64(q.count)/float64(c.count) > sc.ratio {
		return sc.worst
	}

	l1, l2 := q.length, c.length
	long, short := q, c
	if l1 < l2 {
		l1, l2 = l2, l1
		long, short = c, q
	}
	var lcsMin, minCommon int
	if sc.byLength != nil {
		info := sc.byLength[l1]
		lcsMin, minCommon = int(info.lcsMin), int(info.minCommon)
	} else {
		lcsMin, minCommon = sc.lcsMin(l1, l2), sc.minCommon(l1, l2)
	}
	if l2 < lcsMin {
		return sc.worst
	}
	// The 4-mer filter counts identical words: it does not hold when IUPAC
	// codes can match different letters. It lets short sequences pass.
	if l1 >= 12 && l2 >= 12 && !long.ambiguous && !short.ambiguous {
		if common >= 0 {
			common += min(long.table.overflow, short.table.overflow)
		} else {
			common = commonWords(long.table, short.table)
		}
		if common < minCommon {
			return sc.worst
		}
	}

	// The length of the alignment is needed only to normalize by it, or to
	// express a distance with it; the score alone is much cheaper.
	var lcs, alilen int
	var err error
	if sc.ref == refAlignment && (sc.normalize || !sc.lcsmode) {
		lcs, alilen, err = buf.AtLeast(q.seq, c.seq, lcsMin)
	} else {
		lcs, err = buf.ScoreAtLeast(q.seq, c.seq, lcsMin)
	}
	if err != nil {
		log.Fatalf("Cannot align two sequences: %v", err)
	}
	if lcs < 0 {
		// The best alignment has a score under lcsMin: the pair cannot pass.
		return sc.worst
	}

	id := float64(lcs)
	if !sc.lcsmode && !sc.normalize {
		switch sc.ref {
		case refAlignment:
			id = float64(alilen) - id
		case refLongest:
			id = float64(l1) - id
		case refShortest:
			id = float64(l2) - id
		}
	}
	if sc.normalize {
		switch sc.ref {
		case refAlignment:
			id /= float64(alilen)
		case refLongest:
			id /= float64(l1)
		case refShortest:
			id /= float64(l2)
		}
	}
	return id
}

// beats tells if the score s is better than the score best (without any
// threshold).
func (sc *scoring) beats(s, best float64) bool {
	if sc.similarityLike() {
		return s > best
	}
	return s < best
}

// lcsMinFor is lcsMin for sequences of lengths l1 >= l2, from the table when it exists.
func (sc *scoring) lcsMinFor(l1, l2 int) int {
	if sc.byLength != nil {
		return int(sc.byLength[l1].lcsMin)
	}
	return sc.lcsMin(l1, l2)
}

// lengthCompatible tells if two sequences of these lengths may be similar enough:
// the shortest cannot be shorter than the minimum LCS score of the pair.
func (sc *scoring) lengthCompatible(la, lb int) bool {
	l1, l2 := max(la, lb), min(la, lb)
	return l2 >= sc.lcsMinFor(l1, l2)
}

// selects tells if a pair of score s must replace the best pair found so far,
// of score best.
func (sc *scoring) selects(s, best float64) bool {
	if sc.similarityLike() {
		return s >= sc.threshold && s > best
	}
	return s <= sc.threshold && s < best
}

// centerScore is the score of a sequence with itself.
func (sc *scoring) centerScore(length int) float64 {
	switch {
	case sc.normalize && sc.lcsmode:
		return 1.0
	case !sc.lcsmode:
		return 0.0
	default:
		return float64(length)
	}
}

// memberScore is the score reported for a sequence attached to a center
// with the score s.
func (sc *scoring) memberScore(s float64) float64 {
	if !sc.lcsmode && sc.normalize {
		return 1.0 - s
	}
	return s
}
