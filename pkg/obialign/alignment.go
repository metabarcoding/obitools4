package obialign

import (
	"math"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

type _BuildAlignArena struct {
	bufferA []byte
	bufferB []byte
}

// BuildAlignArena defines memory arena usable by the
// BuildAlignment function. The same arena can be reused
// from alignment to alignment to limit memory allocation
// and desallocation process.
type BuildAlignArena struct {
	pointer *_BuildAlignArena
}

// NilBuildAlignArena is the nil instance of the BuildAlignArena
// type.
var NilBuildAlignArena = BuildAlignArena{nil}

// MakeBuildAlignArena makes a new arena for aligning two sequences
// of maximum length indicated by lseqA and lseqB.
func MakeBuildAlignArena(lseqA, lseqB int) BuildAlignArena {
	a := _BuildAlignArena{
		bufferA: make([]byte, lseqA+lseqB),
		bufferB: make([]byte, lseqA+lseqB),
	}

	return BuildAlignArena{&a}
}

func _BuildAlignment(seqA, seqB []byte, path []int, gap byte,
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

	posA := 0
	posB := 0
	for i := 0; i < lp; i++ {
		step := path[i]
		if step < 0 {
			*bufferA = append(*bufferA, seqA[posA:(posA-step)]...)
			for j := 0; j < -step; j++ {
				*bufferB = append(*bufferB, gap)
			}
			posA -= step
		}
		if step > 0 {
			*bufferB = append(*bufferB, seqB[posB:(posB+step)]...)
			for j := 0; j < step; j++ {
				*bufferA = append(*bufferA, gap)
			}
			posB += step
		}

		i++
		step = path[i]
		if step > 0 {
			*bufferA = append(*bufferA, seqA[posA:(posA+step)]...)
			*bufferB = append(*bufferB, seqB[posB:(posB+step)]...)
			posA += step
			posB += step
		}
	}

	return *bufferA, *bufferB
}

// BuildAlignment builds the aligned sequences from an alignemnt path
// returned by one of the alignment procedure.
// The user has to provide both sequences (seqA and seqB), the alignment
// path (path), the symbole used to materialiaze gaps (gap) which is
// usually the dash '-', and a BuildAlignArena (arena). It is always possible
// to provide the NilBuildAlignArena instance for this last parameter.
// In that case an arena will be allocated by the function but, it will not
// be reusable for other alignments and desallocated at the BuildAlignment
// return.
func BuildAlignment(seqA, seqB obiseq.BioSequence,
	path []int, gap byte, arena BuildAlignArena) (obiseq.BioSequence, obiseq.BioSequence) {

	if arena.pointer == nil {
		arena = MakeBuildAlignArena(seqA.Length(), seqB.Length())
	}

	A, B := _BuildAlignment(seqA.Sequence(), seqB.Sequence(), path, gap,
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

// BuildQualityConsensus builds the consensus sequences corresponding to an
// alignement between two sequences.
// The consensus is built from an alignemnt path returned by one of the
// alignment procedure and the quality score associated to the sequence.
// In case of mismatches the nucleotide with the best score is conserved
// in the consensus. In case of score equality, an IUPAC symbol correesponding
// to the ambiguity is used.
// The user has to provide both sequences (seqA and seqB), the alignment
// path (path), and two BuildAlignArena (arena1 and arena2). It is always possible
// to provide the NilBuildAlignArena instance for these two last parameters.
// In that case arenas will be allocated by the function but, they will not
// be reusable for other alignments and desallocated at the BuildQualityConsensus
// return.
func BuildQualityConsensus(seqA, seqB obiseq.BioSequence, path []int,
	arena1, arena2 BuildAlignArena) (obiseq.BioSequence, int) {

	if arena1.pointer == nil {
		arena1 = MakeBuildAlignArena(seqA.Length(), seqB.Length())
	}
	if arena2.pointer == nil {
		arena2 = MakeBuildAlignArena(seqA.Length(), seqB.Length())
	}

	sA, sB := _BuildAlignment(seqA.Sequence(), seqB.Sequence(), path, ' ',
		&arena1.pointer.bufferA,
		&arena1.pointer.bufferB)

	qsA, qsB := _BuildAlignment(seqA.Qualities(), seqB.Qualities(), path, byte(0),
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
			nuc := _FourBitsBaseCode[sA[i]&31] | _FourBitsBaseCode[sB[i]&31]
			consensus = append(consensus, _FourBitsBaseDecode[nuc])
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
