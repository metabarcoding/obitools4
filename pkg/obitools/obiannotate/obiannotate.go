package obiannotate

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiapat"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obicorazick"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obigrep"
)

func DeleteAttributesWorker(toBeDeleted []string) obiseq.SeqWorker {
	f := func(s *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		for _, k := range toBeDeleted {
			s.DeleteAttribute(k)
		}
		return obiseq.BioSequenceSlice{s}, nil
	}

	return f
}

func MatchPatternWorker(pattern, name string, errormax int, bothStrand, allowsIndel bool) obiseq.SeqWorker {
	pat, err := obiapat.MakeApatPattern(pattern, errormax, allowsIndel)
	if err != nil {
		log.Fatalf("error in compiling pattern (%s) : %v", pattern, err)
	}

	cpat, err := pat.ReverseComplement()

	if err != nil {
		log.Fatalf("error in reverse-complementing pattern (%s) : %v", pattern, err)
	}

	slot := "pattern"
	if name != "pattern" && name != "" {
		slot = fmt.Sprintf("%s_pattern", name)
	} else {
		name = "pattern"
	}

	slot_match := fmt.Sprintf("%s_match", name)
	slot_error := fmt.Sprintf("%s_error", name)
	slot_location := fmt.Sprintf("%s_location", name)

	f := func(s *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		apats, err := obiapat.MakeApatSequence(s, false)
		if err != nil {
			log.Fatalf("error in preparing sequence %s : %v", s.Id(), err)
		}

		start, end, nerr, matched := pat.BestMatch(apats, 0, s.Len())

		if matched && start >= 0 && end <= s.Len() {
			annot := s.Annotations()
			annot[slot] = pattern

			if start < 0 {
				start = 0
			}

			match, err := s.Subsequence(start, end, false)
			if err != nil {
				log.Fatalf("Error in extracting pattern of sequence %s [%d;%d[ : %v",
					s.Id(), start, end, err)
			}
			annot[slot_match] = match.String()
			annot[slot_error] = nerr
			annot[slot_location] = fmt.Sprintf("%d..%d", start+1, end)
		} else {
			start, end, nerr, matched := cpat.BestMatch(apats, 0, s.Len())

			if matched && start >= 0 && end <= s.Len() {
				annot := s.Annotations()
				annot[slot] = pattern
				match, err := s.Subsequence(start, end, false)
				if err != nil {
					log.Fatalf("Error in extracting pattern of sequence %s [%d;%d[ : %v",
						s.Id(), start, end, err)
				}
				annot[slot_match] = match.ReverseComplement(true).String()
				annot[slot_error] = nerr
				annot[slot_location] = fmt.Sprintf("complement(%d..%d)", start+1, end)
			}
		}
		return obiseq.BioSequenceSlice{s}, nil
	}

	return f
}

func ToBeKeptAttributesWorker(toBeKept []string) obiseq.SeqWorker {

	d := make(map[string]bool, len(_keepOnly))

	for _, v := range _keepOnly {
		d[v] = true
	}

	f := func(s *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		annot := s.Annotations()
		for key := range annot {
			if _, ok := d[key]; !ok {
				s.DeleteAttribute(key)
			}
		}
		return obiseq.BioSequenceSlice{s}, nil
	}

	return f
}

func CutSequenceWorker(from, to int, breakOnError bool) obiseq.SeqWorker {

	f := func(s *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
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
				err = fmt.Errorf("Cannot cut sequence %s (%v), sequence discarded", s.Id(), err)
			}
		}
		return obiseq.BioSequenceSlice{rep}, err
	}

	if from == 0 && to == 0 {
		f = func(s *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
			return obiseq.BioSequenceSlice{s}, nil
		}
	}

	if from > 0 {
		from--
	}

	return f
}

func ClearAllAttributesWorker() obiseq.SeqWorker {
	f := func(s *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		annot := s.Annotations()
		for key := range annot {
			s.DeleteAttribute(key)
		}
		return obiseq.BioSequenceSlice{s}, nil
	}

	return f
}

func RenameAttributeWorker(toBeRenamed map[string]string) obiseq.SeqWorker {
	f := func(s *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		for newName, oldName := range toBeRenamed {
			s.RenameAttribute(newName, oldName)
		}
		return obiseq.BioSequenceSlice{s}, nil
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
	f := func(s *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		for _, r := range ranks {
			taxonomy.SetTaxonAtRank(s, r)
		}
		return obiseq.BioSequenceSlice{s}, nil
	}

	return f
}

func AddTaxonRankWorker(taxonomy *obitax.Taxonomy) obiseq.SeqWorker {
	f := func(s *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		taxonomy.SetTaxonomicRank(s)
		return obiseq.BioSequenceSlice{s}, nil
	}

	return f
}

func AddScientificNameWorker(taxonomy *obitax.Taxonomy) obiseq.SeqWorker {
	f := func(s *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		taxonomy.SetScientificName(s)
		return obiseq.BioSequenceSlice{s}, nil
	}

	return f
}

func AddSeqLengthWorker() obiseq.SeqWorker {
	f := func(s *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		s.SetAttribute("seq_length", s.Len())
		return obiseq.BioSequenceSlice{s}, nil
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

	if CLISetTaxonomicPath() {
		taxo := obigrep.CLILoadSelectedTaxonomy()
		w := taxo.MakeSetPathWorker()
		annotator = annotator.ChainWorkers(w)
	}

	if CLISetTaxonomicRank() {
		taxo := obigrep.CLILoadSelectedTaxonomy()
		w := AddTaxonRankWorker(taxo)
		annotator = annotator.ChainWorkers(w)
	}

	if CLISetScientificName() {
		taxo := obigrep.CLILoadSelectedTaxonomy()
		w := AddScientificNameWorker(taxo)
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

	if CLIHasPattern() {
		log.Infof("Match pattern %s with %d error", CLIPattern(), obigrep.CLIPatternError())
		w := MatchPatternWorker(CLIPattern(), CLIHasPatternName(),
			obigrep.CLIPatternError(), obigrep.CLIPatternBothStrand(),
			obigrep.CLIPatternInDels())

		annotator = annotator.ChainWorkers(w)
	}

	return annotator
}

func CLIAnnotationPipeline() obiiter.Pipeable {

	predicate := obigrep.CLISequenceSelectionPredicate()
	worker := CLIAnnotationWorker()

	annotator := obiseq.SeqToSliceConditionalWorker(predicate, worker, false)
	f := obiiter.SliceWorkerPipe(annotator, false, obioptions.CLIParallelWorkers())

	return f
}
