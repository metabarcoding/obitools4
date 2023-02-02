package obiannotate

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
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

	if CLIHasSetLengthFlag() {
		w := AddSeqLengthWorker()
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
