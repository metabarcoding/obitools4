package obikmer

import (
	"bufio"
	"encoding/binary"
	"os"
)

// SkmWriter writes super-kmers to a binary .skm file.
//
// Format per super-kmer:
//
//	[len: uint16 LE]          length of the super-kmer in bases
//	[data: ceil(len/4) bytes] sequence encoded 2 bits/base, packed
//
// Nucleotide encoding: A=00, C=01, G=10, T=11.
// The last byte is zero-padded on the low bits if len%4 != 0.
type SkmWriter struct {
	w    *bufio.Writer
	file *os.File
}

// NewSkmWriter creates a new SkmWriter writing to the given file path.
func NewSkmWriter(path string) (*SkmWriter, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return &SkmWriter{
		w:    bufio.NewWriterSize(f, 65536),
		file: f,
	}, nil
}

// Write encodes a SuperKmer to the .skm file.
// The sequence bytes are packed 2 bits per base.
func (sw *SkmWriter) Write(sk SuperKmer) error {
	seq := sk.Sequence
	seqLen := uint16(len(seq))

	// Write length
	var lenbuf [2]byte
	binary.LittleEndian.PutUint16(lenbuf[:], seqLen)
	if _, err := sw.w.Write(lenbuf[:]); err != nil {
		return err
	}

	// Encode and write packed sequence (2 bits/base)
	nBytes := (int(seqLen) + 3) / 4
	for i := 0; i < nBytes; i++ {
		var packed byte
		for j := 0; j < 4; j++ {
			pos := i*4 + j
			packed <<= 2
			if pos < int(seqLen) {
				packed |= __single_base_code__[seq[pos]&31]
			}
		}
		if err := sw.w.WriteByte(packed); err != nil {
			return err
		}
	}

	return nil
}

// Close flushes buffered data and closes the underlying file.
func (sw *SkmWriter) Close() error {
	if err := sw.w.Flush(); err != nil {
		sw.file.Close()
		return err
	}
	return sw.file.Close()
}
