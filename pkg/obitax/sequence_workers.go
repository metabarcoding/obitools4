package obitax

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	log "github.com/sirupsen/logrus"
)

func (taxonomy *Taxonomy) MakeSetTaxonAtRankWorker(rank string) obiseq.SeqWorker {

	if !obiutils.Contains(taxonomy.RankList(), rank) {
		log.Fatalf("%s is not a valid rank (allowed ranks are %v)",
			rank,
			taxonomy.RankList())
	}

	w := func(sequence *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		taxonomy.SetTaxonAtRank(sequence, rank)
		return obiseq.BioSequenceSlice{sequence}, nil
	}

	return w
}

func (taxonomy *Taxonomy) MakeSetSpeciesWorker() obiseq.SeqWorker {

	w := func(sequence *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		taxonomy.SetSpecies(sequence)
		return obiseq.BioSequenceSlice{sequence}, nil
	}

	return w
}

func (taxonomy *Taxonomy) MakeSetGenusWorker() obiseq.SeqWorker {

	w := func(sequence *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		taxonomy.SetGenus(sequence)
		return obiseq.BioSequenceSlice{sequence}, nil
	}

	return w
}

func (taxonomy *Taxonomy) MakeSetFamilyWorker() obiseq.SeqWorker {

	w := func(sequence *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		taxonomy.SetFamily(sequence)
		return obiseq.BioSequenceSlice{sequence}, nil
	}

	return w
}

func (taxonomy *Taxonomy) MakeSetPathWorker() obiseq.SeqWorker {

	w := func(sequence *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		taxonomy.SetPath(sequence)
		return obiseq.BioSequenceSlice{sequence}, nil
	}

	return w

}
