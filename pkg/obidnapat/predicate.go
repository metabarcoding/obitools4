package obidnapat

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// IsPatternMatchSequence returns the predicate that tells whether the pattern
// matches the sequence with at most errormax errors (insertions and deletions
// too if allowIndels), on the sequence itself and, if bothStrand, on its reverse
// complement. It stops the program if the pattern cannot be compiled.
func IsPatternMatchSequence(pattern string, errormax int, bothStrand, allowIndels bool) obiseq.SequencePredicate {
	pat, err := MakeApatPattern(pattern, errormax, allowIndels)

	if err != nil {
		log.Fatalf("error in sequence regular pattern syntax : %v", err)
	}

	cpat, err := pat.ReverseComplement()

	if err != nil {
		log.Fatalf("cannot reverse complement the pattern : %v", err)
	}

	return func(sequence *obiseq.BioSequence) bool {
		match := pat.IsMatching(sequence, 0, sequence.Len())

		if !match && bothStrand {
			match = cpat.IsMatching(sequence, 0, sequence.Len())
		}

		return match
	}
}
