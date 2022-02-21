package obialign

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

var _FourBitsBaseCode = []byte{0b0000,
	// IUPAC nucleotide code	Base
	0b0001, // A	Adenine
	0b1110, // B	C or G or T
	0b0010, // C	Cytosine
	0b1101, // D	A or G or T
	0b0000, // E    not a nucleotide
	0b0000, // F    not a nucleotide
	0b0100, // G	Guanine
	0b1011, // H	A or C or T
	0b0000, // I    not a nucleotide
	0b0000, // J    not a nucleotide
	0b1100, // K	G or T
	0b0000, // L    not a nucleotide
	0b0011, // M	A or C
	0b1111, // N	any base
	0b0000, // O    not a nucleotide
	0b0000, // P    not a nucleotide
	0b0000, // Q    not a nucleotide
	0b0101, // R	A or G
	0b0110, // S	G or C
	0b1000, // T    Thymine
	0b1000, // U    Uracil
	0b0111, // V	A or C or G
	0b1001, // W	A or T
	0b0000, // X    not a nucleotide
	0b1010, // Y	C or T
	0b0000, // Z    not a nucleotide
	0b0000,
	0b0000,
	0b0000,
	0b0000,
	0b0000}

var _FourBitsBaseDecode = []byte{
	// 	0b0000 0b0001 0b0010 0b0011
	'.', 'a', 'c', 'm',
	// 	0b0100 0b0101 0b0110 0b0111
	'g', 'r', 's', 'v',
	// 	0b1000 0b1001 0b1010 0b1011
	't', 'w', 'y', 'h',
	// 	0b1100 0b1101 0b1110 0b1111
	'k', 'd', 'b', 'n',
}

// Encode4bits encodes each nucleotide of a sequence into a binary
// code where the four low weigth bit of a byte correspond respectively
// to the four nucleotides A, C, G, T. Simple bases A, C, G, T are therefore
// represented by a code with only a single bit on, when anbiguous symboles
// like R, D or N have the bits corresponding to each nucleotide represented
// by the ambiguity set to 1.
// A byte slice can be provided (buffer) to preveent allocation of a new
// memory chunk by th function.
func Encode4bits(seq *obiseq.BioSequence, buffer []byte) []byte {
	length := seq.Length()
	rawseq := seq.Sequence()

	if buffer == nil {
		buffer = make([]byte, 0, length)
	} else {
		buffer = buffer[:0]
	}

	var code byte

	for _, nuc := range rawseq {
		if nuc == '.' || nuc == '-' {
			code = 0
		} else {
			code = _FourBitsBaseCode[nuc&31]
		}
		buffer = append(buffer, code)
	}

	return buffer
}
