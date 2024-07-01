package obicleandb

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obialign"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obichunk"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obistats"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obigrep"
)

func SequenceTrust(sequence *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
	sequence.SetAttribute("obicleandb_trusted", 1.0-1.0/float64(sequence.Count()+1))
	sequence.SetAttribute("obicleandb_trusted_on", float64(sequence.Count()))
	return obiseq.BioSequenceSlice{sequence}, nil
}

func diagCoord(x, y, n int) int {
	if x > y {
		x, y = y, x
	}

	if x == y {
		log.Panicf("diagCoord: (%d == %d)", x, y)
	}

	sn := n * (n - 1) / 2
	sx := (n - x) * (n - (1 + x)) / 2
	s := sn - sx

	return s + y - x - 1
}
func SequenceTrustSlice(sequences obiseq.BioSequenceSlice) (obiseq.BioSequenceSlice, error) {
	n := len(sequences)
	score := make([]float64, n*(n-1)/2)
	matrix := make([]uint64, sequences[0].Len()*sequences[0].Len())

	for i, sa := range sequences {
		for j, sb := range sequences[i+1:] {
			lca, lali := obialign.FastLCSScore(sa, sb, -1, &matrix)
			score[diagCoord(i, i+1+j, n)] = float64(lca) / float64(lali)
		}
	}

	for i, sa := range sequences {
		ss := make([]float64, 0, n-1)
		for j, _ := range sequences {
			if i == j {
				continue
			}

			s := score[diagCoord(i, j, n)]
			if s == 0.0 {
				log.Panicf("score[%d, %d] == 0.0", i, j)
			}
			ss = append(ss, score[diagCoord(i, j, n)])

		}
		sa.SetAttribute("obicleandb_dist", ss)
	}

	scoremed := obistats.Median(score)
	scorethr := 1 - 2*(1-scoremed)
	mednorm := (scoremed - scorethr) / 2.0

	for i, s := range score {
		switch {
		case s < scorethr:
			score[i] = -1.0
		case s < scoremed:
			score[i] = (s-scorethr)/mednorm - 1.0
		default:
			score[i] = 1.0
		}
	}

	// Tylos
	for i, sa := range sequences {
		ngroup := float64(sa.Count())
		ss := make([]float64, 0, n-1)
		sc := sa.Count()
		for j, sb := range sequences {
			if i == j {
				continue
			}

			ss = append(ss, score[diagCoord(i, j, n)])
			sc += sb.Count()
			nt, _ := sb.GetFloatAttribute("obicleandb_trusted_on")
			ngroup += score[diagCoord(i, j, n)] * nt
		}
		ngroup = max(0, ngroup)
		sa.SetAttribute("obicleandb_trusted", 1.0-1.0/float64(ngroup+1))
		sa.SetAttribute("obicleandb_on", ngroup)
		sa.SetAttribute("obicleandb_median", scoremed)
		sa.SetAttribute("obicleandb_ss", ss)
	}

	return sequences, nil
}

func ICleanDB(itertator obiiter.IBioSequence) obiiter.IBioSequence {
	var rankPredicate obiseq.SequencePredicate

	options := make([]obichunk.WithOption, 0, 30)

	// Make sequence dereplication with a constraint on the taxid.
	// To be merged, both sequences must have the same taxid.

	options = append(options,
		obichunk.OptionBatchCount(100),
		obichunk.OptionSortOnMemory(),
		obichunk.OptionSubCategory("taxid"),
		obichunk.OptionsParallelWorkers(
			obioptions.CLIParallelWorkers()),
		obichunk.OptionsBatchSize(
			obioptions.CLIBatchSize()),
		obichunk.OptionNAValue("NA"),
	)

	unique, err := obichunk.IUniqueSequence(itertator, options...)

	if err != nil {
		log.Fatal(err)
	}

	taxonomy := obigrep.CLILoadSelectedTaxonomy()

	if len(obigrep.CLIRequiredRanks()) > 0 {
		rankPredicate = obigrep.CLIHasRankDefinedPredicate()
	} else {
		rankPredicate = taxonomy.HasRequiredRank("species").And(taxonomy.HasRequiredRank("genus")).And(taxonomy.HasRequiredRank("family"))
	}

	goodTaxa := taxonomy.IsAValidTaxon(CLIUpdateTaxids()).And(rankPredicate)

	usable := unique.FilterOn(goodTaxa,
		obioptions.CLIBatchSize(),
		obioptions.CLIParallelWorkers())

	annotated := usable.MakeIWorker(taxonomy.MakeSetSpeciesWorker(),
		false,
		obioptions.CLIParallelWorkers(),
	).MakeIWorker(taxonomy.MakeSetGenusWorker(),
		false,
		obioptions.CLIParallelWorkers(),
	).MakeIWorker(taxonomy.MakeSetFamilyWorker(),
		false,
		obioptions.CLIParallelWorkers(),
	).MakeIWorker(SequenceTrust,
		false,
		obioptions.CLIParallelWorkers(),
	)

	genera_iterator, err := obichunk.ISequenceChunk(
		annotated,
		obiseq.AnnotationClassifier("genus_taxid", "NA"),
	)

	if err != nil {
		log.Fatal(err)
	}

	trusted := genera_iterator.MakeISliceWorker(
		SequenceTrustSlice,
		false,
	)

	return trusted
}
