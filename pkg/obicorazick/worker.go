package obicorazick

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"github.com/rrethy/ahocorasick"
)

func AhoCorazickWorker(slot string, patterns []string) obiseq.SeqWorker {

	matcher := ahocorasick.CompileStrings(patterns)

	fslot := slot + "_Fwd"
	rslot := slot + "_Rev"

	f := func(s *obiseq.BioSequence) *obiseq.BioSequence {
		matchesF := len(matcher.FindAllByteSlice(s.Sequence()))
		matchesR := len(matcher.FindAllByteSlice(s.ReverseComplement(false).Sequence()))

		log.Debugln("Macthes = ", matchesF, matchesR)
		matches := matchesF + matchesR
		if matches > 0 {
			s.SetAttribute(slot, matches)
			s.SetAttribute(fslot, matchesF)
			s.SetAttribute(rslot, matchesR)
		}

		return s
	}

	return f
}

func AhoCorazickPredicate(minMatches int, patterns []string) obiseq.SequencePredicate {

	matcher := ahocorasick.CompileStrings(patterns)

	f := func(s *obiseq.BioSequence) bool {
		matches := matcher.FindAllByteSlice(s.Sequence())
		return len(matches) >= minMatches
	}

	return f
}
