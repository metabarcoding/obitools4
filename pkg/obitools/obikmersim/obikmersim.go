package obikmersim

import (
	"math"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obialign"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obifp"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

func _Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func MakeCountMatchWorker[T obifp.FPUint[T]](k *obikmer.KmerMap[T], minKmerCount int) obiseq.SeqWorker {
	return func(sequence *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		matches := k.Query(sequence)
		matches.FilterMinCount(minKmerCount)
		n := matches.Len()

		sequence.SetAttribute("obikmer_match_count", n)
		sequence.SetAttribute("obikmer_kmer_size", k.Kmersize)
		sequence.SetAttribute("obikmer_sparse_kmer", k.SparseAt >= 0)

		return obiseq.BioSequenceSlice{sequence}, nil
	}
}

func MakeKmerAlignWorker[T obifp.FPUint[T]](
	k *obikmer.KmerMap[T],
	minKmerCount int,
	gap, scale float64, delta int, fastScoreRel bool,
	minIdentity float64, withStats bool) obiseq.SeqWorker {
	return func(sequence *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		arena := obialign.MakePEAlignArena(150, 150)
		shifts := make(map[int]int)

		matches := k.Query(sequence)
		matches.FilterMinCount(minKmerCount)

		slice := obiseq.NewBioSequenceSlice(matches.Len())
		*slice = (*slice)[:0]

		candidates := matches.Sequences()
		ks := k.Kmersize
		if k.SparseAt >= 0 {
			ks--
		}

		n := candidates.Len()
		for _, seq := range candidates {
			idmatched_id := seq.Id()

			score, path, fastcount, over, fastscore, directAlignment := obialign.ReadAlign(
				sequence, seq,
				gap, scale, delta,
				fastScoreRel,
				arena, &shifts,
			)

			if !directAlignment {
				idmatched_id = idmatched_id + "-rev"
				seq = seq.ReverseComplement(false)
			}

			cons, match := obialign.BuildQualityConsensus(sequence, seq, path, true, arena)

			left := path[0]
			right := 0
			if path[len(path)-1] == 0 {
				right = path[len(path)-2]
			}
			lcons := cons.Len()
			aliLength := lcons - _Abs(left) - _Abs(right)
			identity := float64(match) / float64(aliLength)
			residual := float64(match-int(ks)) / float64(aliLength-int(ks))
			if aliLength == 0 {
				identity = 0
			}

			rep := cons

			rep.SetAttribute("obikmer_match_count", n)
			rep.SetAttribute("obikmer_match_id", idmatched_id)
			rep.SetAttribute("obikmer_fast_count", fastcount)
			rep.SetAttribute("obikmer_fast_overlap", over)
			rep.SetAttribute("obikmer_fast_score", math.Round(fastscore*1000)/1000)
			rep.SetAttribute("seq_length", cons.Len())

			if directAlignment {
				rep.SetAttribute("obikmer_orientation", "forward")
			} else {
				rep.SetAttribute("obikmer_orientation", "reverse")
			}

			if aliLength >= int(k.KmerSize()) && residual >= minIdentity {
				if withStats {
					if left < 0 {
						rep.SetAttribute("seq_a_single", -left)
						rep.SetAttribute("ali_dir", "left")
					} else {
						rep.SetAttribute("seq_b_single", left)
						rep.SetAttribute("ali_dir", "right")
					}

					if right < 0 {
						right = -right
						rep.SetAttribute("seq_a_single", right)
					} else {
						rep.SetAttribute("seq_b_single", right)
					}
					rep.SetAttribute("obikmer_score", score)
					rep.SetAttribute("obikmer_identity", identity)
					rep.SetAttribute("obikmer_residual_id", residual)
					scoreNorm := float64(0)
					if aliLength > 0 {
						scoreNorm = math.Round(float64(match)/float64(aliLength)*1000) / 1000
					} else {
						scoreNorm = 0
					}

					rep.SetAttribute("obikmer_score_norm", scoreNorm)
					rep.SetAttribute("obikmer_ali_length", aliLength)

					rep.SetAttribute("seq_ab_match", match)

				}

				*slice = append(*slice, rep)
			}

		}

		return *slice, nil
	}
}

func CLILookForSharedKmers(iterator obiiter.IBioSequence) obiiter.IBioSequence {
	var newIter obiiter.IBioSequence

	source, references := CLIReference()

	if iterator == obiiter.NilIBioSequence {
		iterator = obiiter.IBatchOver(source, references, obioptions.CLIBatchSize())
	}

	if CLISelf() {
		iterator = iterator.Speed("Counting similar reads", references.Len())
	} else {
		iterator = iterator.Speed("Counting similar reads")
	}

	kmerMatch := obikmer.NewKmerMap[obifp.Uint128](
		references,
		uint(CLIKmerSize()),
		CLISparseMode(),
		CLIMaxKmerOccurs())

	worker := MakeCountMatchWorker(kmerMatch, CLIMinSharedKmers())
	newIter = iterator.MakeIWorker(worker, false, obioptions.CLIParallelWorkers())

	return newIter.FilterEmpty()
}

func CLIAlignSequences(iterator obiiter.IBioSequence) obiiter.IBioSequence {
	var newIter obiiter.IBioSequence

	source, references := CLIReference()

	if iterator == obiiter.NilIBioSequence {
		iterator = obiiter.IBatchOver(source, references, obioptions.CLIBatchSize())
	}

	if CLISelf() {
		iterator = iterator.Speed("Aligning reads", references.Len())
	} else {
		iterator = iterator.Speed("Aligning reads")
	}
	kmerMatch := obikmer.NewKmerMap[obifp.Uint128](
		references,
		uint(CLIKmerSize()),
		CLISparseMode(),
		CLIMaxKmerOccurs())
	worker := MakeKmerAlignWorker(kmerMatch, CLIMinSharedKmers(), CLIGap(), CLIScale(), CLIDelta(), CLIFastRelativeScore(), 0.8, true)
	newIter = iterator.MakeIWorker(worker, false, obioptions.CLIParallelWorkers())

	return newIter.FilterEmpty()
}
