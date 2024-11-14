package obiseq

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	log "github.com/sirupsen/logrus"
)

func MakeSetTaxonAtRankWorker(taxonomy *obitax.Taxonomy, rank string) SeqWorker {

	if !obiutils.Contains(taxonomy.RankList(), rank) {
		log.Fatalf("%s is not a valid rank (allowed ranks are %v)",
			rank,
			taxonomy.RankList())
	}

	w := func(sequence *BioSequence) (BioSequenceSlice, error) {
		sequence.SetTaxonAtRank(taxonomy, rank)
		return BioSequenceSlice{sequence}, nil
	}

	return w
}

func MakeSetSpeciesWorker(taxonomy *obitax.Taxonomy) SeqWorker {

	w := func(sequence *BioSequence) (BioSequenceSlice, error) {
		sequence.SetSpecies(taxonomy)
		return BioSequenceSlice{sequence}, nil
	}

	return w
}

func MakeSetGenusWorker(taxonomy *obitax.Taxonomy) SeqWorker {

	w := func(sequence *BioSequence) (BioSequenceSlice, error) {
		sequence.SetGenus(taxonomy)
		return BioSequenceSlice{sequence}, nil
	}

	return w
}

func MakeSetFamilyWorker(taxonomy *obitax.Taxonomy) SeqWorker {

	w := func(sequence *BioSequence) (BioSequenceSlice, error) {
		sequence.SetFamily(taxonomy)
		return BioSequenceSlice{sequence}, nil
	}

	return w
}

func MakeSetPathWorker(taxonomy *obitax.Taxonomy) SeqWorker {

	w := func(sequence *BioSequence) (BioSequenceSlice, error) {
		sequence.SetPath(taxonomy)
		return BioSequenceSlice{sequence}, nil
	}

	return w

}
