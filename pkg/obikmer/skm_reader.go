package obikmer

import (
	"bufio"
	"encoding/binary"
	"io"
	"os"
)

// decode2bit maps 2-bit codes back to nucleotide bytes.
var decode2bit = [4]byte{'a', 'c', 'g', 't'}

// SkmReader reads super-kmers from a binary .skm file.
type SkmReader struct {
	r    *bufio.Reader
	file *os.File
}

// NewSkmReader opens a .skm file for reading.
func NewSkmReader(path string) (*SkmReader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &SkmReader{
		r:    bufio.NewReaderSize(f, 65536),
		file: f,
	}, nil
}

// Next reads the next super-kmer from the file.
// Returns the SuperKmer and true, or a zero SuperKmer and false at EOF.
func (sr *SkmReader) Next() (SuperKmer, bool) {
	// Read length
	var lenbuf [2]byte
	if _, err := io.ReadFull(sr.r, lenbuf[:]); err != nil {
		return SuperKmer{}, false
	}
	seqLen := int(binary.LittleEndian.Uint16(lenbuf[:]))

	// Read packed bytes
	nBytes := (seqLen + 3) / 4
	packed := make([]byte, nBytes)
	if _, err := io.ReadFull(sr.r, packed); err != nil {
		return SuperKmer{}, false
	}

	// Decode to nucleotide bytes
	seq := make([]byte, seqLen)
	for i := 0; i < seqLen; i++ {
		byteIdx := i / 4
		bitPos := uint(6 - (i%4)*2)
		code := (packed[byteIdx] >> bitPos) & 0x03
		seq[i] = decode2bit[code]
	}

	return SuperKmer{
		Sequence: seq,
		Start:    0,
		End:      seqLen,
	}, true
}

// Close closes the underlying file.
func (sr *SkmReader) Close() error {
	return sr.file.Close()
}
