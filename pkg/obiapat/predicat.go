package obiapat

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	log "github.com/sirupsen/logrus"
)

func IsPatternMatchSequence(pattern string, errormax int, bothStrand, allowIndels bool) obiseq.SequencePredicate {

	pat, err := MakeApatPattern(pattern, errormax, allowIndels)

	if err != nil {
		log.Fatalf("error in sequence regular pattern syntax : %v", err)
	}

	cpat, err := pat.ReverseComplement()

	if err != nil {
		log.Fatalf("cannot reverse complement the pattern : %v", err)
	}

	f := func(sequence *obiseq.BioSequence) bool {
		aseq, err := MakeApatSequence(sequence, false)

		if err != nil {
			log.Panicf("Cannot convert sequence %s to apat format", sequence.Id())
		}

		match := pat.IsMatching(aseq, 0, aseq.Len())

		if !match && bothStrand {

			match = cpat.IsMatching(aseq, 0, aseq.Len())
		}

		return match
	}

	return f
}
