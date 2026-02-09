package obikmer

import (
	"path/filepath"
	"testing"
)

// writeKdi is a helper that writes sorted kmers to a .kdi file.
func writeKdi(t *testing.T, dir, name string, kmers []uint64) string {
	t.Helper()
	path := filepath.Join(dir, name)
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
	return path
}

func TestKWayMergeBasic(t *testing.T) {
	dir := t.TempDir()

	// Three sorted streams
	p1 := writeKdi(t, dir, "a.kdi", []uint64{1, 3, 5, 7})
	p2 := writeKdi(t, dir, "b.kdi", []uint64{2, 3, 6, 7})
	p3 := writeKdi(t, dir, "c.kdi", []uint64{3, 4, 7, 8})

	r1, _ := NewKdiReader(p1)
	r2, _ := NewKdiReader(p2)
	r3, _ := NewKdiReader(p3)

	m := NewKWayMerge([]*KdiReader{r1, r2, r3})
	defer m.Close()

	type result struct {
		kmer  uint64
		count int
	}
	var results []result
	for {
		kmer, count, ok := m.Next()
		if !ok {
			break
		}
		results = append(results, result{kmer, count})
	}

	expected := []result{
		{1, 1}, {2, 1}, {3, 3}, {4, 1}, {5, 1}, {6, 1}, {7, 3}, {8, 1},
	}
	if len(results) != len(expected) {
		t.Fatalf("got %d results, want %d", len(results), len(expected))
	}
	for i, exp := range expected {
		if results[i] != exp {
			t.Errorf("result %d: got %+v, want %+v", i, results[i], exp)
		}
	}
}

func TestKWayMergeSingleStream(t *testing.T) {
	dir := t.TempDir()
	p := writeKdi(t, dir, "a.kdi", []uint64{10, 20, 30})

	r, _ := NewKdiReader(p)
	m := NewKWayMerge([]*KdiReader{r})
	defer m.Close()

	vals := []uint64{10, 20, 30}
	for _, expected := range vals {
		kmer, count, ok := m.Next()
		if !ok {
			t.Fatal("unexpected EOF")
		}
		if kmer != expected || count != 1 {
			t.Fatalf("got (%d, %d), want (%d, 1)", kmer, count, expected)
		}
	}
	_, _, ok := m.Next()
	if ok {
		t.Fatal("expected EOF")
	}
}

func TestKWayMergeEmpty(t *testing.T) {
	dir := t.TempDir()

	p1 := writeKdi(t, dir, "a.kdi", nil)
	p2 := writeKdi(t, dir, "b.kdi", nil)

	r1, _ := NewKdiReader(p1)
	r2, _ := NewKdiReader(p2)

	m := NewKWayMerge([]*KdiReader{r1, r2})
	defer m.Close()

	_, _, ok := m.Next()
	if ok {
		t.Fatal("expected no results from empty streams")
	}
}

func TestKWayMergeDisjoint(t *testing.T) {
	dir := t.TempDir()

	p1 := writeKdi(t, dir, "a.kdi", []uint64{1, 2, 3})
	p2 := writeKdi(t, dir, "b.kdi", []uint64{10, 20, 30})

	r1, _ := NewKdiReader(p1)
	r2, _ := NewKdiReader(p2)

	m := NewKWayMerge([]*KdiReader{r1, r2})
	defer m.Close()

	expected := []uint64{1, 2, 3, 10, 20, 30}
	for _, exp := range expected {
		kmer, count, ok := m.Next()
		if !ok {
			t.Fatal("unexpected EOF")
		}
		if kmer != exp || count != 1 {
			t.Fatalf("got (%d, %d), want (%d, 1)", kmer, count, exp)
		}
	}
}

func TestKWayMergeAllSame(t *testing.T) {
	dir := t.TempDir()

	p1 := writeKdi(t, dir, "a.kdi", []uint64{42})
	p2 := writeKdi(t, dir, "b.kdi", []uint64{42})
	p3 := writeKdi(t, dir, "c.kdi", []uint64{42})

	r1, _ := NewKdiReader(p1)
	r2, _ := NewKdiReader(p2)
	r3, _ := NewKdiReader(p3)

	m := NewKWayMerge([]*KdiReader{r1, r2, r3})
	defer m.Close()

	kmer, count, ok := m.Next()
	if !ok {
		t.Fatal("expected one result")
	}
	if kmer != 42 || count != 3 {
		t.Fatalf("got (%d, %d), want (42, 3)", kmer, count)
	}
	_, _, ok = m.Next()
	if ok {
		t.Fatal("expected EOF")
	}
}
