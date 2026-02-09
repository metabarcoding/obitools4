package obisuperkmer

import (
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

func TestCLIExtractSuperKmers(t *testing.T) {
	// Create a test sequence
	testSeq := obiseq.NewBioSequence(
		"test_seq",
		[]byte("ACGTACGTACGTACGTACGTACGTACGTACGT"),
		"",
	)

	// Create a batch with the test sequence
	batch := obiseq.NewBioSequenceBatch()
	batch.Add(testSeq)

	// Create an iterator from the batch
	iterator := obiiter.MakeBioSequenceBatchChannel(1)
	go func() {
		iterator.Push(batch)
		iterator.Close()
	}()

	// Set test parameters
	SetKmerSize(15)
	SetMinimizerSize(7)

	// Extract super k-mers
	result := CLIExtractSuperKmers(iterator)

	// Count the number of super k-mers
	count := 0
	for result.Next() {
		batch := result.Get()
		for _, sk := range batch.Slice() {
			count++

			// Verify that the super k-mer has the expected attributes
			if !sk.HasAttribute("minimizer_value") {
				t.Error("Super k-mer missing 'minimizer_value' attribute")
			}
			if !sk.HasAttribute("minimizer_seq") {
				t.Error("Super k-mer missing 'minimizer_seq' attribute")
			}
			if !sk.HasAttribute("k") {
				t.Error("Super k-mer missing 'k' attribute")
			}
			if !sk.HasAttribute("m") {
				t.Error("Super k-mer missing 'm' attribute")
			}
			if !sk.HasAttribute("start") {
				t.Error("Super k-mer missing 'start' attribute")
			}
			if !sk.HasAttribute("end") {
				t.Error("Super k-mer missing 'end' attribute")
			}
			if !sk.HasAttribute("parent_id") {
				t.Error("Super k-mer missing 'parent_id' attribute")
			}

			// Verify attribute values
			k, _ := sk.GetIntAttribute("k")
			m, _ := sk.GetIntAttribute("m")

			if k != 15 {
				t.Errorf("Expected k=15, got k=%d", k)
			}
			if m != 7 {
				t.Errorf("Expected m=7, got m=%d", m)
			}

			parentID, _ := sk.GetStringAttribute("parent_id")
			if parentID != "test_seq" {
				t.Errorf("Expected parent_id='test_seq', got '%s'", parentID)
			}
		}
	}

	if count == 0 {
		t.Error("No super k-mers were extracted")
	}

	t.Logf("Extracted %d super k-mers from test sequence", count)
}

func TestOptionGettersAndSetters(t *testing.T) {
	// Test initial values
	if CLIKmerSize() != 21 {
		t.Errorf("Expected default k-mer size 21, got %d", CLIKmerSize())
	}
	if CLIMinimizerSize() != 11 {
		t.Errorf("Expected default minimizer size 11, got %d", CLIMinimizerSize())
	}

	// Test setters
	SetKmerSize(25)
	SetMinimizerSize(13)

	if CLIKmerSize() != 25 {
		t.Errorf("SetKmerSize failed: expected 25, got %d", CLIKmerSize())
	}
	if CLIMinimizerSize() != 13 {
		t.Errorf("SetMinimizerSize failed: expected 13, got %d", CLIMinimizerSize())
	}

	// Reset to defaults
	SetKmerSize(21)
	SetMinimizerSize(11)
}

func BenchmarkCLIExtractSuperKmers(b *testing.B) {
	// Create a longer test sequence
	longSeq := make([]byte, 1000)
	bases := []byte{'A', 'C', 'G', 'T'}
	for i := range longSeq {
		longSeq[i] = bases[i%4]
	}

	testSeq := obiseq.NewBioSequence("bench_seq", longSeq, "")

	// Set parameters
	SetKmerSize(21)
	SetMinimizerSize(11)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		batch := obiseq.NewBioSequenceBatch()
		batch.Add(testSeq)

		iterator := obiiter.MakeBioSequenceBatchChannel(1)
		go func() {
			iterator.Push(batch)
			iterator.Close()
		}()

		result := CLIExtractSuperKmers(iterator)

		// Consume the iterator
		for result.Next() {
			result.Get()
		}
	}
}
