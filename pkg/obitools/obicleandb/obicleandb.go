package obicleandb

import (
	"math/rand"

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

func MakeSequenceFamilyGenusWorker(references obiseq.BioSequenceSlice) obiseq.SeqWorker {

	genus := make(map[int]*obiseq.BioSequenceSlice)
	family := make(map[int]*obiseq.BioSequenceSlice)

	for _, ref := range references {
		g, _ := ref.GetIntAttribute("genus_taxid")
		f, _ := ref.GetIntAttribute("family_taxid")

		gs, ok := genus[g]
		if !ok {
			gs = obiseq.NewBioSequenceSlice(0)
			genus[g] = gs
		}

		*gs = append(*gs, ref)

		fs, ok := family[f]
		if !ok {
			fs = obiseq.NewBioSequenceSlice(0)
			family[f] = fs
		}

		*fs = append(*fs, ref)
	}

	f := func(sequence *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		sequence.SetAttribute("obicleandb_level", "none")
		pval := 0.0

		g, _ := sequence.GetIntAttribute("genus_taxid")
		sequence.SetAttribute("obicleandb_level", "genus")

		gs := genus[g]

		indist := make([]float64, 0, gs.Len())
		for _, s := range *gs {
			if s != sequence {
				lca, lali := obialign.FastLCSScore(sequence, s, -1, nil)
				indist = append(indist, float64(lali-lca))
			}
		}
		nindist := len(indist)

		f, _ := sequence.GetIntAttribute("family_taxid")
		fs := family[f]

		if nindist < 5 {
			sequence.SetAttribute("obicleandb_level", "family")

			for _, s := range *fs {
				gf, _ := s.GetIntAttribute("genus_taxid")
				if g != gf {
					lca, lali := obialign.FastLCSScore(sequence, s, -1, nil)
					indist = append(indist, float64(lali-lca))
				}
			}

			nindist = len(indist)
		}

		if nindist > 0 {

			next := nindist
			if next <= 20 {
				next = 20
			}

			outdist := make([]float64, 0, next)
			p := rand.Perm(references.Len())
			i := 0
			for _, ir := range p {
				s := references[ir]
				ff, _ := s.GetIntAttribute("family_taxid")

				if ff != f {
					lca, lali := obialign.FastLCSScore(sequence, s, -1, nil)
					outdist = append(outdist, float64(lali-lca))
					i += 1
					if i >= next {
						break
					}
				}
			}

			res, err := obistats.MannWhitneyUTest(outdist, indist, obistats.LocationGreater)

			if err == nil {
				pval = res.P
			}

			// level, _ := sequence.GetAttribute("obicleandb_level")
			// log.Warnf("%s - level: %v", sequence.Id(), level)
			// log.Warnf("%s - gdist: %v", sequence.Id(), indist)
			// log.Warnf("%s - fdist: %v", sequence.Id(), outdist)
			// log.Warnf("%s - pval: %f", sequence.Id(), pval)
		}

		if pval < 0.0 {
			pval = 0.0
		}

		sequence.SetAttribute("obicleandb_trusted", pval)

		return obiseq.BioSequenceSlice{sequence}, nil
	}

	return f

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

	if n > 1 {
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
		scorethr := 1 - 3*(1-scoremed)
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
			ss := make(map[string]float64, n-1)
			sc := sa.Count()
			for j, sb := range sequences {
				if i == j {
					continue
				}

				ss[sb.Id()] = score[diagCoord(i, j, n)]
				sc += sb.Count()
				nt, _ := sb.GetFloatAttribute("obicleandb_trusted_on")
				ngroup += score[diagCoord(i, j, n)] * nt
			}
			ngroup = max(0, ngroup)
			sa.SetAttribute("obicleandb_trusted", 1.0-1.0/float64(ngroup+1))
			sa.SetAttribute("obicleandb_trusted_on", ngroup)
			sa.SetAttribute("obicleandb_median", scoremed)
			sa.SetAttribute("obicleandb_scores", ss)
		}
	} else {
		sequences[0].SetAttribute("obicleandb_median", 1.0)
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
	)
	// .MakeIWorker(SequenceTrust,
	// 	false,
	// 	obioptions.CLIParallelWorkers(),
	// )

	references := annotated.Load()

	mannwithney := MakeSequenceFamilyGenusWorker(references)

	partof := obiiter.IBatchOver(references,
		obioptions.CLIBatchSize())

	// genera_iterator, err := obichunk.ISequenceChunk(
	// 	annotated,
	// 	obiseq.AnnotationClassifier("genus_taxid", "NA"),
	// )

	// if err != nil {
	// 	log.Fatal(err)
	// }

	// trusted := genera_iterator.MakeISliceWorker(
	// 	SequenceTrustSlice,
	// 	false,
	// )

	return partof.MakeIWorker(mannwithney, true).Speed("Testing belonging", references.Len())
}
