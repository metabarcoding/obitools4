package obipairing

import (
	"log"
	"math"
	"os"
	"runtime"
	"time"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obialign"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"github.com/schollz/progressbar/v3"
)

func _Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func JoinPairedSequence(seqA, seqB obiseq.BioSequence, inplace bool) obiseq.BioSequence {

	if !inplace {
		seqA = seqA.Copy()
	}

	seqA.WriteString("..........")
	seqA.Write(seqB.Sequence())

	seqA.WriteQualities(obiseq.Quality{0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	seqA.WriteQualities(seqB.Qualities())

	return seqA
}

// AssemblePESequences assembles two paired sequences following
// the obipairing strategy implemented in obialign.PEAlign using
// the gap and delta parametters.
// If the length of the overlap between both sequences is less than
// overlap_min, The alignment is substituted by a simple pasting
// of the sequences with a strech of 10 dots in between them.
// the quality of the dots is set to 0.
// If the inplace parameter is set to true, the seqA and seqB are
// destroyed during the assembling process and cannot be reuse later on.
func AssemblePESequences(seqA, seqB obiseq.BioSequence,
	gap, delta, overlap_min int, with_stats bool,
	inplace bool,
	arena_align obialign.PEAlignArena,
	arena_cons obialign.BuildAlignArena,
	arena_qual obialign.BuildAlignArena) obiseq.BioSequence {

	score, path := obialign.PEAlign(seqA, seqB, gap, delta, arena_align)
	cons, match := obialign.BuildQualityConsensus(seqA, seqB, path,
		arena_cons, arena_qual)

	left := path[0]
	right := 0
	if path[len(path)-1] == 0 {
		right = path[len(path)-2]
	}
	lcons := cons.Length()
	ali_length := lcons - _Abs(left) - _Abs(right)

	if ali_length >= overlap_min {
		if with_stats {
			annot := cons.Annotations()
			annot["mode"] = "alignment"
			annot["score"] = score

			if left < 0 {
				annot["seq_a_single"] = -left
				annot["ali_dir"] = "left"
			} else {
				annot["seq_b_single"] = left
				annot["ali_dir"] = "right"
			}

			if right < 0 {
				right = -right
				annot["seq_a_single"] = right
			} else {
				annot["seq_b_single"] = right
			}

			score_norm := float64(0)
			if ali_length > 0 {
				score_norm = math.Round(float64(match)/float64(ali_length)*1000) / 1000
			}

			annot["ali_length"] = ali_length
			annot["seq_ab_match"] = match
			annot["score_norm"] = score_norm

			if inplace {
				(&seqA).Recycle()
				(&seqB).Recycle()
			}
		}
	} else {
		cons = JoinPairedSequence(seqA, seqB, inplace)

		if with_stats {
			annot := cons.Annotations()
			annot["mode"] = "join"
		}

		if inplace {
			(&seqB).Recycle()
		}
	}

	return cons
}

func IAssemblePESequencesBatch(iterator obiseq.IPairedBioSequenceBatch,
	gap, delta, overlap_min int, with_stats bool, sizes ...int) obiseq.IBioSequenceBatch {

	nworkers := runtime.NumCPU() - 1
	buffsize := iterator.BufferSize()

	if len(sizes) > 0 {
		nworkers = sizes[0]
	}

	if len(sizes) > 1 {
		buffsize = sizes[1]
	}

	newIter := obiseq.MakeIBioSequenceBatch(buffsize)

	newIter.Add(nworkers)

	go func() {
		newIter.Wait()
		for len(newIter.Channel()) > 0 {
			time.Sleep(time.Millisecond)
		}
		close(newIter.Channel())
		log.Printf("End of the sequence Pairing")
	}()

	bar := progressbar.NewOptions(
		-1,
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetWidth(15),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetDescription("[Sequence Pairing]"))

	f := func(iterator obiseq.IPairedBioSequenceBatch, wid int) {
		arena := obialign.MakePEAlignArena(150, 150)
		barena1 := obialign.MakeBuildAlignArena(150, 150)
		barena2 := obialign.MakeBuildAlignArena(150, 150)

		// log.Printf("\n==> %d Wait data to align\n", wid)
		// start := time.Now()
		for iterator.Next() {
			// elapsed := time.Since(start)
			// log.Printf("\n==>%d got data to align after %s\n", wid, elapsed)
			batch := iterator.Get()
			cons := make(obiseq.BioSequenceSlice, len(batch.Forward()))
			processed := 0
			for i, A := range batch.Forward() {
				B := batch.Reverse()[i]
				cons[i] = AssemblePESequences(A, B, 2, 5, 20, true, true, arena, barena1, barena2)
				if i%59 == 0 {
					bar.Add(59)
					processed += 59
				}
			}
			bar.Add(batch.Length() - processed)
			newIter.Channel() <- obiseq.MakeBioSequenceBatch(
				batch.Order(),
				cons...,
			)
			// log.Printf("\n==> %d Wait data to align\n", wid)
			// start = time.Now()
		}
		newIter.Done()
	}

	log.Printf("Start of the sequence Pairing")

	for i := 0; i < nworkers-1; i++ {
		go f(iterator.Split(), i)
	}
	go f(iterator, nworkers-1)

	return newIter

}
