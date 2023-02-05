package obikmer

import "git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"

var __single_base_code__ = []byte{0,
	//     A,  B,  C,  D,
	0, 0, 1, 0,
	//     E,  F,  G,  H,
	0, 0, 2, 0,
	//     I,  J,  K,  L,
	0, 0, 0, 0,
	//     M,  N,  O,  P,
	0, 0, 0, 0,
	//     Q,  R,  S,  T,
	0, 0, 0, 3,
	//     U,  V,  W,  X,
	3, 0, 0, 0,
	//     Y,  Z,  .,  .,
	0, 0, 0, 0,
	0, 0, 0,
}

// Encode4mer transforms an obiseq.BioSequence into a sequence
// of kmer of length 4. Each letter of the sequence not belonging
// A, C, G, T, U are considered as a A. The kmer is encoded as a byte
// value ranging from 0 to 255. Each nucleotite is represented by
// two bits. The values 0, 1, 2, 3 correspond respectively to A, C, G,
// and T. U is encoded by 3 like T. Therefore AAAA has the code 0 and
// TTTT the code 255 when ACGT is encoded by 00011011 in binary, 0x1B
// in hexadecimal and 27 in decimal. If the buffer parameter is not nil
// the slice is used to store the result, overwise a new slice is
// created.
func Encode4mer(seq *obiseq.BioSequence, buffer *[]byte) []byte {
	slength := seq.Len()
	length := slength - 3
	rawseq := seq.Sequence()

	if length < 0 {
		return nil
	}

	if buffer == nil {
		b := make([]byte, 0, length)
		buffer = &b
	} else {
		*buffer = (*buffer)[:0]
	}

	var code byte
	i := 0
	code = 0
	for ; i < 4; i++ {
		code <<= 2
		code += __single_base_code__[rawseq[i]&31]
	}

	*buffer = append((*buffer), code)

	for ; i < slength; i++ {
		code <<= 2
		code |= __single_base_code__[rawseq[i]&31]
		*buffer = append((*buffer), code)
	}

	return *buffer
}

// Index4mer returns an index where the occurrence position of every fourmer is
// stored. The index is returned as an array of slices of integer.  The first
// dimention corresponds to the code of the 4mer, the second
func Index4mer(seq *obiseq.BioSequence, index *[][]int, buffer *[]byte) [][]int {

	iternal_buffer := Encode4mer(seq, buffer)

	if index == nil || cap(*index) < 256 {
		// A new index is created
		i := make([][]int, 256)
		if index == nil {
			index = &i
		} else {
			*index = i
		}
	}

	// Every cells of the index is emptied
	for i := 0; i < 256; i++ {
		(*index)[i] = (*index)[i][:0]
	}

	for pos, code := range iternal_buffer {
		(*index)[code] = append((*index)[code], pos)
	}

	return *index
}

// FastShiftFourMer runs a Fast algorithm (similar to the one used in FASTA) to compare two sequences.
// The returned values are two integer values. The shift between both the sequences and the count of
// matching 4mer when this shift is applied between both the sequences.
func FastShiftFourMer(index [][]int, seq *obiseq.BioSequence, buffer *[]byte) (int, int) {

	iternal_buffer := Encode4mer(seq, buffer)

	shifts := make(map[int]int, 3*seq.Len())

	for pos, code := range iternal_buffer {
		for _, refpos := range index[code] {
			shift := refpos - pos
			count, ok := shifts[shift]
			if ok {
				shifts[shift] = count + 1
			} else {
				shifts[shift] = 1
			}
		}
	}

	maxshift := 0
	maxcount := -1

	for shift, count := range shifts {
		if count > maxcount {
			maxshift = shift
			maxcount = count
		} else {
			if count == maxcount && shift < maxshift {
				maxshift = shift
			}
		}
	}

	return maxshift, maxcount
}
