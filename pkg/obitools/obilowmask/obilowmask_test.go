package obilowmask

import (
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

func TestLowMaskWorker(t *testing.T) {
	worker := LowMaskWorker(31, 6, 0.3, Mask, 'n')

	seq := obiseq.NewBioSequence("test", []byte("acgtacgtacgtacgtacgtacgtacgtacgt"), "test")
	result, err := worker(seq)
	if err != nil {
		t.Fatalf("Worker failed: %v", err)
	}

	if result.Len() != 1 {
		t.Fatalf("Expected 1 sequence, got %d", result.Len())
	}

	resultSeq := result[0]
	if resultSeq.Len() != 32 {
		t.Fatalf("Expected sequence length 32, got %d", resultSeq.Len())
	}
}

func TestLowMaskWorkerWithAmbiguity(t *testing.T) {
	worker := LowMaskWorker(31, 6, 0.3, Mask, 'n')

	seq := obiseq.NewBioSequence("test", []byte("acgtNcgtacgtacgtacgtacgtacgtacgt"), "test")
	result, err := worker(seq)
	if err != nil {
		t.Fatalf("Worker failed: %v", err)
	}

	if result.Len() != 1 {
		t.Fatalf("Expected 1 sequence, got %d", result.Len())
	}
}
