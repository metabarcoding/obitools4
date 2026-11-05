package obidnapat

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obialign"
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

// IsMatching tells whether the pattern matches somewhere in the region.
func (p ApatPattern) IsMatching(seq *obiseq.BioSequence, begin, length int) bool {
	return len(p.AppendAllIndex(nil, seq, begin, length)) > 0
}

// IsMatchingCircular is IsMatching on a circular sequence.
func (p ApatPattern) IsMatchingCircular(seq *obiseq.BioSequence, begin, length int) bool {
	return len(p.AppendAllIndexCircular(nil, seq, begin, length)) > 0
}

// BestMatch returns the first match with the fewest errors.
//
// For a pattern compiled with allowsIndel, a match with errors is only a nominal
// window (see scan_indel.go): it is realigned, as obiapat does, by
// obialign.LocatePattern on the part of the sequence around it, which gives its
// real start, end and number of errors. When that alignment cannot be
// reconstructed (the fragment is too short for the pattern), there is no match.
func (p ApatPattern) BestMatch(seq *obiseq.BioSequence, begin, length int) (start int, end int, nerr int, matched bool) {
	best := -1
	res := p.AppendAllIndex(nil, seq, begin, length)
	for i, m := range res {
		if best < 0 || m[2] < res[best][2] {
			best = i
		}
	}
	if best < 0 {
		return 0, 0, 0, false
	}
	start, end, nerr = res[best][0], res[best][1], res[best][2]
	if nerr == 0 || !p.c.allowsIndel {
		return start, end, nerr, true
	}

	patlen := len(p.c.positions)
	n := seq.Len()
	fragStart := max(start-nerr, 0)
	fragEnd := min(start+patlen+nerr, n)
	pattern := p.c.raw[:min(patlen, len(p.c.raw))]
	from, to, score := obialign.LocatePattern(seq.Id(), []byte(pattern), seq.Sequence()[fragStart:fragEnd])
	if from < 0 {
		return 0, 0, 0, false
	}
	return fragStart + from, fragStart + to, score, true
}

// FilterBestMatch keeps, of the matches that overlap, the one with the fewest
// errors (the first one when they tie).
func (p ApatPattern) FilterBestMatch(seq *obiseq.BioSequence, begin, length int) [][3]int {
	const none = 10000
	res := p.AppendAllIndex(nil, seq, begin, length)
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
	return nilIfEmpty(filtered)
}

// AllMatches returns the matches that do not overlap a better one.
func (p ApatPattern) AllMatches(seq *obiseq.BioSequence, begin, length int) [][3]int {
	return p.FilterBestMatch(seq, begin, length)
}
