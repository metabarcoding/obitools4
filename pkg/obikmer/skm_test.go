package obikmer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSkmRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.skm")

	// Create super-kmers from a known sequence
	seq := []byte("ACGTACGTACGTACGTACGTACGTACGTACGTACGTACGT")
	k := 21
	m := 9
	superKmers := ExtractSuperKmers(seq, k, m, nil)
	if len(superKmers) == 0 {
		t.Fatal("no super-kmers extracted")
	}

	// Write
	w, err := NewSkmWriter(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, sk := range superKmers {
		if err := w.Write(sk); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	// Read back
	r, err := NewSkmReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	idx := 0
	for {
		sk, ok := r.Next()
		if !ok {
			break
		}
		if idx >= len(superKmers) {
			t.Fatal("read more super-kmers than written")
		}
		expected := superKmers[idx]
		if len(sk.Sequence) != len(expected.Sequence) {
			t.Fatalf("super-kmer %d: length mismatch: got %d, want %d",
				idx, len(sk.Sequence), len(expected.Sequence))
		}
		// Compare nucleotide-by-nucleotide (case insensitive since decode produces lowercase)
		for j := range sk.Sequence {
			got := sk.Sequence[j] | 0x20
			want := expected.Sequence[j] | 0x20
			if got != want {
				t.Fatalf("super-kmer %d pos %d: got %c, want %c", idx, j, got, want)
			}
		}
		idx++
	}
	if idx != len(superKmers) {
		t.Fatalf("read %d super-kmers, want %d", idx, len(superKmers))
	}
}

func TestSkmEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.skm")

	// Write nothing
	w, err := NewSkmWriter(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	// Read back
	r, err := NewSkmReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	_, ok := r.Next()
	if ok {
		t.Fatal("expected no super-kmers in empty file")
	}
}

func TestSkmSingleBase(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "single.skm")

	// Test with sequences of various lengths to check padding
	sequences := [][]byte{
		[]byte("A"),
		[]byte("AC"),
		[]byte("ACG"),
		[]byte("ACGT"),
		[]byte("ACGTA"),
	}

	w, err := NewSkmWriter(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, seq := range sequences {
		sk := SuperKmer{Sequence: seq}
		if err := w.Write(sk); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	r, err := NewSkmReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	for i, expected := range sequences {
		sk, ok := r.Next()
		if !ok {
			t.Fatalf("expected super-kmer %d, got EOF", i)
		}
		if len(sk.Sequence) != len(expected) {
			t.Fatalf("sk %d: length %d, want %d", i, len(sk.Sequence), len(expected))
		}
		for j := range sk.Sequence {
			got := sk.Sequence[j] | 0x20
			want := expected[j] | 0x20
			if got != want {
				t.Fatalf("sk %d pos %d: got %c, want %c", i, j, got, want)
			}
		}
	}
}

func TestSkmFileSize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "size.skm")

	// Write a sequence of known length
	seq := []byte("ACGTACGTAC") // 10 bases
	sk := SuperKmer{Sequence: seq}

	w, err := NewSkmWriter(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Write(sk); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	// Expected: 2 bytes (length) + ceil(10/4)=3 bytes (data) = 5 bytes
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != 5 {
		t.Fatalf("file size: got %d, want 5", info.Size())
	}
}
