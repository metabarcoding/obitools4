package obingslibrary

import "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"

// oldPrimerMatches is the search of the primers before the joint search: four
// AllMatches per marker, the reverse primers looked for from the start of the
// first match of the other end only. The markers are numbered like the set does.
func (library *NGSLibrary) OldPrimerMatches(sequence *obiseq.BioSequence) []PrimerMatch {
	var matches []PrimerMatch
	for i := 1; i < len(library.search.pairs); i++ {
		marker := library.Markers[library.search.pairs[i]]
		for _, side := range []struct {
			first, other interface {
				AllMatches(*obiseq.BioSequence, int, int) [][3]int
			}
			forward bool
		}{{marker.forward, marker.creverse, true}, {marker.reverse, marker.cforward, false}} {
			locs := side.first.AllMatches(sequence, 0, -1)
			for _, l := range locs {
				matches = append(matches, PrimerMatch{l[0], l[1], l[2], i, side.forward})
			}
			if len(locs) > 0 {
				for _, l := range side.other.AllMatches(sequence, locs[0][0]+1, -1) {
					matches = append(matches, PrimerMatch{l[0], l[1], l[2], -i, side.forward})
				}
			}
		}
	}
	return matches
}

// PrimerMatchesOf is the joint search of the primers, for the tests.
func (library *NGSLibrary) PrimerMatchesOf(sequence *obiseq.BioSequence) []PrimerMatch {
	return library.search.primerMatches(nil, sequence)
}
