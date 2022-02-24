package obipairing

import (
	log "github.com/sirupsen/logrus"
	"math"
	"os"
	"runtime"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obialign"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"github.com/schollz/progressbar/v3"
)

func _Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// JoinPairedSequence paste two sequences.
//
// Both input sequences are pasted and 10 dots are used as separator.
// if both sequences have quality scores, a quality of 0 is associated
// to the added dots.
//
// Parameters
//
// - seqA, seqB: the pair of sequences to align.
//
// - inplace: if  is set to true, the seqA and seqB are
// destroyed during the assembling process and cannot be reuse later on.
// the gap and delta parametters.
//
// Returns
//
// An obiseq.BioSequence corresponding to the pasting of the both
// input sequences.
//
// Examples:
//
// .
//  seqA := obiseq.BioSequence("A","cgatgcta","Sequence A")
//  seqB := obiseq.BioSequence("B","aatcgtacga","Sequence B")
//  seqC := obipairing.JoinPairedSequence(seqA, seqB, false)
//  fmt.Println(seqC.String())
//
// Outputs:
//  cgatgcta..........aatcgtacga
//
func JoinPairedSequence(seqA, seqB *obiseq.BioSequence, inplace bool) *obiseq.BioSequence {

	if !inplace {
		seqA = seqA.Copy()
	}

	seqA.WriteString("..........")
	seqA.Write(seqB.Sequence())

	if seqA.HasQualities() && seqB.HasQualities() {
		seqA.WriteQualities(obiseq.Quality{0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
		seqA.WriteQualities(seqB.Qualities())
	}

	if inplace {
		seqB.Recycle()
	}

	return seqA
}

// AssemblePESequences assembles two paired sequences.
//
// The function assembles two paired sequences following
// the obipairing strategy implemented in obialign.PEAlign.
// If the alignment does not result in an overlap of at least
// a given length, it is discarded and booth sequences are only
// pasted using the obipairing.JoinPairedSequence function.
//
// Parameters
//
// - seqA, seqB: the pair of sequences to align.
//
// - gap: the gap penality is expressed as a multiplicator factor of the cost
// of a mismatch between two bases having a quality score of 40.
//
// - delta: the extension in number of base pairs added on both sides of the
// overlap detected by the FAST algorithm before the optimal alignment.
//
// - minOverlap: the minimal length of the overlap to accept the alignment of
// the paired reads as correct. If the actual length is below this limit. The
// the alignment is discarded and both sequences are pasted.
//
// - withStats: indicates (true value) if the algorithm adds annotation to each
// sequence on the quality of the aligned overlap.
//
// - inplace: if  is set to true, the seqA and seqB are
// destroyed during the assembling process and cannot be reuse later on.
// the gap and delta parametters.
//
// Returns
//
// An obiseq.BioSequence corresponding to the assembling of the both
// input sequence.
//
func AssemblePESequences(seqA, seqB *obiseq.BioSequence,
	gap float64, delta, minOverlap int, minIdentity float64, withStats bool,
	inplace bool,
	arenaAlign obialign.PEAlignArena) *obiseq.BioSequence {

	score, path := obialign.PEAlign(seqA, seqB, gap, delta, arenaAlign)
	cons, match := obialign.BuildQualityConsensus(seqA, seqB, path)

	left := path[0]
	right := 0
	if path[len(path)-1] == 0 {
		right = path[len(path)-2]
	}
	lcons := cons.Length()
	aliLength := lcons - _Abs(left) - _Abs(right)
	identity := float64(match) / float64(aliLength)

	if aliLength >= minOverlap && identity >= minIdentity {
		if withStats {
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

			scoreNorm := float64(0)
			if aliLength > 0 {
				scoreNorm = math.Round(float64(match)/float64(aliLength)*1000) / 1000
			}

			annot["ali_length"] = aliLength
			annot["seq_ab_match"] = match
			annot["score_norm"] = scoreNorm

			if inplace {
				seqA.Recycle()
				seqB.Recycle()
			}
		}
	} else {
		cons = JoinPairedSequence(seqA, seqB, inplace)

		if withStats {
			annot := cons.Annotations()
			annot["mode"] = "join"
		}

	}

	return cons
}

// IAssemblePESequencesBatch aligns paired reads.
//
// The function consumes an iterator over batches of paired sequences and
// aligns each pair of sequences if they overlap. If they do not, both
// sequences are pasted together and a strech of ten dots is added at the
// juction of both the sequences.
//
// Parameters
//
// - iterator: is an iterator of paired sequences as produced by the method
// IBioSequenceBatch.PairWith
//
// - gap: the gap penality is expressed as a multiplicator factor of the cost
// of a mismatch between two bases having a quality score of 40.
//
// - delta: the extension in number of base pairs added on both sides of the
// overlap detected by the FAST algorithm before the optimal alignment.
//
// - minOverlap: the minimal length of the overlap to accept the alignment of
// the paired reads as correct. If the actual length is below this limit. The
// the alignment is discarded and both sequences are pasted.
//
// - withStats: indicates (true value) if the algorithm adds annotation to each
// sequence on the quality of the aligned overlap.
//
// Two extra interger parameters can be added during the call of the function.
// The first one indicates how many parallel workers run for aligning the sequences.
// The second allows too specify the size of the channel buffer.
//
// Returns
//
// The function returns an iterator over batches of obiseq.Biosequence object.
// each pair of processed sequences produces one sequence in the result iterator.
//
func IAssemblePESequencesBatch(iterator obiiter.IPairedBioSequenceBatch,
	gap float64, delta, minOverlap int,
	minIdentity float64,
	withStats bool, sizes ...int) obiiter.IBioSequenceBatch {

	nworkers := runtime.NumCPU() * 3 / 2
	buffsize := iterator.BufferSize()

	if len(sizes) > 0 {
		nworkers = sizes[0]
	}

	if len(sizes) > 1 {
		buffsize = sizes[1]
	}

	newIter := obiiter.MakeIBioSequenceBatch(buffsize)

	newIter.Add(nworkers)

	go func() {
		newIter.WaitAndClose()
		log.Printf("End of the sequence Pairing")
	}()

	bar := progressbar.NewOptions(
		-1,
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetWidth(15),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetDescription("[Sequence Pairing]"))

	f := func(iterator obiiter.IPairedBioSequenceBatch, wid int) {
		arena := obialign.MakePEAlignArena(150, 150)

		for iterator.Next() {
			batch := iterator.Get()
			cons := make(obiseq.BioSequenceSlice, len(batch.Forward()))
			processed := 0
			for i, A := range batch.Forward() {
				B := batch.Reverse()[i]
				cons[i] = AssemblePESequences(A, B, gap, delta, minOverlap, minIdentity, withStats, true, arena)
				if i%59 == 0 {
					bar.Add(59)
					processed += 59
				}
			}
			bar.Add(batch.Length() - processed)
			newIter.Push(obiiter.MakeBioSequenceBatch(
				batch.Order(),
				cons,
			))
		}
		newIter.Done()
	}

	log.Printf("Start of the sequence Pairing using %d workers\n", nworkers)

	for i := 0; i < nworkers-1; i++ {
		go f(iterator.Split(), i)
	}
	go f(iterator, nworkers-1)
	return newIter

}
