package obialign

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

var __four_bits_base_code__ = []byte{0b0000,
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

var __four_bits_base_decode__ = []byte{
	// 	0b0000 0b0001 0b0010 0b0011
	'.', 'a', 'c', 'm',
	// 	0b0100 0b0101 0b0110 0b0111
	'g', 'r', 's', 'v',
	// 	0b1000 0b1001 0b1010 0b1011
	't', 'w', 'y', 'h',
	// 	0b1100 0b1101 0b1110 0b1111
	'k', 'd', 'b', 'n',
}

func Encode4bits(seq obiseq.BioSequence, buffer []byte) []byte {
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
			code = __four_bits_base_code__[nuc&31]
		}
		buffer = append(buffer, code)
	}

	return buffer
}
