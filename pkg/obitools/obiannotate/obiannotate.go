package obiannotate

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
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

func ToBeKeptAttributesWorker(toBeKept map[string]bool) obiseq.SeqWorker {

	f := func(s *obiseq.BioSequence) *obiseq.BioSequence {
		annot := s.Annotations()
		for key := range annot {
			if _, ok := toBeKept[key]; !ok {
				s.DeleteAttribute(key)
			}
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

func CLIAnnotationWorker() obiseq.SeqWorker {
	var annotator obiseq.SeqWorker
	annotator = nil

	if CLIHasAttributeToBeRenamed() {
		w := RenameAttributeWorker(CLIAttributeToBeRenamed())
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

	return annotator
}

func CLIAnnotationPipeline() obiiter.Pipeable {

	predicate := obigrep.CLISequenceSelectionPredicate()
	worker := CLIAnnotationWorker()

	annotator := obiseq.SeqToSliceConditionalWorker(worker, predicate, true)
	f := obiiter.SliceWorkerPipe(annotator)

	return f
}
