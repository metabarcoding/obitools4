package obikmer

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// KDX file magic bytes: "KDX\x01"
var kdxMagic = [4]byte{'K', 'D', 'X', 0x01}

// defaultKdxStride is the number of k-mers between consecutive index entries.
const defaultKdxStride = 4096

// kdxEntry is a single entry in the sparse index: the absolute k-mer value
// and the byte offset in the corresponding .kdi file where that k-mer is stored.
type kdxEntry struct {
	kmer   uint64
	offset uint64 // absolute byte offset in .kdi file
}

// KdxIndex is a sparse, in-memory index for a .kdi file.
// It stores one entry every `stride` k-mers, enabling O(log N / stride)
// binary search followed by at most `stride` linear scan steps.
type KdxIndex struct {
	stride  int
	entries []kdxEntry
}

// LoadKdxIndex reads a .kdx file into memory.
// Returns (nil, nil) if the file does not exist (graceful degradation).
func LoadKdxIndex(path string) (*KdxIndex, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	// Read magic
	var magic [4]byte
	if _, err := io.ReadFull(f, magic[:]); err != nil {
		return nil, fmt.Errorf("kdx: read magic: %w", err)
	}
	if magic != kdxMagic {
		return nil, fmt.Errorf("kdx: bad magic %v", magic)
	}

	// Read stride (uint32 LE)
	var buf4 [4]byte
	if _, err := io.ReadFull(f, buf4[:]); err != nil {
		return nil, fmt.Errorf("kdx: read stride: %w", err)
	}
	stride := int(binary.LittleEndian.Uint32(buf4[:]))

	// Read count (uint32 LE)
	if _, err := io.ReadFull(f, buf4[:]); err != nil {
		return nil, fmt.Errorf("kdx: read count: %w", err)
	}
	count := int(binary.LittleEndian.Uint32(buf4[:]))

	// Read entries
	entries := make([]kdxEntry, count)
	var buf16 [16]byte
	for i := 0; i < count; i++ {
		if _, err := io.ReadFull(f, buf16[:]); err != nil {
			return nil, fmt.Errorf("kdx: read entry %d: %w", i, err)
		}
		entries[i] = kdxEntry{
			kmer:   binary.LittleEndian.Uint64(buf16[0:8]),
			offset: binary.LittleEndian.Uint64(buf16[8:16]),
		}
	}

	return &KdxIndex{
		stride:  stride,
		entries: entries,
	}, nil
}

// FindOffset locates the best starting point in the .kdi file to scan for
// the target k-mer. It returns:
//   - offset: the byte offset in the .kdi file to seek to (positioned after
//     the indexed k-mer, ready to read the next delta)
//   - skipCount: the number of k-mers already consumed at that offset
//     (to set the reader's internal counter)
//   - ok: true if the index provides a useful starting point
//
// Index entries are recorded at k-mer count positions stride, 2*stride, etc.
// Entry i corresponds to the k-mer written at count = (i+1)*stride.
func (idx *KdxIndex) FindOffset(target uint64) (offset uint64, skipCount uint64, ok bool) {
	if idx == nil || len(idx.entries) == 0 {
		return 0, 0, false
	}

	// Binary search: find the largest entry with kmer <= target
	i := sort.Search(len(idx.entries), func(i int) bool {
		return idx.entries[i].kmer > target
	})
	// i is the first entry with kmer > target, so i-1 is the last with kmer <= target
	if i == 0 {
		// Target is before the first index entry.
		// No useful jump point — caller should scan from the beginning.
		return 0, 0, false
	}

	i-- // largest entry with kmer <= target
	// Entry i was recorded after writing k-mer at count = (i+1)*stride
	skipCount = uint64(i+1) * uint64(idx.stride)
	return idx.entries[i].offset, skipCount, true
}

// Stride returns the stride of this index.
func (idx *KdxIndex) Stride() int {
	return idx.stride
}

// Len returns the number of entries in this index.
func (idx *KdxIndex) Len() int {
	return len(idx.entries)
}

// WriteKdxIndex writes a .kdx file from a slice of entries.
func WriteKdxIndex(path string, stride int, entries []kdxEntry) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Magic
	if _, err := f.Write(kdxMagic[:]); err != nil {
		return err
	}

	// Stride (uint32 LE)
	var buf4 [4]byte
	binary.LittleEndian.PutUint32(buf4[:], uint32(stride))
	if _, err := f.Write(buf4[:]); err != nil {
		return err
	}

	// Count (uint32 LE)
	binary.LittleEndian.PutUint32(buf4[:], uint32(len(entries)))
	if _, err := f.Write(buf4[:]); err != nil {
		return err
	}

	// Entries
	var buf16 [16]byte
	for _, e := range entries {
		binary.LittleEndian.PutUint64(buf16[0:8], e.kmer)
		binary.LittleEndian.PutUint64(buf16[8:16], e.offset)
		if _, err := f.Write(buf16[:]); err != nil {
			return err
		}
	}

	return nil
}

// KdxPathForKdi returns the .kdx path corresponding to a .kdi path.
func KdxPathForKdi(kdiPath string) string {
	return strings.TrimSuffix(kdiPath, ".kdi") + ".kdx"
}
