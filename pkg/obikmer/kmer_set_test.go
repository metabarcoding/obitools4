package obikmer

import (
	"math"
	"testing"
)

func TestJaccardDistanceIdentical(t *testing.T) {
	ks1 := NewKmerSet(5)
	ks1.AddKmerCode(100)
	ks1.AddKmerCode(200)
	ks1.AddKmerCode(300)

	ks2 := NewKmerSet(5)
	ks2.AddKmerCode(100)
	ks2.AddKmerCode(200)
	ks2.AddKmerCode(300)

	distance := ks1.JaccardDistance(ks2)
	similarity := ks1.JaccardSimilarity(ks2)

	if distance != 0.0 {
		t.Errorf("Expected distance 0.0 for identical sets, got %f", distance)
	}

	if similarity != 1.0 {
		t.Errorf("Expected similarity 1.0 for identical sets, got %f", similarity)
	}
}

func TestJaccardDistanceDisjoint(t *testing.T) {
	ks1 := NewKmerSet(5)
	ks1.AddKmerCode(100)
	ks1.AddKmerCode(200)
	ks1.AddKmerCode(300)

	ks2 := NewKmerSet(5)
	ks2.AddKmerCode(400)
	ks2.AddKmerCode(500)
	ks2.AddKmerCode(600)

	distance := ks1.JaccardDistance(ks2)
	similarity := ks1.JaccardSimilarity(ks2)

	if distance != 1.0 {
		t.Errorf("Expected distance 1.0 for disjoint sets, got %f", distance)
	}

	if similarity != 0.0 {
		t.Errorf("Expected similarity 0.0 for disjoint sets, got %f", similarity)
	}
}

func TestJaccardDistancePartialOverlap(t *testing.T) {
	// Set 1: {1, 2, 3}
	ks1 := NewKmerSet(5)
	ks1.AddKmerCode(1)
	ks1.AddKmerCode(2)
	ks1.AddKmerCode(3)

	// Set 2: {2, 3, 4}
	ks2 := NewKmerSet(5)
	ks2.AddKmerCode(2)
	ks2.AddKmerCode(3)
	ks2.AddKmerCode(4)

	// Intersection: {2, 3} -> cardinality = 2
	// Union: {1, 2, 3, 4} -> cardinality = 4
	// Similarity = 2/4 = 0.5
	// Distance = 1 - 0.5 = 0.5

	distance := ks1.JaccardDistance(ks2)
	similarity := ks1.JaccardSimilarity(ks2)

	expectedDistance := 0.5
	expectedSimilarity := 0.5

	if math.Abs(distance-expectedDistance) > 1e-10 {
		t.Errorf("Expected distance %f, got %f", expectedDistance, distance)
	}

	if math.Abs(similarity-expectedSimilarity) > 1e-10 {
		t.Errorf("Expected similarity %f, got %f", expectedSimilarity, similarity)
	}
}

func TestJaccardDistanceOneSubsetOfOther(t *testing.T) {
	// Set 1: {1, 2}
	ks1 := NewKmerSet(5)
	ks1.AddKmerCode(1)
	ks1.AddKmerCode(2)

	// Set 2: {1, 2, 3, 4}
	ks2 := NewKmerSet(5)
	ks2.AddKmerCode(1)
	ks2.AddKmerCode(2)
	ks2.AddKmerCode(3)
	ks2.AddKmerCode(4)

	// Intersection: {1, 2} -> cardinality = 2
	// Union: {1, 2, 3, 4} -> cardinality = 4
	// Similarity = 2/4 = 0.5
	// Distance = 1 - 0.5 = 0.5

	distance := ks1.JaccardDistance(ks2)
	similarity := ks1.JaccardSimilarity(ks2)

	expectedDistance := 0.5
	expectedSimilarity := 0.5

	if math.Abs(distance-expectedDistance) > 1e-10 {
		t.Errorf("Expected distance %f, got %f", expectedDistance, distance)
	}

	if math.Abs(similarity-expectedSimilarity) > 1e-10 {
		t.Errorf("Expected similarity %f, got %f", expectedSimilarity, similarity)
	}
}

func TestJaccardDistanceEmptySets(t *testing.T) {
	ks1 := NewKmerSet(5)
	ks2 := NewKmerSet(5)

	distance := ks1.JaccardDistance(ks2)
	similarity := ks1.JaccardSimilarity(ks2)

	// By convention, distance = 1.0 for empty sets
	if distance != 1.0 {
		t.Errorf("Expected distance 1.0 for empty sets, got %f", distance)
	}

	if similarity != 0.0 {
		t.Errorf("Expected similarity 0.0 for empty sets, got %f", similarity)
	}
}

func TestJaccardDistanceOneEmpty(t *testing.T) {
	ks1 := NewKmerSet(5)
	ks1.AddKmerCode(1)
	ks1.AddKmerCode(2)
	ks1.AddKmerCode(3)

	ks2 := NewKmerSet(5)

	distance := ks1.JaccardDistance(ks2)
	similarity := ks1.JaccardSimilarity(ks2)

	// Intersection: {} -> cardinality = 0
	// Union: {1, 2, 3} -> cardinality = 3
	// Similarity = 0/3 = 0.0
	// Distance = 1.0

	if distance != 1.0 {
		t.Errorf("Expected distance 1.0 when one set is empty, got %f", distance)
	}

	if similarity != 0.0 {
		t.Errorf("Expected similarity 0.0 when one set is empty, got %f", similarity)
	}
}

func TestJaccardDistanceDifferentK(t *testing.T) {
	ks1 := NewKmerSet(5)
	ks1.AddKmerCode(1)

	ks2 := NewKmerSet(7)
	ks2.AddKmerCode(1)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic when computing Jaccard distance with different k values")
		}
	}()

	_ = ks1.JaccardDistance(ks2)
}

func TestJaccardDistanceSimilarityRelation(t *testing.T) {
	// Test that distance + similarity = 1.0 for all cases
	testCases := []struct {
		name string
		ks1  *KmerSet
		ks2  *KmerSet
	}{
		{
			name: "partial overlap",
			ks1: func() *KmerSet {
				ks := NewKmerSet(5)
				ks.AddKmerCode(1)
				ks.AddKmerCode(2)
				ks.AddKmerCode(3)
				return ks
			}(),
			ks2: func() *KmerSet {
				ks := NewKmerSet(5)
				ks.AddKmerCode(2)
				ks.AddKmerCode(3)
				ks.AddKmerCode(4)
				ks.AddKmerCode(5)
				return ks
			}(),
		},
		{
			name: "identical",
			ks1: func() *KmerSet {
				ks := NewKmerSet(5)
				ks.AddKmerCode(10)
				ks.AddKmerCode(20)
				return ks
			}(),
			ks2: func() *KmerSet {
				ks := NewKmerSet(5)
				ks.AddKmerCode(10)
				ks.AddKmerCode(20)
				return ks
			}(),
		},
		{
			name: "disjoint",
			ks1: func() *KmerSet {
				ks := NewKmerSet(5)
				ks.AddKmerCode(1)
				return ks
			}(),
			ks2: func() *KmerSet {
				ks := NewKmerSet(5)
				ks.AddKmerCode(100)
				return ks
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			distance := tc.ks1.JaccardDistance(tc.ks2)
			similarity := tc.ks1.JaccardSimilarity(tc.ks2)

			sum := distance + similarity

			if math.Abs(sum-1.0) > 1e-10 {
				t.Errorf("Expected distance + similarity = 1.0, got %f + %f = %f",
					distance, similarity, sum)
			}
		})
	}
}

func TestJaccardDistanceSymmetry(t *testing.T) {
	ks1 := NewKmerSet(5)
	ks1.AddKmerCode(1)
	ks1.AddKmerCode(2)
	ks1.AddKmerCode(3)

	ks2 := NewKmerSet(5)
	ks2.AddKmerCode(2)
	ks2.AddKmerCode(3)
	ks2.AddKmerCode(4)

	distance1 := ks1.JaccardDistance(ks2)
	distance2 := ks2.JaccardDistance(ks1)

	similarity1 := ks1.JaccardSimilarity(ks2)
	similarity2 := ks2.JaccardSimilarity(ks1)

	if math.Abs(distance1-distance2) > 1e-10 {
		t.Errorf("Jaccard distance not symmetric: %f vs %f", distance1, distance2)
	}

	if math.Abs(similarity1-similarity2) > 1e-10 {
		t.Errorf("Jaccard similarity not symmetric: %f vs %f", similarity1, similarity2)
	}
}
