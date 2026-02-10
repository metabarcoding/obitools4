package obikmer

import (
	"bufio"
	"container/heap"
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
)

// KSP file magic bytes: "KSP\x01" (K-mer SPectrum v1)
var kspMagic = [4]byte{'K', 'S', 'P', 0x01}

// SpectrumEntry represents one entry in a k-mer frequency spectrum.
type SpectrumEntry struct {
	Frequency int    // how many times a k-mer was observed
	Count     uint64 // how many distinct k-mers have this frequency
}

// KmerSpectrum represents the frequency distribution of k-mers.
// Entries are sorted by Frequency in ascending order and only include
// non-zero counts.
type KmerSpectrum struct {
	Entries []SpectrumEntry
}

// MaxFrequency returns the highest frequency in the spectrum, or 0 if empty.
func (s *KmerSpectrum) MaxFrequency() int {
	if len(s.Entries) == 0 {
		return 0
	}
	return s.Entries[len(s.Entries)-1].Frequency
}

// ToMap converts a KmerSpectrum back to a map for easy lookup.
func (s *KmerSpectrum) ToMap() map[int]uint64 {
	m := make(map[int]uint64, len(s.Entries))
	for _, e := range s.Entries {
		m[e.Frequency] = e.Count
	}
	return m
}

// MapToSpectrum converts a map[int]uint64 to a sorted KmerSpectrum.
func MapToSpectrum(m map[int]uint64) *KmerSpectrum {
	entries := make([]SpectrumEntry, 0, len(m))
	for freq, count := range m {
		if count > 0 {
			entries = append(entries, SpectrumEntry{Frequency: freq, Count: count})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Frequency < entries[j].Frequency
	})
	return &KmerSpectrum{Entries: entries}
}

// MergeSpectraMaps adds all entries from b into a.
func MergeSpectraMaps(a, b map[int]uint64) {
	for freq, count := range b {
		a[freq] += count
	}
}

// WriteSpectrum writes a KmerSpectrum to a binary file.
//
// Format:
//
//	[magic: 4 bytes "KSP\x01"]
//	[n_entries: varint]
//	For each entry (sorted by frequency ascending):
//	  [frequency: varint]
//	  [count: varint]
func WriteSpectrum(path string, spectrum *KmerSpectrum) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create spectrum file: %w", err)
	}
	w := bufio.NewWriterSize(f, 65536)

	// Magic
	if _, err := w.Write(kspMagic[:]); err != nil {
		f.Close()
		return err
	}

	// Number of entries
	if _, err := EncodeVarint(w, uint64(len(spectrum.Entries))); err != nil {
		f.Close()
		return err
	}

	// Entries
	for _, e := range spectrum.Entries {
		if _, err := EncodeVarint(w, uint64(e.Frequency)); err != nil {
			f.Close()
			return err
		}
		if _, err := EncodeVarint(w, e.Count); err != nil {
			f.Close()
			return err
		}
	}

	if err := w.Flush(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// ReadSpectrum reads a KmerSpectrum from a binary file.
func ReadSpectrum(path string) (*KmerSpectrum, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := bufio.NewReaderSize(f, 65536)

	// Check magic
	var magic [4]byte
	if _, err := r.Read(magic[:]); err != nil {
		return nil, fmt.Errorf("read spectrum magic: %w", err)
	}
	if magic != kspMagic {
		return nil, fmt.Errorf("invalid spectrum file magic: %v", magic)
	}

	// Number of entries
	nEntries, err := DecodeVarint(r)
	if err != nil {
		return nil, fmt.Errorf("read spectrum entry count: %w", err)
	}

	entries := make([]SpectrumEntry, nEntries)
	for i := uint64(0); i < nEntries; i++ {
		freq, err := DecodeVarint(r)
		if err != nil {
			return nil, fmt.Errorf("read spectrum freq at entry %d: %w", i, err)
		}
		count, err := DecodeVarint(r)
		if err != nil {
			return nil, fmt.Errorf("read spectrum count at entry %d: %w", i, err)
		}
		entries[i] = SpectrumEntry{
			Frequency: int(freq),
			Count:     count,
		}
	}

	return &KmerSpectrum{Entries: entries}, nil
}

// KmerFreq associates a k-mer (encoded as uint64) with its observed frequency.
type KmerFreq struct {
	Kmer uint64
	Freq int
}

// kmerFreqHeap is a min-heap of KmerFreq ordered by Freq (lowest first).
// Used to maintain a top-N most frequent k-mers set.
type kmerFreqHeap []KmerFreq

func (h kmerFreqHeap) Len() int            { return len(h) }
func (h kmerFreqHeap) Less(i, j int) bool  { return h[i].Freq < h[j].Freq }
func (h kmerFreqHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *kmerFreqHeap) Push(x interface{}) { *h = append(*h, x.(KmerFreq)) }
func (h *kmerFreqHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// TopNKmers maintains a collection of the N most frequent k-mers
// using a min-heap. Thread-safe usage requires external synchronization.
type TopNKmers struct {
	n int
	h kmerFreqHeap
}

// NewTopNKmers creates a new top-N collector.
func NewTopNKmers(n int) *TopNKmers {
	return &TopNKmers{
		n: n,
		h: make(kmerFreqHeap, 0, n+1),
	}
}

// Add considers a k-mer with the given frequency for inclusion in the top-N.
func (t *TopNKmers) Add(kmer uint64, freq int) {
	if t.n <= 0 {
		return
	}
	if len(t.h) < t.n {
		heap.Push(&t.h, KmerFreq{Kmer: kmer, Freq: freq})
	} else if freq > t.h[0].Freq {
		t.h[0] = KmerFreq{Kmer: kmer, Freq: freq}
		heap.Fix(&t.h, 0)
	}
}

// Results returns the collected k-mers sorted by frequency descending.
func (t *TopNKmers) Results() []KmerFreq {
	result := make([]KmerFreq, len(t.h))
	copy(result, t.h)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Freq > result[j].Freq
	})
	return result
}

// MergeTopN merges another TopNKmers into this one.
func (t *TopNKmers) MergeTopN(other *TopNKmers) {
	if other == nil {
		return
	}
	for _, kf := range other.h {
		t.Add(kf.Kmer, kf.Freq)
	}
}

// WriteTopKmersCSV writes the top k-mers to a CSV file.
// Columns: sequence, frequency
func WriteTopKmersCSV(path string, topKmers []KmerFreq, k int) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create top-kmers file: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{"sequence", "frequency"}); err != nil {
		return err
	}

	buf := make([]byte, k)
	for _, kf := range topKmers {
		seq := DecodeKmer(kf.Kmer, k, buf)
		if err := w.Write([]string{string(seq), strconv.Itoa(kf.Freq)}); err != nil {
			return err
		}
	}

	return nil
}
