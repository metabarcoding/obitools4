package obitax

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/goutils"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	log "github.com/sirupsen/logrus"
)

func (taxonomy *Taxonomy) MakeSetTaxonAtRankWorker(rank string) obiiter.SeqWorker {

	if !goutils.Contains(taxonomy.RankList(), rank) {
		log.Fatalf("%s is not a valid rank (allowed ranks are %v)",
			rank,
			taxonomy.RankList())
	}

	w := func(sequence *obiseq.BioSequence) *obiseq.BioSequence {
		taxonomy.SetTaxonAtRank(sequence, rank)
		return sequence
	}

	return w
}

func (taxonomy *Taxonomy) MakeSetSpeciesWorker() obiiter.SeqWorker {

	w := func(sequence *obiseq.BioSequence) *obiseq.BioSequence {
		taxonomy.SetSpecies(sequence)
		return sequence
	}

	return w
}

func (taxonomy *Taxonomy) MakeSetGenusWorker() obiiter.SeqWorker {

	w := func(sequence *obiseq.BioSequence) *obiseq.BioSequence {
		taxonomy.SetGenus(sequence)
		return sequence
	}

	return w
}

func (taxonomy *Taxonomy) MakeSetFamilyWorker() obiiter.SeqWorker {

	w := func(sequence *obiseq.BioSequence) *obiseq.BioSequence {
		taxonomy.SetFamily(sequence)
		return sequence
	}

	return w
}


