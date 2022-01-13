package obialign

import (
	"math"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

type __build_align_arena__ struct {
	bufferA []byte
	bufferB []byte
}

type BuildAlignArena struct {
	pointer *__build_align_arena__
}

var NilBuildAlignArena = BuildAlignArena{nil}

func MakeBuildAlignArena(lseqA, lseqB int) BuildAlignArena {
	a := __build_align_arena__{
		bufferA: make([]byte, lseqA+lseqB),
		bufferB: make([]byte, lseqA+lseqB),
	}

	return BuildAlignArena{&a}
}

func __build_alignment__(seqA, seqB []byte, path []int, gap byte,
	bufferA, bufferB *[]byte) ([]byte, []byte) {

	if bufferA == nil {
		b := make([]byte, 0, len(seqA)+len(seqB))
		bufferA = &b
	}

	if bufferB == nil {
		b := make([]byte, 0, len(seqA)+len(seqB))
		bufferB = &b
	}

	*bufferA = (*bufferA)[:0]
	*bufferB = (*bufferB)[:0]

	lp := len(path)

	pos_a := 0
	pos_b := 0
	for i := 0; i < lp; i++ {
		step := path[i]
		if step < 0 {
			*bufferA = append(*bufferA, seqA[pos_a:(pos_a-step)]...)
			for j := 0; j < -step; j++ {
				*bufferB = append(*bufferB, gap)
			}
			pos_a -= step
		}
		if step > 0 {
			*bufferB = append(*bufferB, seqB[pos_b:(pos_b+step)]...)
			for j := 0; j < step; j++ {
				*bufferA = append(*bufferA, gap)
			}
			pos_b += step
		}

		i++
		step = path[i]
		if step > 0 {
			*bufferA = append(*bufferA, seqA[pos_a:(pos_a+step)]...)
			*bufferB = append(*bufferB, seqB[pos_b:(pos_b+step)]...)
			pos_a += step
			pos_b += step
		}
	}

	return *bufferA, *bufferB
}

func BuildAlignment(seqA, seqB obiseq.BioSequence,
	path []int, gap byte, arena BuildAlignArena) (obiseq.BioSequence, obiseq.BioSequence) {

	if arena.pointer == nil {
		arena = MakeBuildAlignArena(seqA.Length(), seqB.Length())
	}

	A, B := __build_alignment__(seqA.Sequence(), seqB.Sequence(), path, gap,
		&arena.pointer.bufferA,
		&arena.pointer.bufferB)

	seqA = obiseq.MakeBioSequence(seqA.Id(),
		A,
		seqA.Definition())

	seqB = obiseq.MakeBioSequence(seqB.Id(),
		B,
		seqB.Definition())

	return seqA, seqB
}

func BuildQualityConsensus(seqA, seqB obiseq.BioSequence, path []int,
	arena1, arena2 BuildAlignArena) (obiseq.BioSequence, int) {

	if arena1.pointer == nil {
		arena1 = MakeBuildAlignArena(seqA.Length(), seqB.Length())
	}
	if arena2.pointer == nil {
		arena2 = MakeBuildAlignArena(seqA.Length(), seqB.Length())
	}

	sA, sB := __build_alignment__(seqA.Sequence(), seqB.Sequence(), path, ' ',
		&arena1.pointer.bufferA,
		&arena1.pointer.bufferB)

	qsA, qsB := __build_alignment__(seqA.Qualities(), seqB.Qualities(), path, byte(0),
		&arena2.pointer.bufferA,
		&arena2.pointer.bufferB)

	consensus := make([]byte, 0, len(sA))
	qualities := make([]byte, 0, len(sA))

	var qA, qB byte
	var qM, qm byte
	var i int

	match := 0

	for i, qA = range qsA {
		qB = qsB[i]

		if qA > qB {
			consensus = append(consensus, sA[i])
			qM = qA
			qm = qB
		}
		if qB > qA {
			consensus = append(consensus, sB[i])
			qM = qB
			qm = qA
		}
		if qB == qA {
			nuc := __four_bits_base_code__[sA[i]&31] | __four_bits_base_code__[sB[i]&31]
			consensus = append(consensus, __four_bits_base_decode__[nuc])
		}

		q := qA + qB

		if qA > 0 && qB > 0 {
			if sA[i] != sB[i] {
				q = qM - byte(math.Log10(1-math.Pow(10, -float64(qm)/30))*10+0.5)
			}
			if sA[i] == sB[i] {
				match++
			}
		}

		if q > 90 {
			q = 90
		}
		qualities = append(qualities, q)
	}

	seq := obiseq.MakeBioSequence(seqA.Id(), consensus, seqA.Definition())
	seq.SetQualities(qualities)

	return seq, match
}
