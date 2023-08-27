package obiapat

import (
	"testing"
)

func TestMakeApatPattern(t *testing.T) {
	// Test case 1: pattern with no errors allowed
	pattern1 := "ACGT"
	errormax1 := 0
	allowsIndel1 := false
	actual1, err1 := MakeApatPattern(pattern1, errormax1, allowsIndel1)
	if err1 != nil {
		t.Errorf("Error in test case 1: %v", err1)
	}
	if actual1.pointer == nil {
		t.Errorf("Incorrect result in test case 1. Expected a non-nil ApatPattern pointer, but got nil")
	}

	// Test case 2: pattern with errors allowed and indels allowed
	pattern2 := "A[T]C!GT"
	errormax2 := 2
	allowsIndel2 := true
	actual2, err2 := MakeApatPattern(pattern2, errormax2, allowsIndel2)
	if err2 != nil {
		t.Errorf("Error in test case 2: %v", err2)
	}
	if actual2.pointer == nil {
		t.Errorf("Incorrect result in test case 2. Expected a non-nil ApatPattern pointer, but got nil")
	}

	// Test case 3: pattern with errors allowed and indels not allowed
	pattern3 := "A[T]C!GT"
	errormax3 := 2
	allowsIndel3 := false
	actual3, err3 := MakeApatPattern(pattern3, errormax3, allowsIndel3)
	if err3 != nil {
		t.Errorf("Error in test case 3: %v", err3)
	}
	if actual3.pointer == nil {
		t.Errorf("Incorrect result in test case 3. Expected a non-nil ApatPattern pointer, but got nil")
	}
}
