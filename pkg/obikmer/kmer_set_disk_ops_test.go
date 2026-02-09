package obikmer

import (
	"path/filepath"
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// buildGroupFromSeqs creates a KmerSetGroup with one set per sequence.
func buildGroupFromSeqs(t *testing.T, dir string, k, m int, seqs []string) *KmerSetGroup {
	t.Helper()
	n := len(seqs)
	builder, err := NewKmerSetGroupBuilder(dir, k, m, n, 64)
	if err != nil {
		t.Fatal(err)
	}
	for i, s := range seqs {
		seq := obiseq.NewBioSequence("", []byte(s), "")
		builder.AddSequence(i, seq)
	}
	ksg, err := builder.Close()
	if err != nil {
		t.Fatal(err)
	}
	return ksg
}

func collectKmers(t *testing.T, ksg *KmerSetGroup, setIdx int) []uint64 {
	t.Helper()
	var result []uint64
	for kmer := range ksg.Iterator(setIdx) {
		result = append(result, kmer)
	}
	return result
}

func TestDiskOpsUnion(t *testing.T) {
	dir := t.TempDir()
	indexDir := filepath.Join(dir, "index")
	outDir := filepath.Join(dir, "union")

	// Two sequences with some overlap
	seqs := []string{
		"ACGATCGATCTAGCTAGCTGATCGATCGATCG",
		"CTAGCTAGCTGATCGATCGATCGTTTAAACCC",
	}
	ksg := buildGroupFromSeqs(t, indexDir, 15, 7, seqs)

	result, err := ksg.Union(outDir)
	if err != nil {
		t.Fatal(err)
	}

	// Union should have at least as many k-mers as each individual set
	unionLen := result.Len(0)
	if unionLen == 0 {
		t.Fatal("union is empty")
	}
	if unionLen < ksg.Len(0) || unionLen < ksg.Len(1) {
		t.Fatalf("union (%d) smaller than an input set (%d, %d)", unionLen, ksg.Len(0), ksg.Len(1))
	}

	// Union should not exceed the sum of both sets
	if unionLen > ksg.Len(0)+ksg.Len(1) {
		t.Fatalf("union (%d) larger than sum of sets (%d)", unionLen, ksg.Len(0)+ksg.Len(1))
	}
}

func TestDiskOpsIntersect(t *testing.T) {
	dir := t.TempDir()
	indexDir := filepath.Join(dir, "index")
	outDir := filepath.Join(dir, "intersect")

	// Two sequences with some shared k-mers
	seqs := []string{
		"ACGATCGATCTAGCTAGCTGATCGATCGATCG",
		"CTAGCTAGCTGATCGATCGATCGTTTAAACCC",
	}
	ksg := buildGroupFromSeqs(t, indexDir, 15, 7, seqs)

	result, err := ksg.Intersect(outDir)
	if err != nil {
		t.Fatal(err)
	}

	interLen := result.Len(0)
	// Intersection should not be bigger than any individual set
	if interLen > ksg.Len(0) || interLen > ksg.Len(1) {
		t.Fatalf("intersection (%d) larger than input sets (%d, %d)", interLen, ksg.Len(0), ksg.Len(1))
	}
}

func TestDiskOpsDifference(t *testing.T) {
	dir := t.TempDir()
	indexDir := filepath.Join(dir, "index")
	outDir := filepath.Join(dir, "diff")

	seqs := []string{
		"ACGATCGATCTAGCTAGCTGATCGATCGATCG",
		"CTAGCTAGCTGATCGATCGATCGTTTAAACCC",
	}
	ksg := buildGroupFromSeqs(t, indexDir, 15, 7, seqs)

	result, err := ksg.Difference(outDir)
	if err != nil {
		t.Fatal(err)
	}

	diffLen := result.Len(0)
	// Difference = set_0 - set_1, so should be <= set_0
	if diffLen > ksg.Len(0) {
		t.Fatalf("difference (%d) larger than set_0 (%d)", diffLen, ksg.Len(0))
	}
}

func TestDiskOpsConsistency(t *testing.T) {
	dir := t.TempDir()
	indexDir := filepath.Join(dir, "index")

	seqs := []string{
		"ACGATCGATCTAGCTAGCTGATCGATCGATCG",
		"CTAGCTAGCTGATCGATCGATCGTTTAAACCC",
	}
	ksg := buildGroupFromSeqs(t, indexDir, 15, 7, seqs)

	unionResult, err := ksg.Union(filepath.Join(dir, "union"))
	if err != nil {
		t.Fatal(err)
	}
	interResult, err := ksg.Intersect(filepath.Join(dir, "intersect"))
	if err != nil {
		t.Fatal(err)
	}
	diffResult, err := ksg.Difference(filepath.Join(dir, "diff"))
	if err != nil {
		t.Fatal(err)
	}

	unionLen := unionResult.Len(0)
	interLen := interResult.Len(0)
	diffLen := diffResult.Len(0)

	// |A ∪ B| = |A| + |B| - |A ∩ B|
	expectedUnion := ksg.Len(0) + ksg.Len(1) - interLen
	if unionLen != expectedUnion {
		t.Fatalf("|A∪B|=%d, expected |A|+|B|-|A∩B|=%d+%d-%d=%d",
			unionLen, ksg.Len(0), ksg.Len(1), interLen, expectedUnion)
	}

	// |A \ B| = |A| - |A ∩ B|
	expectedDiff := ksg.Len(0) - interLen
	if diffLen != expectedDiff {
		t.Fatalf("|A\\B|=%d, expected |A|-|A∩B|=%d-%d=%d",
			diffLen, ksg.Len(0), interLen, expectedDiff)
	}
}

func TestDiskOpsQuorum(t *testing.T) {
	dir := t.TempDir()
	indexDir := filepath.Join(dir, "index")

	// Three sets
	seqs := []string{
		"ACGATCGATCTAGCTAGCTGATCGATCGATCG",
		"CTAGCTAGCTGATCGATCGATCGTTTAAACCC",
		"GATCGATCGATCGAAATTTCCCGGG",
	}
	ksg := buildGroupFromSeqs(t, indexDir, 15, 7, seqs)

	// QuorumAtLeast(1) = Union
	q1, err := ksg.QuorumAtLeast(1, filepath.Join(dir, "q1"))
	if err != nil {
		t.Fatal(err)
	}
	union, err := ksg.Union(filepath.Join(dir, "union"))
	if err != nil {
		t.Fatal(err)
	}
	if q1.Len(0) != union.Len(0) {
		t.Fatalf("QuorumAtLeast(1)=%d != Union=%d", q1.Len(0), union.Len(0))
	}

	// QuorumAtLeast(3) = Intersect
	q3, err := ksg.QuorumAtLeast(3, filepath.Join(dir, "q3"))
	if err != nil {
		t.Fatal(err)
	}
	inter, err := ksg.Intersect(filepath.Join(dir, "inter"))
	if err != nil {
		t.Fatal(err)
	}
	if q3.Len(0) != inter.Len(0) {
		t.Fatalf("QuorumAtLeast(3)=%d != Intersect=%d", q3.Len(0), inter.Len(0))
	}

	// QuorumAtLeast(2) should be between Intersect and Union
	q2, err := ksg.QuorumAtLeast(2, filepath.Join(dir, "q2"))
	if err != nil {
		t.Fatal(err)
	}
	if q2.Len(0) < q3.Len(0) || q2.Len(0) > q1.Len(0) {
		t.Fatalf("QuorumAtLeast(2)=%d not between intersect=%d and union=%d",
			q2.Len(0), q3.Len(0), q1.Len(0))
	}
}

func TestDiskOpsJaccard(t *testing.T) {
	dir := t.TempDir()
	indexDir := filepath.Join(dir, "index")

	seqs := []string{
		"ACGATCGATCTAGCTAGCTGATCGATCGATCG",
		"ACGATCGATCTAGCTAGCTGATCGATCGATCG", // identical to first
		"TTTTTTTTTTTTTTTTTTTTTTTTT",        // completely different
	}
	ksg := buildGroupFromSeqs(t, indexDir, 15, 7, seqs)

	dm := ksg.JaccardDistanceMatrix()
	if dm == nil {
		t.Fatal("JaccardDistanceMatrix returned nil")
	}

	// Identical sets should have distance 0
	d01 := dm.Get(0, 1)
	if d01 != 0.0 {
		t.Fatalf("distance(0,1) = %f, expected 0.0 for identical sets", d01)
	}

	// Completely different sets should have distance 1.0
	d02 := dm.Get(0, 2)
	if d02 != 1.0 {
		t.Fatalf("distance(0,2) = %f, expected 1.0 for disjoint sets", d02)
	}

	// Similarity matrix
	sm := ksg.JaccardSimilarityMatrix()
	if sm == nil {
		t.Fatal("JaccardSimilarityMatrix returned nil")
	}

	s01 := sm.Get(0, 1)
	if s01 != 1.0 {
		t.Fatalf("similarity(0,1) = %f, expected 1.0 for identical sets", s01)
	}

	s02 := sm.Get(0, 2)
	if s02 != 0.0 {
		t.Fatalf("similarity(0,2) = %f, expected 0.0 for disjoint sets", s02)
	}
}
