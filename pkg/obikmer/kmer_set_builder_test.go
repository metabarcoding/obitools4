package obikmer

import (
	"sort"
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

func TestBuilderBasic(t *testing.T) {
	dir := t.TempDir()

	builder, err := NewKmerSetGroupBuilder(dir, 15, 7, 1, 64)
	if err != nil {
		t.Fatal(err)
	}

	seq := obiseq.NewBioSequence("test", []byte("ACGTACGTACGTACGTACGTACGTACGT"), "")
	builder.AddSequence(0, seq)

	ksg, err := builder.Close()
	if err != nil {
		t.Fatal(err)
	}

	if ksg.K() != 15 {
		t.Fatalf("K() = %d, want 15", ksg.K())
	}
	if ksg.M() != 7 {
		t.Fatalf("M() = %d, want 7", ksg.M())
	}
	if ksg.Partitions() != 64 {
		t.Fatalf("Partitions() = %d, want 64", ksg.Partitions())
	}
	if ksg.Size() != 1 {
		t.Fatalf("Size() = %d, want 1", ksg.Size())
	}
	if ksg.Len(0) == 0 {
		t.Fatal("Len(0) = 0, expected some k-mers")
	}

	// Verify k-mers match what we'd compute directly
	var expected []uint64
	for kmer := range IterCanonicalKmers(seq.Sequence(), 15) {
		expected = append(expected, kmer)
	}
	sort.Slice(expected, func(i, j int) bool { return expected[i] < expected[j] })
	// Dedup
	deduped := expected[:0]
	for i, v := range expected {
		if i == 0 || v != expected[i-1] {
			deduped = append(deduped, v)
		}
	}

	if ksg.Len(0) != uint64(len(deduped)) {
		t.Fatalf("Len(0) = %d, expected %d unique k-mers", ksg.Len(0), len(deduped))
	}

	// Check iterator
	var fromIter []uint64
	for kmer := range ksg.Iterator(0) {
		fromIter = append(fromIter, kmer)
	}
	// The iterator does a k-way merge so should be sorted
	for i := 1; i < len(fromIter); i++ {
		if fromIter[i] <= fromIter[i-1] {
			t.Fatalf("iterator not sorted at %d: %d <= %d", i, fromIter[i], fromIter[i-1])
		}
	}
	if len(fromIter) != len(deduped) {
		t.Fatalf("iterator yielded %d k-mers, expected %d", len(fromIter), len(deduped))
	}
	for i, v := range fromIter {
		if v != deduped[i] {
			t.Fatalf("iterator kmer %d: got %d, want %d", i, v, deduped[i])
		}
	}
}

func TestBuilderMultipleSequences(t *testing.T) {
	dir := t.TempDir()

	builder, err := NewKmerSetGroupBuilder(dir, 15, 7, 1, 64)
	if err != nil {
		t.Fatal(err)
	}

	seqs := []string{
		"ACGTACGTACGTACGTACGTACGTACGT",
		"TTTTTTTTTTTTTTTTTTTTTTTTT",
		"GGGGGGGGGGGGGGGGGGGGGGGG",
	}
	for _, s := range seqs {
		seq := obiseq.NewBioSequence("", []byte(s), "")
		builder.AddSequence(0, seq)
	}

	ksg, err := builder.Close()
	if err != nil {
		t.Fatal(err)
	}

	if ksg.Len(0) == 0 {
		t.Fatal("expected k-mers after multiple sequences")
	}
}

func TestBuilderFrequencyFilter(t *testing.T) {
	dir := t.TempDir()

	builder, err := NewKmerSetGroupBuilder(dir, 15, 7, 1, 64,
		WithMinFrequency(3))
	if err != nil {
		t.Fatal(err)
	}

	// Add same sequence 3 times — all k-mers should survive freq=3
	seq := obiseq.NewBioSequence("test", []byte("ACGTACGTACGTACGTACGTACGTACGT"), "")
	for i := 0; i < 3; i++ {
		builder.AddSequence(0, seq)
	}

	ksg, err := builder.Close()
	if err != nil {
		t.Fatal(err)
	}

	// All k-mers appear exactly 3 times → all should survive
	var expected []uint64
	for kmer := range IterCanonicalKmers(seq.Sequence(), 15) {
		expected = append(expected, kmer)
	}
	sort.Slice(expected, func(i, j int) bool { return expected[i] < expected[j] })
	deduped := expected[:0]
	for i, v := range expected {
		if i == 0 || v != expected[i-1] {
			deduped = append(deduped, v)
		}
	}

	if ksg.Len(0) != uint64(len(deduped)) {
		t.Fatalf("Len(0) = %d, expected %d (all k-mers at freq=3)", ksg.Len(0), len(deduped))
	}
}

func TestBuilderFrequencyFilterRejects(t *testing.T) {
	dir := t.TempDir()

	builder, err := NewKmerSetGroupBuilder(dir, 15, 7, 1, 64,
		WithMinFrequency(5))
	if err != nil {
		t.Fatal(err)
	}

	// Use a non-repetitive sequence so each canonical k-mer appears once per pass.
	// Adding it twice gives freq=2 per kmer, which is < minFreq=5 → all rejected.
	seq := obiseq.NewBioSequence("test",
		[]byte("ACGATCGATCTAGCTAGCTGATCGATCGATCG"), "")
	builder.AddSequence(0, seq)
	builder.AddSequence(0, seq)

	ksg, err := builder.Close()
	if err != nil {
		t.Fatal(err)
	}

	if ksg.Len(0) != 0 {
		t.Fatalf("Len(0) = %d, expected 0 (all k-mers at freq=2 < minFreq=5)", ksg.Len(0))
	}
}

func TestBuilderMultipleSets(t *testing.T) {
	dir := t.TempDir()

	builder, err := NewKmerSetGroupBuilder(dir, 15, 7, 3, 64)
	if err != nil {
		t.Fatal(err)
	}

	seqs := []string{
		"ACGTACGTACGTACGTACGTACGTACGT",
		"TTTTTTTTTTTTTTTTTTTTTTTTT",
		"GGGGGGGGGGGGGGGGGGGGGGGG",
	}
	for i, s := range seqs {
		seq := obiseq.NewBioSequence("", []byte(s), "")
		builder.AddSequence(i, seq)
	}

	ksg, err := builder.Close()
	if err != nil {
		t.Fatal(err)
	}

	if ksg.Size() != 3 {
		t.Fatalf("Size() = %d, want 3", ksg.Size())
	}
	for s := 0; s < 3; s++ {
		if ksg.Len(s) == 0 {
			t.Fatalf("Len(%d) = 0, expected some k-mers", s)
		}
	}
}

func TestBuilderOpenRoundTrip(t *testing.T) {
	dir := t.TempDir()

	builder, err := NewKmerSetGroupBuilder(dir, 15, 7, 1, 64)
	if err != nil {
		t.Fatal(err)
	}

	seq := obiseq.NewBioSequence("test", []byte("ACGTACGTACGTACGTACGTACGTACGT"), "")
	builder.AddSequence(0, seq)

	ksg1, err := builder.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Reopen
	ksg2, err := OpenKmerSetGroup(dir)
	if err != nil {
		t.Fatal(err)
	}

	if ksg2.K() != ksg1.K() {
		t.Fatalf("K mismatch: %d vs %d", ksg2.K(), ksg1.K())
	}
	if ksg2.M() != ksg1.M() {
		t.Fatalf("M mismatch: %d vs %d", ksg2.M(), ksg1.M())
	}
	if ksg2.Partitions() != ksg1.Partitions() {
		t.Fatalf("Partitions mismatch: %d vs %d", ksg2.Partitions(), ksg1.Partitions())
	}
	if ksg2.Len(0) != ksg1.Len(0) {
		t.Fatalf("Len mismatch: %d vs %d", ksg2.Len(0), ksg1.Len(0))
	}
}

func TestBuilderAttributes(t *testing.T) {
	dir := t.TempDir()

	builder, err := NewKmerSetGroupBuilder(dir, 15, 7, 1, 64)
	if err != nil {
		t.Fatal(err)
	}

	seq := obiseq.NewBioSequence("test", []byte("ACGTACGTACGTACGTACGTACGTACGT"), "")
	builder.AddSequence(0, seq)

	ksg, err := builder.Close()
	if err != nil {
		t.Fatal(err)
	}

	ksg.SetId("my_index")
	ksg.SetAttribute("organism", "test")
	ksg.SaveMetadata()

	// Reopen and check
	ksg2, err := OpenKmerSetGroup(dir)
	if err != nil {
		t.Fatal(err)
	}

	if ksg2.Id() != "my_index" {
		t.Fatalf("Id() = %q, want %q", ksg2.Id(), "my_index")
	}
	if !ksg2.HasAttribute("organism") {
		t.Fatal("expected 'organism' attribute")
	}
	v, _ := ksg2.GetAttribute("organism")
	if v != "test" {
		t.Fatalf("organism = %v, want 'test'", v)
	}
}
