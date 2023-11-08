package obiseq

import (
	"reflect"
	"testing"
)

// TestComplement is a test function that tests the complement function.
//
// It tests the complement function by providing a set of input bytes and their
// expected output bytes. It verifies that the complement function correctly
// computes the complement of each input byte.
//
// Parameters:
//   - t: *testing.T - the testing object for running test cases and reporting
//     failures.
//
// Returns: None.
func TestNucComplement(t *testing.T) {
	tests := []struct {
		input byte
		want  byte
	}{
		{'.', '.'},
		{'-', '-'},
		{'a', 't'},
		{'G', 'c'},
		{'T', 'a'},
		{'C', 'g'},
		{'n', 'n'},
		{'[', ']'},
		{']', '['},
	}

	for _, tc := range tests {
		got := nucComplement(tc.input)
		if got != tc.want {
			t.Errorf("complement(%c) = %c, want %c", tc.input, got, tc.want)
		}
	}
}

// TestReverseComplement is a test function for the ReverseComplement method.
//
// It tests the behavior of the ReverseComplement method under different scenarios.
// The function checks if the ReverseComplement method returns the expected result
// when the 'inplace' parameter is set to false or true. It also verifies if the
// ReverseComplement method correctly handles BioSequences with qualities.
// The function uses the NewBioSequence and NewBioSequenceWithQualities functions
// to create BioSequence objects with different sequences and qualities.
// It compares the result of the ReverseComplement method with the expected result
// and reports an error if they are not equal. Additionally, it compares the
// qualities of the result BioSequence with the expected qualities and reports
// an error if they are not equal.
func TestReverseComplement(t *testing.T) {
	// Test when inplace is false
	seq := NewBioSequence("123", []byte("ATCG"), "")
	expected := NewBioSequence("123", []byte("CGAT"), "")
	result := seq.ReverseComplement(false)
	if result.String() != expected.String() {
		t.Errorf("Expected %v, but got %v", expected, result)
	}

	// Test when inplace is true
	seq = NewBioSequence("123", []byte("ATCG"), "")
	expected = NewBioSequence("123", []byte("CGAT"), "")
	result = seq.ReverseComplement(true)
	if result.String() != expected.String() {
		t.Errorf("Expected %v, but got %v", expected, result)
	}

	// Test when BioSequence has qualities
	seq = NewBioSequenceWithQualities("123", []byte("ATCG"), "", []byte{30, 20, 10, 40})
	expected = NewBioSequenceWithQualities("123", []byte("CGAT"), "", []byte{40, 10, 20, 30})
	result = seq.ReverseComplement(false)
	if result.String() != expected.String() {
		t.Errorf("Expected %v, but got %v", expected, result)
	}
	if !reflect.DeepEqual(result.Qualities(), expected.Qualities()) {
		t.Errorf("Expected %v, but got %v", expected.Qualities(), result.Qualities())
	}

	// Test when BioSequence has qualities and inplace is true
	seq = NewBioSequenceWithQualities("123", []byte("ATCG"), "", []byte{30, 20, 10, 40})
	expected = NewBioSequenceWithQualities("123", []byte("CGAT"), "", []byte{40, 10, 20, 30})
	result = seq.ReverseComplement(true)
	if result.String() != expected.String() {
		t.Errorf("Expected %v, but got %v", expected, result)
	}
	if !reflect.DeepEqual(result.Qualities(), expected.Qualities()) {
		t.Errorf("Expected %v, but got %v", expected.Qualities(), result.Qualities())
	}
}
