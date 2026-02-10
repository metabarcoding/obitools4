package obikmer

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestKdiRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.kdi")

	// Sorted k-mer values
	kmers := []uint64{10, 20, 30, 100, 200, 500, 10000, 1 << 40, 1<<62 - 1}

	w, err := NewKdiWriter(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range kmers {
		if err := w.Write(v); err != nil {
			t.Fatal(err)
		}
	}
	if w.Count() != uint64(len(kmers)) {
		t.Fatalf("writer count: got %d, want %d", w.Count(), len(kmers))
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	// Read back
	r, err := NewKdiReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	if r.Count() != uint64(len(kmers)) {
		t.Fatalf("reader count: got %d, want %d", r.Count(), len(kmers))
	}

	for i, expected := range kmers {
		got, ok := r.Next()
		if !ok {
			t.Fatalf("unexpected EOF at index %d", i)
		}
		if got != expected {
			t.Fatalf("kmer %d: got %d, want %d", i, got, expected)
		}
	}

	_, ok := r.Next()
	if ok {
		t.Fatal("expected EOF after all k-mers")
	}
}

func TestKdiEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.kdi")

	w, err := NewKdiWriter(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	r, err := NewKdiReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	if r.Count() != 0 {
		t.Fatalf("expected count 0, got %d", r.Count())
	}

	_, ok := r.Next()
	if ok {
		t.Fatal("expected no k-mers in empty file")
	}
}

func TestKdiSingleValue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "single.kdi")

	w, err := NewKdiWriter(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Write(42); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	r, err := NewKdiReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	if r.Count() != 1 {
		t.Fatalf("expected count 1, got %d", r.Count())
	}

	v, ok := r.Next()
	if !ok {
		t.Fatal("expected one k-mer")
	}
	if v != 42 {
		t.Fatalf("got %d, want 42", v)
	}
}

func TestKdiFileSize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "size.kdi")

	// Write: magic(4) + count(8) + first(8) = 20 bytes
	w, err := NewKdiWriter(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Write(0); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	// magic(4) + count(8) + first(8) = 20
	if info.Size() != 20 {
		t.Fatalf("file size: got %d, want 20", info.Size())
	}
}

func TestKdiDeltaCompression(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "delta.kdi")

	// Dense consecutive values should compress well
	n := 10000
	kmers := make([]uint64, n)
	for i := range kmers {
		kmers[i] = uint64(i * 2) // even numbers
	}

	w, err := NewKdiWriter(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range kmers {
		if err := w.Write(v); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	// Each delta is 2, encoded as 1 byte varint
	// Total: magic(4) + count(8) + first(8) + (n-1)*1 = 20 + 9999 bytes
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	expected := int64(20 + n - 1)
	if info.Size() != expected {
		t.Fatalf("file size: got %d, want %d", info.Size(), expected)
	}

	// Verify round-trip
	r, err := NewKdiReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	for i, expected := range kmers {
		got, ok := r.Next()
		if !ok {
			t.Fatalf("unexpected EOF at index %d", i)
		}
		if got != expected {
			t.Fatalf("kmer %d: got %d, want %d", i, got, expected)
		}
	}
}

func TestKdiFromRealKmers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "real.kdi")

	// Extract k-mers from a sequence, sort, dedup, write to KDI
	seq := []byte("ACGTACGTACGTACGTACGTACGTACGTACGTACGTACGT")
	k := 15

	var kmers []uint64
	for kmer := range IterCanonicalKmers(seq, k) {
		kmers = append(kmers, kmer)
	}
	sort.Slice(kmers, func(i, j int) bool { return kmers[i] < kmers[j] })
	// Dedup
	deduped := kmers[:0]
	for i, v := range kmers {
		if i == 0 || v != kmers[i-1] {
			deduped = append(deduped, v)
		}
	}

	w, err := NewKdiWriter(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range deduped {
		if err := w.Write(v); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	// Read back and verify
	r, err := NewKdiReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	if r.Count() != uint64(len(deduped)) {
		t.Fatalf("count: got %d, want %d", r.Count(), len(deduped))
	}

	for i, expected := range deduped {
		got, ok := r.Next()
		if !ok {
			t.Fatalf("unexpected EOF at index %d", i)
		}
		if got != expected {
			t.Fatalf("kmer %d: got %d, want %d", i, got, expected)
		}
	}
}
