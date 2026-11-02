package obisimilarity

import (
	"errors"
	"math"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obialign"
	log "github.com/sirupsen/logrus"
)

// Reference is the length used to normalize a score (or to express a raw
// distance): the alignment length, the longest or the shortest sequence.
type Reference int

const (
	RefAlignment Reference = iota
	RefLongest
	RefShortest
)

// Params describe how a pair of sequences is scored and compared to a threshold.
type Params struct {
	// Threshold: with a normalized similarity an identity (0.95 for 95%), with
	// a normalized distance 1 - identity, with a raw similarity the length of
	// the LCS, with a raw distance the number of differences with the
	// reference length. 0 means no threshold: all the pairs are reported.
	Threshold  float64
	Normalized bool // the score is divided by the reference length
	Distance   bool // the score is a distance, not a similarity
	Reference  Reference
	// Ratio is the maximum ratio between the counts of a sequence and of the
	// one it is compared to (clustering only): a larger ratio gives no match.
	Ratio float64
}

// Scoring scores the pairs of sequences and decides whether they pass the
// threshold. It must not be modified once built.
type Scoring struct {
	threshold float64 // threshold as used for the comparison (1 - t for a normalized distance)
	normalize bool
	lcsmode   bool // true: score is a similarity, false: it is a distance
	ref       Reference
	ratio     float64
	worst     float64 // score of a pair that does not pass
	all       bool    // no threshold: every pair is reported

	// byLength gives, for the pairs whose longest sequence has a length l,
	// the values that depend on l only. It is built when the reference is not
	// the shortest sequence, otherwise those values depend on both lengths.
	byLength []lengthInfo
}

type lengthInfo struct {
	lcsMin    int32 // see lcsMin
	minCommon int32 // see minCommon
}

// NewScoring builds the scoring of the sequences `entries`, which are the
// ones to be compared (the longest of them is needed for a raw distance).
func NewScoring(p Params, entries []Entry) (*Scoring, error) {
	if p.Threshold < 0 || (p.Normalized && p.Threshold > 1.0) {
		return nil, errors.New("the threshold must not be negative, and must be at most 1 when scores are normalized")
	}

	sc := &Scoring{
		normalize: p.Normalized,
		lcsmode:   !p.Distance,
		ref:       p.Reference,
		ratio:     p.Ratio,
	}
	sc.threshold = p.Threshold
	sc.all = p.Threshold == 0
	if !sc.lcsmode && sc.normalize && p.Threshold > 0 {
		sc.threshold = 1.0 - p.Threshold
	}

	lmax := 0
	for i := range entries {
		lmax = max(lmax, entries[i].Length)
	}
	if !sc.similarityLike() {
		sc.worst = float64(lmax)
	}

	if sc.ref != RefShortest {
		sc.byLength = make([]lengthInfo, lmax+1)
		for l := range sc.byLength {
			sc.byLength[l] = lengthInfo{int32(sc.lcsMin(l, l)), int32(sc.minCommon(l, l))}
		}
	}
	return sc, nil
}

// Worst is the score of a pair that does not pass the threshold.
func (sc *Scoring) Worst() float64 { return sc.worst }

// Threshold is the threshold the scores are compared to.
func (sc *Scoring) Threshold() float64 { return sc.threshold }

// SimilarityLike tells if the comparison with the threshold is "score >= threshold".
func (sc *Scoring) SimilarityLike() bool { return sc.similarityLike() }

func (sc *Scoring) similarityLike() bool { return sc.lcsmode || sc.normalize }

// lcsMin is the smallest LCS score that a pair of lengths l1 >= l2 can have
// and still pass the threshold.
func (sc *Scoring) lcsMin(l1, l2 int) int {
	switch {
	case sc.threshold <= 0:
		return 0
	case sc.normalize:
		if sc.ref == RefShortest {
			return int(sc.threshold * float64(l2))
		}
		return int(sc.threshold * float64(l1))
	case sc.lcsmode:
		return int(sc.threshold)
	case sc.ref == RefShortest:
		return int(float64(l2) - sc.threshold)
	default:
		return int(float64(l1) - sc.threshold)
	}
}

// minCommon is the number of 4-mers that two sequences of lengths l1 >= l2
// must share to be possibly similar enough.
func (sc *Scoring) minCommon(l1, l2 int) int {
	if sc.all {
		return 0 // no threshold, no filter
	}
	reflen := l2
	if sc.ref == RefAlignment || sc.ref == RefLongest {
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

// lcsMinFor is lcsMin for sequences of lengths l1 >= l2, from the table when it exists.
func (sc *Scoring) lcsMinFor(l1, l2 int) int {
	if sc.byLength != nil {
		return int(sc.byLength[l1].lcsMin)
	}
	return sc.lcsMin(l1, l2)
}

// LengthCompatible tells if two sequences of these lengths may be similar
// enough: the shortest cannot be shorter than the minimum LCS score of the pair.
func (sc *Scoring) LengthCompatible(la, lb int) bool {
	l1, l2 := max(la, lb), min(la, lb)
	return l2 >= sc.lcsMinFor(l1, l2)
}

// PairScore scores the pair (q, c). ok is false when the pair is known not to
// reach the threshold, in which case no score is given.
//
// common, if not negative, is the sum of the minima of the 4-mer counters of
// the two sequences, already computed by a block of tables (see Index).
func (sc *Scoring) PairScore(q, c *Entry, common int, buf *obialign.BandedLCSBuffer) (score float64, ok bool) {
	// The quotient of two counts is not above 1 when the first is not above
	// the second, which spares the division in the usual case.
	if !(sc.ratio >= 1 && q.Count <= c.Count) && float64(q.Count)/float64(c.Count) > sc.ratio {
		return 0, false
	}

	l1, l2 := q.Length, c.Length
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
		return 0, false
	}
	// The 4-mer filter counts identical words: it does not hold when IUPAC
	// codes can match different letters. It lets short sequences pass.
	if l1 >= 12 && l2 >= 12 && !long.Ambiguous && !short.Ambiguous {
		if common >= 0 {
			common += min(long.Table.Overflow, short.Table.Overflow)
		} else {
			common = commonWords(long.Table, short.Table)
		}
		if common < minCommon {
			return 0, false
		}
	}

	// The length of the alignment is needed only to normalize by it, or to
	// express a distance with it; the score alone is much cheaper.
	var lcs, alilen int
	var err error
	if sc.ref == RefAlignment && (sc.normalize || !sc.lcsmode) {
		lcs, alilen, err = buf.AtLeast(q.Seq, c.Seq, lcsMin)
	} else {
		lcs, err = buf.ScoreAtLeast(q.Seq, c.Seq, lcsMin)
	}
	if err != nil {
		log.Fatalf("Cannot align two sequences: %v", err)
	}
	if lcs < 0 {
		// The best alignment has a score under lcsMin: the pair cannot pass.
		return 0, false
	}

	id := float64(lcs)
	if !sc.lcsmode && !sc.normalize {
		switch sc.ref {
		case RefAlignment:
			id = float64(alilen) - id
		case RefLongest:
			id = float64(l1) - id
		case RefShortest:
			id = float64(l2) - id
		}
	}
	if sc.normalize {
		switch sc.ref {
		case RefAlignment:
			id /= float64(alilen)
		case RefLongest:
			id /= float64(l1)
		case RefShortest:
			id /= float64(l2)
		}
	}
	return id, true
}

// Passes tells if the score s of a pair reaches the threshold.
func (sc *Scoring) Passes(s float64) bool {
	if sc.all {
		return true
	}
	if sc.similarityLike() {
		return s >= sc.threshold
	}
	return s <= sc.threshold
}

// Beats tells if the score s is better than the score best (without any
// threshold).
func (sc *Scoring) Beats(s, best float64) bool {
	if sc.similarityLike() {
		return s > best
	}
	return s < best
}

// SelfScore is the score of a sequence with itself.
func (sc *Scoring) SelfScore(length int) float64 {
	switch {
	case sc.normalize && sc.lcsmode:
		return 1.0
	case !sc.lcsmode:
		return 0.0
	default:
		return float64(length)
	}
}

// Reported is the score to report for a pair whose score is s: a normalized
// distance is computed as a similarity against the converted threshold.
func (sc *Scoring) Reported(s float64) float64 {
	if !sc.lcsmode && sc.normalize {
		return 1.0 - s
	}
	return s
}
