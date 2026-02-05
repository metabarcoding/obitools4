package obidist

import (
	"math"
	"testing"
)

func TestNewDistMatrix(t *testing.T) {
	dm := NewDistMatrix(5)

	if dm.Size() != 5 {
		t.Errorf("Expected size 5, got %d", dm.Size())
	}

	// Check that all values are initialized to 0
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			if dm.Get(i, j) != 0.0 {
				t.Errorf("Expected 0.0 at (%d, %d), got %f", i, j, dm.Get(i, j))
			}
		}
	}
}

func TestDistMatrixDiagonal(t *testing.T) {
	dm := NewDistMatrix(5)

	// Diagonal should always be 0
	for i := 0; i < 5; i++ {
		if dm.Get(i, i) != 0.0 {
			t.Errorf("Expected diagonal element (%d, %d) to be 0.0, got %f", i, i, dm.Get(i, i))
		}
	}

	// Try to set diagonal (should be ignored)
	dm.Set(2, 2, 5.0)
	if dm.Get(2, 2) != 0.0 {
		t.Errorf("Diagonal should remain 0.0 even after Set, got %f", dm.Get(2, 2))
	}
}

func TestDistMatrixSymmetry(t *testing.T) {
	dm := NewDistMatrix(4)

	dm.Set(0, 1, 1.5)
	dm.Set(0, 2, 2.5)
	dm.Set(1, 3, 3.5)

	// Check symmetry
	if dm.Get(0, 1) != dm.Get(1, 0) {
		t.Errorf("Matrix not symmetric: Get(0,1)=%f, Get(1,0)=%f", dm.Get(0, 1), dm.Get(1, 0))
	}

	if dm.Get(0, 2) != dm.Get(2, 0) {
		t.Errorf("Matrix not symmetric: Get(0,2)=%f, Get(2,0)=%f", dm.Get(0, 2), dm.Get(2, 0))
	}

	if dm.Get(1, 3) != dm.Get(3, 1) {
		t.Errorf("Matrix not symmetric: Get(1,3)=%f, Get(3,1)=%f", dm.Get(1, 3), dm.Get(3, 1))
	}
}

func TestDistMatrixSetGet(t *testing.T) {
	dm := NewDistMatrix(4)

	testCases := []struct {
		i     int
		j     int
		value float64
	}{
		{0, 1, 1.5},
		{0, 2, 2.5},
		{0, 3, 3.5},
		{1, 2, 4.5},
		{1, 3, 5.5},
		{2, 3, 6.5},
	}

	for _, tc := range testCases {
		dm.Set(tc.i, tc.j, tc.value)
	}

	for _, tc := range testCases {
		got := dm.Get(tc.i, tc.j)
		if math.Abs(got-tc.value) > 1e-10 {
			t.Errorf("Get(%d, %d): expected %f, got %f", tc.i, tc.j, tc.value, got)
		}

		// Check symmetry
		got = dm.Get(tc.j, tc.i)
		if math.Abs(got-tc.value) > 1e-10 {
			t.Errorf("Get(%d, %d) (symmetric): expected %f, got %f", tc.j, tc.i, tc.value, got)
		}
	}
}

func TestDistMatrixLabels(t *testing.T) {
	labels := []string{"A", "B", "C", "D"}
	dm := NewDistMatrixWithLabels(labels)

	if dm.Size() != 4 {
		t.Errorf("Expected size 4, got %d", dm.Size())
	}

	for i, label := range labels {
		if dm.GetLabel(i) != label {
			t.Errorf("Expected label %s at index %d, got %s", label, i, dm.GetLabel(i))
		}
	}

	// Modify a label
	dm.SetLabel(1, "Modified")
	if dm.GetLabel(1) != "Modified" {
		t.Errorf("Expected label 'Modified' at index 1, got %s", dm.GetLabel(1))
	}

	// Check Labels() returns a copy
	labelsCopy := dm.Labels()
	labelsCopy[0] = "ChangedCopy"
	if dm.GetLabel(0) != "A" {
		t.Errorf("Modifying Labels() return value should not affect original labels")
	}
}

func TestDistMatrixMinDistance(t *testing.T) {
	dm := NewDistMatrix(4)

	dm.Set(0, 1, 2.5)
	dm.Set(0, 2, 1.5) // minimum
	dm.Set(0, 3, 3.5)
	dm.Set(1, 2, 4.5)
	dm.Set(1, 3, 5.5)
	dm.Set(2, 3, 6.5)

	minDist, minI, minJ := dm.MinDistance()

	if math.Abs(minDist-1.5) > 1e-10 {
		t.Errorf("Expected min distance 1.5, got %f", minDist)
	}

	if (minI != 0 || minJ != 2) && (minI != 2 || minJ != 0) {
		t.Errorf("Expected min at (0, 2) or (2, 0), got (%d, %d)", minI, minJ)
	}
}

func TestDistMatrixMaxDistance(t *testing.T) {
	dm := NewDistMatrix(4)

	dm.Set(0, 1, 2.5)
	dm.Set(0, 2, 1.5)
	dm.Set(0, 3, 3.5)
	dm.Set(1, 2, 4.5)
	dm.Set(1, 3, 5.5)
	dm.Set(2, 3, 6.5) // maximum

	maxDist, maxI, maxJ := dm.MaxDistance()

	if math.Abs(maxDist-6.5) > 1e-10 {
		t.Errorf("Expected max distance 6.5, got %f", maxDist)
	}

	if (maxI != 2 || maxJ != 3) && (maxI != 3 || maxJ != 2) {
		t.Errorf("Expected max at (2, 3) or (3, 2), got (%d, %d)", maxI, maxJ)
	}
}

func TestDistMatrixGetRow(t *testing.T) {
	dm := NewDistMatrix(3)

	dm.Set(0, 1, 1.0)
	dm.Set(0, 2, 2.0)
	dm.Set(1, 2, 3.0)

	row0 := dm.GetRow(0)
	expected0 := []float64{0.0, 1.0, 2.0}

	for i, val := range expected0 {
		if math.Abs(row0[i]-val) > 1e-10 {
			t.Errorf("Row 0[%d]: expected %f, got %f", i, val, row0[i])
		}
	}

	row1 := dm.GetRow(1)
	expected1 := []float64{1.0, 0.0, 3.0}

	for i, val := range expected1 {
		if math.Abs(row1[i]-val) > 1e-10 {
			t.Errorf("Row 1[%d]: expected %f, got %f", i, val, row1[i])
		}
	}
}

func TestDistMatrixCopy(t *testing.T) {
	dm := NewDistMatrixWithLabels([]string{"A", "B", "C"})
	dm.Set(0, 1, 1.5)
	dm.Set(0, 2, 2.5)
	dm.Set(1, 2, 3.5)

	dmCopy := dm.Copy()

	// Check values are copied
	if dmCopy.Get(0, 1) != dm.Get(0, 1) {
		t.Errorf("Copy has different value at (0, 1)")
	}

	// Check labels are copied
	if dmCopy.GetLabel(0) != dm.GetLabel(0) {
		t.Errorf("Copy has different label at index 0")
	}

	// Modify copy and ensure original unchanged
	dmCopy.Set(0, 1, 99.9)
	if dm.Get(0, 1) == 99.9 {
		t.Errorf("Modifying copy affected original matrix")
	}

	dmCopy.SetLabel(0, "Modified")
	if dm.GetLabel(0) == "Modified" {
		t.Errorf("Modifying copy label affected original matrix")
	}
}

func TestDistMatrixToFullMatrix(t *testing.T) {
	dm := NewDistMatrix(3)
	dm.Set(0, 1, 1.0)
	dm.Set(0, 2, 2.0)
	dm.Set(1, 2, 3.0)

	full := dm.ToFullMatrix()

	expected := [][]float64{
		{0.0, 1.0, 2.0},
		{1.0, 0.0, 3.0},
		{2.0, 3.0, 0.0},
	}

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if math.Abs(full[i][j]-expected[i][j]) > 1e-10 {
				t.Errorf("Full matrix[%d][%d]: expected %f, got %f",
					i, j, expected[i][j], full[i][j])
			}
		}
	}
}

func TestDistMatrixBoundsChecking(t *testing.T) {
	dm := NewDistMatrix(3)

	// Test Get out of bounds
	testPanic := func(f func()) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic, but didn't get one")
			}
		}()
		f()
	}

	testPanic(func() { dm.Get(-1, 0) })
	testPanic(func() { dm.Get(0, 3) })
	testPanic(func() { dm.Set(3, 0, 1.0) })
	testPanic(func() { dm.GetLabel(-1) })
	testPanic(func() { dm.SetLabel(3, "Invalid") })
	testPanic(func() { dm.GetRow(3) })
}

func TestDistMatrixEmptyMatrix(t *testing.T) {
	dm := NewDistMatrix(0)

	if dm.Size() != 0 {
		t.Errorf("Expected size 0, got %d", dm.Size())
	}

	minDist, minI, minJ := dm.MinDistance()
	if minDist != 0.0 || minI != -1 || minJ != -1 {
		t.Errorf("Empty matrix MinDistance should return (0.0, -1, -1), got (%f, %d, %d)",
			minDist, minI, minJ)
	}

	maxDist, maxI, maxJ := dm.MaxDistance()
	if maxDist != 0.0 || maxI != -1 || maxJ != -1 {
		t.Errorf("Empty matrix MaxDistance should return (0.0, -1, -1), got (%f, %d, %d)",
			maxDist, maxI, maxJ)
	}
}

func TestDistMatrixSingleElement(t *testing.T) {
	dm := NewDistMatrix(1)

	if dm.Size() != 1 {
		t.Errorf("Expected size 1, got %d", dm.Size())
	}

	// Only element is diagonal (always 0)
	if dm.Get(0, 0) != 0.0 {
		t.Errorf("Expected 0.0 at (0, 0), got %f", dm.Get(0, 0))
	}

	minDist, minI, minJ := dm.MinDistance()
	if minDist != 0.0 || minI != -1 || minJ != -1 {
		t.Errorf("Single element matrix MinDistance should return (0.0, -1, -1), got (%f, %d, %d)",
			minDist, minI, minJ)
	}
}

func TestNewSimilarityMatrix(t *testing.T) {
	sm := NewSimilarityMatrix(4)

	if sm.Size() != 4 {
		t.Errorf("Expected size 4, got %d", sm.Size())
	}

	// Check diagonal is 1.0
	for i := 0; i < 4; i++ {
		if sm.Get(i, i) != 1.0 {
			t.Errorf("Expected diagonal element (%d, %d) to be 1.0, got %f", i, i, sm.Get(i, i))
		}
	}

	// Check off-diagonal is 0.0
	if sm.Get(0, 1) != 0.0 {
		t.Errorf("Expected off-diagonal to be 0.0, got %f", sm.Get(0, 1))
	}
}

func TestNewSimilarityMatrixWithLabels(t *testing.T) {
	labels := []string{"A", "B", "C"}
	sm := NewSimilarityMatrixWithLabels(labels)

	if sm.Size() != 3 {
		t.Errorf("Expected size 3, got %d", sm.Size())
	}

	// Check labels
	for i, label := range labels {
		if sm.GetLabel(i) != label {
			t.Errorf("Expected label %s at index %d, got %s", label, i, sm.GetLabel(i))
		}
	}

	// Check diagonal is 1.0
	for i := 0; i < 3; i++ {
		if sm.Get(i, i) != 1.0 {
			t.Errorf("Expected diagonal element (%d, %d) to be 1.0, got %f", i, i, sm.Get(i, i))
		}
	}

	// Set some similarities
	sm.Set(0, 1, 0.8)
	sm.Set(0, 2, 0.6)
	sm.Set(1, 2, 0.7)

	// Check values
	if math.Abs(sm.Get(0, 1)-0.8) > 1e-10 {
		t.Errorf("Expected 0.8 at (0, 1), got %f", sm.Get(0, 1))
	}

	if math.Abs(sm.Get(1, 0)-0.8) > 1e-10 {
		t.Errorf("Expected 0.8 at (1, 0) (symmetry), got %f", sm.Get(1, 0))
	}
}

func TestSimilarityMatrixCopy(t *testing.T) {
	sm := NewSimilarityMatrix(3)
	sm.Set(0, 1, 0.9)
	sm.Set(0, 2, 0.7)

	smCopy := sm.Copy()

	// Check diagonal is preserved
	if smCopy.Get(0, 0) != 1.0 {
		t.Errorf("Copied similarity matrix should have diagonal 1.0, got %f", smCopy.Get(0, 0))
	}

	// Check values are preserved
	if math.Abs(smCopy.Get(0, 1)-0.9) > 1e-10 {
		t.Errorf("Copy should preserve values, expected 0.9, got %f", smCopy.Get(0, 1))
	}

	// Modify copy and ensure original unchanged
	smCopy.Set(0, 1, 0.5)
	if math.Abs(sm.Get(0, 1)-0.9) > 1e-10 {
		t.Errorf("Modifying copy should not affect original, expected 0.9, got %f", sm.Get(0, 1))
	}
}
