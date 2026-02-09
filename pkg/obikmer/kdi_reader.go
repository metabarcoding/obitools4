package obikmer

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// KdiReader reads k-mers from a .kdi file using streaming delta-varint decoding.
type KdiReader struct {
	r       *bufio.Reader
	file    *os.File
	count   uint64 // total number of k-mers
	read    uint64 // number of k-mers already consumed
	prev    uint64 // last decoded value
	started bool   // whether first value has been read
}

// NewKdiReader opens a .kdi file for streaming reading.
func NewKdiReader(path string) (*KdiReader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	r := bufio.NewReaderSize(f, 65536)

	// Read and verify magic
	var magic [4]byte
	if _, err := io.ReadFull(r, magic[:]); err != nil {
		f.Close()
		return nil, fmt.Errorf("kdi: read magic: %w", err)
	}
	if magic != kdiMagic {
		f.Close()
		return nil, fmt.Errorf("kdi: bad magic %v", magic)
	}

	// Read count
	var countBuf [8]byte
	if _, err := io.ReadFull(r, countBuf[:]); err != nil {
		f.Close()
		return nil, fmt.Errorf("kdi: read count: %w", err)
	}
	count := binary.LittleEndian.Uint64(countBuf[:])

	return &KdiReader{
		r:     r,
		file:  f,
		count: count,
	}, nil
}

// Next returns the next k-mer and true, or (0, false) when exhausted.
func (kr *KdiReader) Next() (uint64, bool) {
	if kr.read >= kr.count {
		return 0, false
	}

	if !kr.started {
		// Read first value as absolute uint64 LE
		var buf [8]byte
		if _, err := io.ReadFull(kr.r, buf[:]); err != nil {
			return 0, false
		}
		kr.prev = binary.LittleEndian.Uint64(buf[:])
		kr.started = true
		kr.read++
		return kr.prev, true
	}

	// Read delta varint
	delta, err := DecodeVarint(kr.r)
	if err != nil {
		return 0, false
	}
	kr.prev += delta
	kr.read++
	return kr.prev, true
}

// Count returns the total number of k-mers in this partition.
func (kr *KdiReader) Count() uint64 {
	return kr.count
}

// Remaining returns how many k-mers have not been read yet.
func (kr *KdiReader) Remaining() uint64 {
	return kr.count - kr.read
}

// Close closes the underlying file.
func (kr *KdiReader) Close() error {
	return kr.file.Close()
}
