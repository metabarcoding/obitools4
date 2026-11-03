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
	w       *bufio.Writer
	file    *os.File
	scratch []byte
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

// appendPacked appends seq to dst packed 2 bits per base (A=00, C=01, G=10,
// T=11), four bases per byte, the last byte zero-padded on the low bits.
func appendPacked(dst, seq []byte) []byte {
	for i := 0; i < len(seq); i += 4 {
		var packed byte
		for j := 0; j < 4; j++ {
			packed <<= 2
			if i+j < len(seq) {
				packed |= __single_base_code__[seq[i+j]&31]
			}
		}
		dst = append(dst, packed)
	}
	return dst
}

// Write encodes a SuperKmer to the .skm file.
// The sequence bytes are packed 2 bits per base.
func (sw *SkmWriter) Write(sk SuperKmer) error {
	// Write length
	var lenbuf [2]byte
	binary.LittleEndian.PutUint16(lenbuf[:], uint16(len(sk.Sequence)))
	if _, err := sw.w.Write(lenbuf[:]); err != nil {
		return err
	}

	// Encode and write packed sequence (2 bits/base)
	sw.scratch = appendPacked(sw.scratch[:0], sk.Sequence)
	_, err := sw.w.Write(sw.scratch)
	return err
}

// Close flushes buffered data and closes the underlying file.
func (sw *SkmWriter) Close() error {
	if err := sw.w.Flush(); err != nil {
		sw.file.Close()
		return err
	}
	return sw.file.Close()
}
