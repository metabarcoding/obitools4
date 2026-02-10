package obikmer

import (
	"bufio"
	"encoding/binary"
	"os"
)

// KDI file magic bytes: "KDI\x01"
var kdiMagic = [4]byte{'K', 'D', 'I', 0x01}

// kdiHeaderSize is the size of the KDI header: magic(4) + count(8) = 12 bytes.
const kdiHeaderSize = 12

// KdiWriter writes a sorted sequence of uint64 k-mers to a .kdi file
// using delta-varint encoding.
//
// Format:
//
//	[magic: 4 bytes "KDI\x01"]
//	[count: uint64 LE]        number of k-mers
//	[first: uint64 LE]        first k-mer (absolute value)
//	[delta_1: varint]          arr[1] - arr[0]
//	[delta_2: varint]          arr[2] - arr[1]
//	...
//
// The caller must write k-mers in strictly increasing order.
//
// On Close(), a companion .kdx sparse index file is written alongside
// the .kdi file for fast random access.
type KdiWriter struct {
	w            *bufio.Writer
	file         *os.File
	count        uint64
	prev         uint64
	first        bool
	path         string
	bytesWritten uint64     // bytes written after header (data section offset)
	indexEntries []kdxEntry // sparse index entries collected during writes
}

// NewKdiWriter creates a new KdiWriter writing to the given file path.
// The header (magic + count placeholder) is written immediately.
// Count is patched on Close().
func NewKdiWriter(path string) (*KdiWriter, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	w := bufio.NewWriterSize(f, 65536)

	// Write magic
	if _, err := w.Write(kdiMagic[:]); err != nil {
		f.Close()
		return nil, err
	}
	// Write placeholder for count (will be patched on Close)
	var countBuf [8]byte
	if _, err := w.Write(countBuf[:]); err != nil {
		f.Close()
		return nil, err
	}

	return &KdiWriter{
		w:            w,
		file:         f,
		first:        true,
		path:         path,
		bytesWritten: 0,
		indexEntries: make([]kdxEntry, 0, 256),
	}, nil
}

// Write adds a k-mer to the file. K-mers must be written in strictly
// increasing order.
func (kw *KdiWriter) Write(kmer uint64) error {
	if kw.first {
		// Write first value as absolute uint64 LE
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], kmer)
		if _, err := kw.w.Write(buf[:]); err != nil {
			return err
		}
		kw.bytesWritten += 8
		kw.prev = kmer
		kw.first = false
	} else {
		delta := kmer - kw.prev
		n, err := EncodeVarint(kw.w, delta)
		if err != nil {
			return err
		}
		kw.bytesWritten += uint64(n)
		kw.prev = kmer
	}
	kw.count++

	// Record sparse index entry every defaultKdxStride k-mers.
	// The offset recorded is AFTER writing this k-mer, so it points to
	// where the next k-mer's data will start. SeekTo uses this: it seeks
	// to the recorded offset, sets prev = indexedKmer, and Next() reads
	// the delta of the following k-mer.
	if kw.count%defaultKdxStride == 0 {
		kw.indexEntries = append(kw.indexEntries, kdxEntry{
			kmer:   kmer,
			offset: kdiHeaderSize + kw.bytesWritten,
		})
	}

	return nil
}

// Count returns the number of k-mers written so far.
func (kw *KdiWriter) Count() uint64 {
	return kw.count
}

// Close flushes buffered data, patches the count in the header,
// writes the companion .kdx index file, and closes the file.
func (kw *KdiWriter) Close() error {
	if err := kw.w.Flush(); err != nil {
		kw.file.Close()
		return err
	}

	// Patch count at offset 4 (after magic)
	if _, err := kw.file.Seek(4, 0); err != nil {
		kw.file.Close()
		return err
	}
	var countBuf [8]byte
	binary.LittleEndian.PutUint64(countBuf[:], kw.count)
	if _, err := kw.file.Write(countBuf[:]); err != nil {
		kw.file.Close()
		return err
	}

	if err := kw.file.Close(); err != nil {
		return err
	}

	// Write .kdx index file if there are entries to index
	if len(kw.indexEntries) > 0 {
		kdxPath := KdxPathForKdi(kw.path)
		if err := WriteKdxIndex(kdxPath, defaultKdxStride, kw.indexEntries); err != nil {
			return err
		}
	}

	return nil
}
