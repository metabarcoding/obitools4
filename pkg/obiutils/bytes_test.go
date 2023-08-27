package obiutils

import (
	"reflect"
	"testing"
)

// TestInPlaceToLower is a test function that tests the InPlaceToLower function.
//
// It tests the behavior of the InPlaceToLower function by providing different inputs and comparing the expected output.
// The function takes a slice of bytes as input and converts all uppercase letters to lowercase in-place.
// It returns nothing.
func TestInPlaceToLower(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "All uppercase letters",
			input:    []byte("HELLO WORLD"),
			expected: []byte("hello world"),
		},
		{
			name:     "Empty input",
			input:    []byte(""),
			expected: []byte(""),
		},
		{
			name:     "No uppercase letters",
			input:    []byte("hello world"),
			expected: []byte("hello world"),
		},
		{
			name:     "Mixed case letters",
			input:    []byte("Hello WoRlD"),
			expected: []byte("hello world"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := InPlaceToLower(test.input)
			if string(result) != string(test.expected) {
				t.Errorf("Expected %s, but got %s", test.expected, result)
			}
		})
	}
}

// TestMake2DNumericArray is a Go function that tests the Make2DNumericArray function.
//
// Parameter(s):
// - t: a pointer to the testing.T type
//
// Return type(s): None
func TestMake2DNumericArray(t *testing.T) {
	// Test case 1: Create a 2D numeric array with 3 rows and 4 columns, not zeroed
	matrix := Make2DNumericArray[int](3, 4, false)
	expected := Matrix[int]{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}}
	if !reflect.DeepEqual(matrix, expected) {
		t.Errorf("Test case 1 failed. Expected %v, but got %v", expected, matrix)
	}

	// Test case 2: Create a 2D numeric array with 2 rows and 2 columns, zeroed
	matrix = Make2DNumericArray[int](2, 2, true)
	expected = Matrix[int]{{0, 0}, {0, 0}}
	if !reflect.DeepEqual(matrix, expected) {
		t.Errorf("Test case 2 failed. Expected %v, but got %v", expected, matrix)
	}

	// Test case 3: Create a 2D numeric array with 1 row and 1 column, zeroed
	matrix = Make2DNumericArray[int](1, 1, true)
	expected = Matrix[int]{{0}}
	if !reflect.DeepEqual(matrix, expected) {
		t.Errorf("Test case 3 failed. Expected %v, but got %v", expected, matrix)
	}
}
