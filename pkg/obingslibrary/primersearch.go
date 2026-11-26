package obingslibrary

import (
	"cmp"
	"fmt"
	"math"
	"slices"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidnapat"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// primerSearch looks for the primers of all the markers of a library at once.
// It is built once by Compile2 and never modified: it can be shared by the
// goroutines that read the sequences.
//
// The markers are numbered from 1 in the order of their pair of primers (the
// iteration order of a map would change the order of the matches that start at
// the same place from one sequence to the next). The pattern 2*(i-1) of the set
// is the forward primer of the marker i, the pattern 2*(i-1)+1 its reverse one;
// the strand -1 of a pattern is its reverse complement.
type primerSearch struct {
	set   *obidnapat.PatternSet
	pairs []PrimerPair // pairs[i]: primers of the marker i (pairs[0] is not used)
}

// compileSearch builds the search of the primers of the compiled markers.
func (library *NGSLibrary) compileSearch() error {
	pairs := make([]PrimerPair, 0, len(library.Markers))
	for pair := range library.Markers {
		pairs = append(pairs, pair)
	}
	slices.SortFunc(pairs, func(a, b PrimerPair) int {
		return cmp.Or(cmp.Compare(a.Forward, b.Forward), cmp.Compare(a.Reverse, b.Reverse))
	})

	patterns := make([]obidnapat.ApatPattern, 0, 2*len(pairs))
	search := &primerSearch{pairs: append([]PrimerPair{{}}, pairs...)}
	for _, pair := range pairs {
		m := library.Markers[pair]
		patterns = append(patterns, m.forward, m.reverse)
	}
	set, err := obidnapat.MakePatternSet(patterns, true)
	if err != nil {
		return fmt.Errorf("obingslibrary: searching the primers: %w", err)
	}
	search.set = set
	library.search = search
	return nil
}

// primerMatches returns the matches of the primers of all the markers in the
// sequence, by increasing begin and, for the same begin, by marker, then in
// the order in which the separate searches of a marker used to append them
// (forward primer, reverse complement of the reverse primer, reverse primer,
// reverse complement of the forward primer): the pairing of the matches depends
// on the order of those that start at the same place.
//
// For every marker and orientation the matches of the primer of the other end
// (the reverse complement of the reverse primer after a forward one, the one of
// the forward primer after a reverse one) are kept from the begin of the first
// match of the primer of the first end on: those before it cannot pair with
// it, and they would break the pairing of the other markers.
func (s *primerSearch) primerMatches(dst []PrimerMatch, sequence *obiseq.BioSequence) []PrimerMatch {
	occ := s.set.AppendAllMatches(nil, sequence, 0, -1)

	// where the first match of the first end of each marker and orientation
	// starts: first[2*marker + forward]
	first := make([]int, 2*len(s.pairs))
	firstEnd := make([]int, 2*len(s.pairs))
	for i := range firstEnd {
		first[i], firstEnd[i] = math.MinInt, math.MaxInt
	}
	token := func(o obidnapat.Occurrence) (marker int, forward bool) {
		marker = o.Pattern/2 + 1
		forward = (o.Pattern%2 == 0) == (o.Strand > 0)
		return marker, forward
	}
	key := func(marker int, forward bool) int {
		if forward {
			return 2*marker + 1
		}
		return 2 * marker
	}
	for _, o := range occ {
		if o.Strand > 0 {
			marker, forward := token(o)
			if k := key(marker, forward); o.End < firstEnd[k] {
				first[k], firstEnd[k] = o.Start, o.End
			}
		}
	}
	for _, o := range occ {
		marker, forward := token(o)
		if o.Strand < 0 && (firstEnd[key(marker, forward)] == math.MaxInt || o.Start < first[key(marker, forward)]+1) {
			continue
		}
		if o.Strand < 0 {
			marker = -marker
		}
		dst = append(dst, PrimerMatch{Begin: o.Start, End: o.End, Mismatches: o.Errors, Marker: marker, Forward: forward})
	}
	slices.SortStableFunc(dst, func(a, b PrimerMatch) int {
		return cmp.Or(a.Begin-b.Begin, abs(a.Marker)-abs(b.Marker), role(a)-role(b))
	})
	return dst
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

// role is the rank of a match among those of one marker: see primerMatches.
func role(m PrimerMatch) int {
	switch {
	case m.Forward && m.Marker > 0:
		return 0
	case m.Forward:
		return 1
	case m.Marker > 0:
		return 2
	}
	return 3
}
