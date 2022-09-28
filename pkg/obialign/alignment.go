// obialign : function for aligning two sequences
//
// The obialign package provides a set of functions
// foor aligning two objects of type obiseq.BioSequence.
package obialign

import (
	"math"
	"sync"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

// A pool of byte slices.
var _BuildAlignArenaPool = sync.Pool{
	New: func() interface{} {
		bs := make([]byte, 0, 300)
		return &bs
	},
}

// It takes two sequences, a path, a gap character, and two buffers, and it builds the alignment by
// walking the path and copying the sequences into the buffers
func _BuildAlignment(seqA, seqB []byte, path []int, gap byte, bufferA, bufferB *[]byte) {

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
func BuildAlignment(seqA, seqB *obiseq.BioSequence,
	path []int, gap byte) (*obiseq.BioSequence, *obiseq.BioSequence) {

	bufferSA := obiseq.GetSlice(seqA.Length())
	defer obiseq.RecycleSlice(&bufferSA)

	bufferSB := obiseq.GetSlice(seqB.Length())
	defer obiseq.RecycleSlice(&bufferSB)

	_BuildAlignment(seqA.Sequence(), seqB.Sequence(), path, gap,
		&bufferSA,
		&bufferSB)

	seqA = obiseq.NewBioSequence(seqA.Id(),
		bufferSA,
		seqA.Definition())

	seqB = obiseq.NewBioSequence(seqB.Id(),
		bufferSB,
		seqB.Definition())

	return seqA, seqB
}

// func _logSlice(x *[]byte) {
// 	l := len(*x)
// 	if l > 10 {
// 		l = 10
// 	}
// 	log.Printf("%v (%10s): slice=%p array=%p cap=%d len=%d\n", (*x)[:l], string((*x)[:l]), x, (*x), cap(*x), len(*x))
// }

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
func BuildQualityConsensus(seqA, seqB *obiseq.BioSequence, path []int) (*obiseq.BioSequence, int) {

	bufferSA := obiseq.GetSlice(seqA.Length())
	bufferSB := obiseq.GetSlice(seqB.Length())
	defer obiseq.RecycleSlice(&bufferSB)

	bufferQA := obiseq.GetSlice(seqA.Length())
	bufferQB := obiseq.GetSlice(seqB.Length())
	defer obiseq.RecycleSlice(&bufferQB)

	_BuildAlignment(seqA.Sequence(), seqB.Sequence(), path, ' ',
		&bufferSA, &bufferSB)

	// log.Printf("#1 %s--> la : %d,%p lb : %d,%p qa : %d,%p qb : %d,%p\n", stamp,
	// 	len(*bufferSA), bufferSA, len(*bufferSB), bufferSB,
	// 	len(*bufferQA), bufferQA, len(*bufferQB), bufferQB)

	_BuildAlignment(seqA.Qualities(), seqB.Qualities(), path, byte(0),
		&bufferQA, &bufferQB)

	// log.Printf("#2 %s--> la : %d,%p lb : %d,%p qa : %d,%p qb : %d,%p\n", stamp,
	// 	len(*bufferSA), bufferSA, len(*bufferSB), bufferSB,
	// 	len(*bufferQA), bufferQA, len(*bufferQB), bufferQB)
	// log.Printf("#3 %s--> la : %d lb : %d, qa : %d qb : %d\n", stamp, len(sA), len(sB), len(qsA), len(qsB))

	var qA, qB byte
	var qM, qm byte
	var i int

	match := 0

	for i, qA = range bufferQA {
		nA := bufferSA[i]
		nB := bufferSB[i]
		qB = bufferQB[i]

		if qA > qB {
			qM = qA
			qm = qB
		}
		if qB > qA {
			bufferSA[i] = bufferSB[i]
			qM = qB
			qm = qA
		}
		if qB == qA && nA != nB {
			nuc := _FourBitsBaseCode[nA&31] | _FourBitsBaseCode[nB&31]
			bufferSA[i] = _FourBitsBaseDecode[nuc]
		}

		q := qA + qB

		if qA > 0 && qB > 0 {
			if nA != nB {
				q = qM - byte(math.Log10(1-math.Pow(10, -float64(qm)/30))*10+0.5)
			}
			if nA == nB {
				match++
			}
		}

		if q > 90 {
			q = 90
		}

		bufferQA[i] = q
	}

	consSeq := obiseq.NewBioSequence(
		seqA.Id(),
		bufferSA,
		seqA.Definition(),
	)
	consSeq.SetQualities(bufferQA)

	return consSeq, match
}
