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
	count   uint64    // total number of k-mers
	read    uint64    // number of k-mers already consumed
	prev    uint64    // last decoded value
	started bool      // whether first value has been read
	index   *KdxIndex // optional sparse index for seeking
}

// NewKdiReader opens a .kdi file for streaming reading (no index).
func NewKdiReader(path string) (*KdiReader, error) {
	return openKdiReader(path, nil)
}

// NewKdiIndexedReader opens a .kdi file with its companion .kdx index
// loaded for fast seeking. If the .kdx file does not exist, it gracefully
// falls back to sequential reading.
func NewKdiIndexedReader(path string) (*KdiReader, error) {
	kdxPath := KdxPathForKdi(path)
	idx, err := LoadKdxIndex(kdxPath)
	if err != nil {
		// Index load failed — fall back to non-indexed
		return openKdiReader(path, nil)
	}
	// idx may be nil if file does not exist — that's fine
	return openKdiReader(path, idx)
}

func openKdiReader(path string, idx *KdxIndex) (*KdiReader, error) {
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
		index: idx,
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

// SeekTo positions the reader near the target k-mer using the sparse .kdx index.
// After SeekTo, the reader is positioned so that the next call to Next()
// returns the k-mer immediately after the indexed entry at or before target.
//
// If the reader has no index, or the target is before the current position,
// SeekTo does nothing (linear scan continues from current position).
func (kr *KdiReader) SeekTo(target uint64) error {
	if kr.index == nil {
		return nil
	}

	// If we've already passed the target, we can't seek backwards
	if kr.started && kr.prev >= target {
		return nil
	}

	offset, skipCount, ok := kr.index.FindOffset(target)
	if !ok {
		return nil
	}

	// skipCount is the number of k-mers consumed at the indexed position.
	// The index was recorded AFTER writing the k-mer at position skipCount-1
	// (since count%stride==0 after incrementing count). So the actual number
	// of k-mers consumed is skipCount (the entry's kmer is the last one
	// before the offset).

	// Only seek if it would skip significant work
	if kr.started && skipCount <= kr.read {
		return nil
	}

	// The index entry stores (kmer_value, byte_offset_after_that_kmer).
	// skipCount = (entryIdx+1)*stride, so entryIdx = skipCount/stride - 1
	// We seek to that offset, set prev = indexedKmer, and the next Next()
	// call will read the delta-varint of the following k-mer.
	entryIdx := int(skipCount)/kr.index.stride - 1
	if entryIdx < 0 || entryIdx >= len(kr.index.entries) {
		return nil
	}
	indexedKmer := kr.index.entries[entryIdx].kmer

	if _, err := kr.file.Seek(int64(offset), io.SeekStart); err != nil {
		return fmt.Errorf("kdi: seek: %w", err)
	}
	kr.r.Reset(kr.file)

	kr.prev = indexedKmer
	kr.started = true
	kr.read = skipCount

	return nil
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
