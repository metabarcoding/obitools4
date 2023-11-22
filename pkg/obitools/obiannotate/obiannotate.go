package obiannotate

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obicorazick"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitax"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obigrep"
)

func DeleteAttributesWorker(toBeDeleted []string) obiseq.SeqWorker {
	f := func(s *obiseq.BioSequence) *obiseq.BioSequence {
		for _, k := range toBeDeleted {
			s.DeleteAttribute(k)
		}
		return s
	}

	return f
}

// func MatchPatternWorker(pattern string, errormax int, allowsIndel bool) obiseq.SeqWorker {
// 	pat, err := obiapat.MakeApatPattern(pattern, errormax, allowsIndel)
// 	f := func(s *obiseq.BioSequence) *obiseq.BioSequence {
// 		apats := obiapat.MakeApatSequence(s, false)
// 		pat.BestMatch(apats, 0)
// 		return s
// 	}
// }

func ToBeKeptAttributesWorker(toBeKept []string) obiseq.SeqWorker {

	d := make(map[string]bool, len(_keepOnly))

	for _, v := range _keepOnly {
		d[v] = true
	}

	f := func(s *obiseq.BioSequence) *obiseq.BioSequence {
		annot := s.Annotations()
		for key := range annot {
			if _, ok := d[key]; !ok {
				s.DeleteAttribute(key)
			}
		}
		return s
	}

	return f
}

func CutSequenceWorker(from, to int, breakOnError bool) obiseq.SeqWorker {

	f := func(s *obiseq.BioSequence) *obiseq.BioSequence {
		var f, t int

		switch {
		case from < 0:
			f = s.Len() + from + 1
		case from > 0:
			f = from
		}

		switch {
		case to < 0:
			t = s.Len() + to + 1
		case to > 0:
			t = to
		}

		if from < 0 {
			from = 0
		}

		if to >= s.Len() {
			to = s.Len()
		}

		rep, err := s.Subsequence(f, t, false)
		if err != nil {
			if breakOnError {
				log.Fatalf("Cannot cut sequence %s (%v)", s.Id(), err)
			} else {
				log.Warnf("Cannot cut sequence %s (%v), sequence discarded", s.Id(), err)
				return nil
			}
		}
		return rep
	}

	if from == 0 && to == 0 {
		f = func(s *obiseq.BioSequence) *obiseq.BioSequence {
			return s
		}
	}

	if from > 0 {
		from--
	}

	return f
}

func ClearAllAttributesWorker() obiseq.SeqWorker {
	f := func(s *obiseq.BioSequence) *obiseq.BioSequence {
		annot := s.Annotations()
		for key := range annot {
			s.DeleteAttribute(key)
		}
		return s
	}

	return f
}

func RenameAttributeWorker(toBeRenamed map[string]string) obiseq.SeqWorker {
	f := func(s *obiseq.BioSequence) *obiseq.BioSequence {
		for newName, oldName := range toBeRenamed {
			s.RenameAttribute(newName, oldName)
		}
		return s
	}

	return f
}

func EvalAttributeWorker(expression map[string]string) obiseq.SeqWorker {
	var w obiseq.SeqWorker
	w = nil

	for a, e := range expression {
		if w == nil {
			w = obiseq.EditAttributeWorker(a, e)
		} else {
			w.ChainWorkers(obiseq.EditAttributeWorker(a, e))
		}
	}

	return w
}

func AddTaxonAtRankWorker(taxonomy *obitax.Taxonomy, ranks ...string) obiseq.SeqWorker {
	f := func(s *obiseq.BioSequence) *obiseq.BioSequence {
		for _, r := range ranks {
			taxonomy.SetTaxonAtRank(s, r)
		}
		return s
	}

	return f
}

func AddSeqLengthWorker() obiseq.SeqWorker {
	f := func(s *obiseq.BioSequence) *obiseq.BioSequence {
		s.SetAttribute("seq_length", s.Len())
		return s
	}

	return f

}

func CLIAnnotationWorker() obiseq.SeqWorker {
	var annotator obiseq.SeqWorker
	annotator = nil

	if CLIHasClearAllFlag() {
		w := ClearAllAttributesWorker()
		annotator = annotator.ChainWorkers(w)
	}

	if CLIHasSetId() {
		w := obiseq.EditIdWorker(CLSetIdExpression())
		annotator = annotator.ChainWorkers(w)
	}

	if CLIHasAttibuteToDelete() {
		w := DeleteAttributesWorker(CLIAttibuteToDelete())
		annotator = annotator.ChainWorkers(w)
	}

	if CLIHasToBeKeptAttributes() {
		w := ToBeKeptAttributesWorker(CLIToBeKeptAttributes())
		annotator = annotator.ChainWorkers(w)
	}

	if CLIHasAttributeToBeRenamed() {
		w := RenameAttributeWorker(CLIAttributeToBeRenamed())
		annotator = annotator.ChainWorkers(w)
	}

	if CLIHasTaxonAtRank() {
		taxo := obigrep.CLILoadSelectedTaxonomy()
		w := AddTaxonAtRankWorker(taxo, CLITaxonAtRank()...)
		annotator = annotator.ChainWorkers(w)
	}

	if CLIHasAddLCA() {
		taxo := obigrep.CLILoadSelectedTaxonomy()
		w := obitax.AddLCAWorker(taxo, CLILCASlotName(), CLILCAThreshold())
		annotator = annotator.ChainWorkers(w)
	}

	if CLIHasSetLengthFlag() {
		w := AddSeqLengthWorker()
		annotator = annotator.ChainWorkers(w)
	}

	if CLIHasSetAttributeExpression() {
		w := EvalAttributeWorker(CLISetAttributeExpression())
		annotator = annotator.ChainWorkers(w)
	}

	if CLIHasAhoCorasick() {
		patterns := CLIAhoCorazick()
		log.Println("Matching : ", len(patterns), " patterns on sequences")
		w := obicorazick.AhoCorazickWorker("aho_corasick", patterns)
		log.Println("Automata built")
		annotator = annotator.ChainWorkers(w)
	}

	if CLIHasCut() {
		from, to := CLICut()
		w := CutSequenceWorker(from, to, false)

		annotator = annotator.ChainWorkers(w)
	}

	return annotator
}

func CLIAnnotationPipeline() obiiter.Pipeable {

	predicate := obigrep.CLISequenceSelectionPredicate()
	worker := CLIAnnotationWorker()

	annotator := obiseq.SeqToSliceConditionalWorker(worker, predicate, true, false)
	f := obiiter.SliceWorkerPipe(annotator, obioptions.CLIParallelWorkers())

	return f
}
