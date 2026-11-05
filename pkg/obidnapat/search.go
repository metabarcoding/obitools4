package obidnapat

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// The matches are [3]int{start, end, errors}: end is not included in the match.
// begin and length delimit the region of the search (length < 0: up to the end
// of the sequence). The *Circular methods read the sequence as circular: a
// match may cross the junction, and then its end goes past the length of the
// sequence. The sequence is read in place and must not be modified during the call.

// AppendAllIndex appends to dst the matches of the pattern in seq, by increasing
// start, one per start with its smallest number of errors.
func (p ApatPattern) AppendAllIndex(dst [][3]int, seq *obiseq.BioSequence, begin, length int) [][3]int {
	return p.c.search(dst, seq.Sequence(), begin, length, false)
}

// AppendAllIndexCircular is AppendAllIndex on a circular sequence.
func (p ApatPattern) AppendAllIndexCircular(dst [][3]int, seq *obiseq.BioSequence, begin, length int) [][3]int {
	return p.c.search(dst, seq.Sequence(), begin, length, true)
}

// FindAllIndex returns the matches of the pattern in seq, nil if there is none.
func (p ApatPattern) FindAllIndex(seq *obiseq.BioSequence, begin, length int) [][3]int {
	return nilIfEmpty(p.AppendAllIndex(nil, seq, begin, length))
}

// FindAllIndexCircular is FindAllIndex on a circular sequence.
func (p ApatPattern) FindAllIndexCircular(seq *obiseq.BioSequence, begin, length int) [][3]int {
	return nilIfEmpty(p.AppendAllIndexCircular(nil, seq, begin, length))
}

func nilIfEmpty(loc [][3]int) [][3]int {
	if len(loc) == 0 {
		return nil
	}
	return loc
}

// IsMatching tells whether the pattern matches somewhere in the region. For a
// pattern with indels any occurrence found counts, even one whose nominal start
// falls before the region (a match with deletions, at the very beginning of the
// sequence): the search starts at the beginning of the region, so the match lies in it.
func (p ApatPattern) IsMatching(seq *obiseq.BioSequence, begin, length int) bool {
	if p.c.indel {
		return len(p.c.indelHits(nil, seq.Sequence(), max(begin, 0), length, false)) > 0
	}
	return len(p.AppendAllIndex(nil, seq, begin, length)) > 0
}

// IsMatchingCircular is IsMatching on a circular sequence.
func (p ApatPattern) IsMatchingCircular(seq *obiseq.BioSequence, begin, length int) bool {
	return len(p.AppendAllIndexCircular(nil, seq, begin, length)) > 0
}

// BestMatch returns the first match with the fewest errors: the first window of
// AllMatches with the fewest errors (see AllMatches for the start of a window
// with indels).
func (p ApatPattern) BestMatch(seq *obiseq.BioSequence, begin, length int) (start int, end int, nerr int, matched bool) {
	res := p.AllMatches(seq, begin, length)
	best := -1
	for i, m := range res {
		if best < 0 || m[2] < res[best][2] {
			best = i
		}
	}
	if best < 0 {
		return 0, 0, 0, false
	}
	return res[best][0], res[best][1], res[best][2], true
}

// FilterBestMatch keeps, of the matches that overlap, the one with the fewest
// errors (the first one when they tie). The windows are the nominal ones of
// AppendAllIndex; see AllMatches for the real start of a window with indels.
func (p ApatPattern) FilterBestMatch(seq *obiseq.BioSequence, begin, length int) [][3]int {
	return nilIfEmpty(bestOfOverlaps(p.AppendAllIndex(nil, seq, begin, length)))
}

// bestOfOverlaps keeps, of the windows that overlap (by their ends and starts, the
// errors included), the one with the fewest errors, the first when they tie. The
// windows come by increasing end.
func bestOfOverlaps(res [][3]int) [][3]int {
	const none = 10000
	filtered := make([][3]int, 0, len(res))
	best := [3]int{0, 0, none}
	for _, m := range res {
		if (m[0] - m[2]) < best[1]+best[2] { // overlapping matches
			if m[2] < best[2] {
				best = m
			}
		} else if best[2] < none {
			filtered = append(filtered, best)
			best = m
		}
	}
	if best[2] < none {
		filtered = append(filtered, best)
	}
	return filtered
}

// AllMatches returns the matches that do not overlap a better one (FilterBestMatch).
//
// For a pattern compiled with allowsIndel the start of a match with errors is its
// real one: the end of a window is the first position, going right, where an
// alignment of that cost ends (an end that ties with the next ones, see
// scan_indel.go), and its start is found by the backward search of
// scan_backward.go (the latest start of an alignment of the whole pattern with
// that many errors, the shortest window), instead of the nominal end - Len(), which
// is off by the number of insertions and deletions. A window whose real start
// cannot be found (only with # in the first positions of the pattern) and the
// matches of a circular sequence keep the nominal start. A match belongs to the
// region when its real start does.
func (p ApatPattern) AllMatches(seq *obiseq.BioSequence, begin, length int) [][3]int {
	if !p.c.indel {
		return p.FilterBestMatch(seq, begin, length)
	}
	begin = max(begin, 0)
	res := bestOfOverlaps(p.c.indelHits(nil, seq.Sequence(), begin, length, false))
	k := 0
	for _, w := range res {
		if w[2] > 0 {
			if s, ok := p.c.backwardStart(seq.Sequence(), w[1]-1, w[2], begin); ok {
				w[0] = s
			}
		}
		if w[0] >= begin {
			res[k] = w
			k++
		}
	}
	return nilIfEmpty(res[:k])
}
