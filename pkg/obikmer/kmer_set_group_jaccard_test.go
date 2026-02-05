package obikmer

import (
	"math"
	"testing"
)

func TestKmerSetGroupJaccardDistanceMatrix(t *testing.T) {
	ksg := NewKmerSetGroup(5, 3)

	// Set 0: {1, 2, 3}
	ksg.Get(0).AddKmerCode(1)
	ksg.Get(0).AddKmerCode(2)
	ksg.Get(0).AddKmerCode(3)
	ksg.Get(0).SetId("set_A")

	// Set 1: {2, 3, 4}
	ksg.Get(1).AddKmerCode(2)
	ksg.Get(1).AddKmerCode(3)
	ksg.Get(1).AddKmerCode(4)
	ksg.Get(1).SetId("set_B")

	// Set 2: {5, 6, 7}
	ksg.Get(2).AddKmerCode(5)
	ksg.Get(2).AddKmerCode(6)
	ksg.Get(2).AddKmerCode(7)
	ksg.Get(2).SetId("set_C")

	dm := ksg.JaccardDistanceMatrix()

	// Check labels
	if dm.GetLabel(0) != "set_A" {
		t.Errorf("Expected label 'set_A' at index 0, got '%s'", dm.GetLabel(0))
	}
	if dm.GetLabel(1) != "set_B" {
		t.Errorf("Expected label 'set_B' at index 1, got '%s'", dm.GetLabel(1))
	}
	if dm.GetLabel(2) != "set_C" {
		t.Errorf("Expected label 'set_C' at index 2, got '%s'", dm.GetLabel(2))
	}

	// Check distances
	// Distance(0, 1):
	// Intersection: {2, 3} -> 2 elements
	// Union: {1, 2, 3, 4} -> 4 elements
	// Similarity: 2/4 = 0.5
	// Distance: 1 - 0.5 = 0.5
	expectedDist01 := 0.5
	actualDist01 := dm.Get(0, 1)
	if math.Abs(actualDist01-expectedDist01) > 1e-10 {
		t.Errorf("Distance(0, 1): expected %f, got %f", expectedDist01, actualDist01)
	}

	// Distance(0, 2):
	// Intersection: {} -> 0 elements
	// Union: {1, 2, 3, 5, 6, 7} -> 6 elements
	// Similarity: 0/6 = 0
	// Distance: 1 - 0 = 1.0
	expectedDist02 := 1.0
	actualDist02 := dm.Get(0, 2)
	if math.Abs(actualDist02-expectedDist02) > 1e-10 {
		t.Errorf("Distance(0, 2): expected %f, got %f", expectedDist02, actualDist02)
	}

	// Distance(1, 2):
	// Intersection: {} -> 0 elements
	// Union: {2, 3, 4, 5, 6, 7} -> 6 elements
	// Similarity: 0/6 = 0
	// Distance: 1 - 0 = 1.0
	expectedDist12 := 1.0
	actualDist12 := dm.Get(1, 2)
	if math.Abs(actualDist12-expectedDist12) > 1e-10 {
		t.Errorf("Distance(1, 2): expected %f, got %f", expectedDist12, actualDist12)
	}

	// Check symmetry
	if dm.Get(0, 1) != dm.Get(1, 0) {
		t.Errorf("Matrix not symmetric: Get(0, 1) = %f, Get(1, 0) = %f",
			dm.Get(0, 1), dm.Get(1, 0))
	}

	// Check diagonal
	if dm.Get(0, 0) != 0.0 {
		t.Errorf("Diagonal should be 0, got %f", dm.Get(0, 0))
	}
	if dm.Get(1, 1) != 0.0 {
		t.Errorf("Diagonal should be 0, got %f", dm.Get(1, 1))
	}
	if dm.Get(2, 2) != 0.0 {
		t.Errorf("Diagonal should be 0, got %f", dm.Get(2, 2))
	}
}

func TestKmerSetGroupJaccardSimilarityMatrix(t *testing.T) {
	ksg := NewKmerSetGroup(5, 3)

	// Set 0: {1, 2, 3}
	ksg.Get(0).AddKmerCode(1)
	ksg.Get(0).AddKmerCode(2)
	ksg.Get(0).AddKmerCode(3)

	// Set 1: {2, 3, 4}
	ksg.Get(1).AddKmerCode(2)
	ksg.Get(1).AddKmerCode(3)
	ksg.Get(1).AddKmerCode(4)

	// Set 2: {1, 2, 3} (same as set 0)
	ksg.Get(2).AddKmerCode(1)
	ksg.Get(2).AddKmerCode(2)
	ksg.Get(2).AddKmerCode(3)

	sm := ksg.JaccardSimilarityMatrix()

	// Check similarities
	// Similarity(0, 1): 0.5 (as calculated above)
	expectedSim01 := 0.5
	actualSim01 := sm.Get(0, 1)
	if math.Abs(actualSim01-expectedSim01) > 1e-10 {
		t.Errorf("Similarity(0, 1): expected %f, got %f", expectedSim01, actualSim01)
	}

	// Similarity(0, 2): 1.0 (identical sets)
	expectedSim02 := 1.0
	actualSim02 := sm.Get(0, 2)
	if math.Abs(actualSim02-expectedSim02) > 1e-10 {
		t.Errorf("Similarity(0, 2): expected %f, got %f", expectedSim02, actualSim02)
	}

	// Similarity(1, 2): 0.5
	// Intersection: {2, 3} -> 2
	// Union: {1, 2, 3, 4} -> 4
	// Similarity: 2/4 = 0.5
	expectedSim12 := 0.5
	actualSim12 := sm.Get(1, 2)
	if math.Abs(actualSim12-expectedSim12) > 1e-10 {
		t.Errorf("Similarity(1, 2): expected %f, got %f", expectedSim12, actualSim12)
	}

	// Check diagonal (similarity to self = 1.0)
	if sm.Get(0, 0) != 1.0 {
		t.Errorf("Diagonal should be 1.0, got %f", sm.Get(0, 0))
	}
	if sm.Get(1, 1) != 1.0 {
		t.Errorf("Diagonal should be 1.0, got %f", sm.Get(1, 1))
	}
	if sm.Get(2, 2) != 1.0 {
		t.Errorf("Diagonal should be 1.0, got %f", sm.Get(2, 2))
	}
}

func TestKmerSetGroupJaccardMatricesRelation(t *testing.T) {
	ksg := NewKmerSetGroup(5, 4)

	// Create different sets
	ksg.Get(0).AddKmerCode(1)
	ksg.Get(0).AddKmerCode(2)

	ksg.Get(1).AddKmerCode(2)
	ksg.Get(1).AddKmerCode(3)

	ksg.Get(2).AddKmerCode(1)
	ksg.Get(2).AddKmerCode(2)
	ksg.Get(2).AddKmerCode(3)

	ksg.Get(3).AddKmerCode(10)
	ksg.Get(3).AddKmerCode(20)

	dm := ksg.JaccardDistanceMatrix()
	sm := ksg.JaccardSimilarityMatrix()

	// For all pairs (including diagonal), distance + similarity should equal 1.0
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			distance := dm.Get(i, j)
			similarity := sm.Get(i, j)
			sum := distance + similarity

			if math.Abs(sum-1.0) > 1e-10 {
				t.Errorf("At (%d, %d): distance %f + similarity %f = %f, expected 1.0",
					i, j, distance, similarity, sum)
			}
		}
	}
}

func TestKmerSetGroupJaccardMatrixLabels(t *testing.T) {
	ksg := NewKmerSetGroup(5, 3)

	// Don't set IDs - should use default labels
	ksg.Get(0).AddKmerCode(1)
	ksg.Get(1).AddKmerCode(2)
	ksg.Get(2).AddKmerCode(3)

	dm := ksg.JaccardDistanceMatrix()

	// Check default labels
	if dm.GetLabel(0) != "set_0" {
		t.Errorf("Expected default label 'set_0', got '%s'", dm.GetLabel(0))
	}
	if dm.GetLabel(1) != "set_1" {
		t.Errorf("Expected default label 'set_1', got '%s'", dm.GetLabel(1))
	}
	if dm.GetLabel(2) != "set_2" {
		t.Errorf("Expected default label 'set_2', got '%s'", dm.GetLabel(2))
	}
}

func TestKmerSetGroupJaccardMatrixSize(t *testing.T) {
	ksg := NewKmerSetGroup(5, 5)

	for i := 0; i < 5; i++ {
		ksg.Get(i).AddKmerCode(uint64(i))
	}

	dm := ksg.JaccardDistanceMatrix()

	if dm.Size() != 5 {
		t.Errorf("Expected matrix size 5, got %d", dm.Size())
	}

	// All sets are disjoint, so all distances should be 1.0
	for i := 0; i < 5; i++ {
		for j := i + 1; j < 5; j++ {
			dist := dm.Get(i, j)
			if math.Abs(dist-1.0) > 1e-10 {
				t.Errorf("Expected distance 1.0 for disjoint sets (%d, %d), got %f",
					i, j, dist)
			}
		}
	}
}
