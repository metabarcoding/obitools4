package obiutils

import (
	"reflect"
	"testing"
)

func TestMake2DArray(t *testing.T) {
	// Testing with rows = 0 and cols = 0
	matrix := Make2DArray[int](0, 0)
	if len(matrix) != 0 {
		t.Errorf("Expected length of matrix to be 0, but got %d", len(matrix))
	}

	// Testing with rows = 3 and cols = 4
	matrix = Make2DArray[int](3, 4)
	if len(matrix) != 3 {
		t.Errorf("Expected length of matrix to be 3, but got %d", len(matrix))
	}
	for i := 0; i < len(matrix); i++ {
		if len(matrix[i]) != 4 {
			t.Errorf("Expected length of row %d to be 4, but got %d", i, len(matrix[i]))
		}
	}

	// Testing with rows = 2 and cols = 2 using string type
	stringMatrix := Make2DArray[string](2, 2)
	expectedStringMatrix := Matrix[string]{{"", ""}, {"", ""}}
	if !reflect.DeepEqual(stringMatrix, expectedStringMatrix) {
		t.Errorf("Expected matrix to be %v, but got %v", expectedStringMatrix, stringMatrix)
	}

	// Testing with rows = 4 and cols = 2 using custom struct type
	type Person struct {
		Name string
		Age  int
	}
	personMatrix := Make2DArray[Person](4, 2)
	expectedPersonMatrix := Matrix[Person]{
		{{Name: "", Age: 0}, {Name: "", Age: 0}},
		{{Name: "", Age: 0}, {Name: "", Age: 0}},
		{{Name: "", Age: 0}, {Name: "", Age: 0}},
		{{Name: "", Age: 0}, {Name: "", Age: 0}},
	}
	if !reflect.DeepEqual(personMatrix, expectedPersonMatrix) {
		t.Errorf("Expected matrix to be %v, but got %v", expectedPersonMatrix, personMatrix)
	}

}

// TestMatrix_Column tests the Column method of the Matrix struct.
//
// Test case 1: Retrieving the first column of a 3x3 matrix of integers.
// Parameter(s): None.
// Return type(s): []int.
//
// Test case 2: Retrieving the second column of a 2x4 matrix of strings.
// Parameter(s): None.
// Return type(s): []string.
//
// Test case 3: Retrieving the third column of a 4x2 matrix of custom struct type.
// Parameter(s): None.
// Return type(s): []Person.
func TestMatrix_Column(t *testing.T) {

	type Person struct {
		Name string
		Age  int
	}

	// Test case 1: Retrieving the first column of a 3x3 matrix of integers
	intMatrix := Matrix[int]{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	expectedInt := []int{1, 4, 7}
	resultInt := intMatrix.Column(0)
	if !reflect.DeepEqual(resultInt, expectedInt) {
		t.Errorf("Expected %v, but got %v", expectedInt, resultInt)
	}

	// Test case 2: Retrieving the second column of a 2x4 matrix of strings
	stringMatrix := Matrix[string]{{"a", "b", "c", "d"}, {"e", "f", "g", "h"}}
	expectedString := []string{"b", "f"}
	resultString := stringMatrix.Column(1)
	if !reflect.DeepEqual(resultString, expectedString) {
		t.Errorf("Expected %v, but got %v", expectedString, resultString)
	}

	// Test case 3: Retrieving the third column of a 4x2 matrix of custom struct type
	personMatrix := Matrix[Person]{
		{{Name: "Alice", Age: 25}, {Name: "Bob", Age: 30}},
		{{Name: "Charlie", Age: 35}, {Name: "Dave", Age: 40}},
		{{Name: "Eve", Age: 45}, {Name: "Frank", Age: 50}},
		{{Name: "Grace", Age: 55}, {Name: "Henry", Age: 60}},
	}
	expectedPerson := []Person{{Name: "Bob", Age: 30}, {Name: "Dave", Age: 40}, {Name: "Frank", Age: 50}, {Name: "Henry", Age: 60}}
	resultPerson := personMatrix.Column(1)
	if !reflect.DeepEqual(resultPerson, expectedPerson) {
		t.Errorf("Expected %v, but got %v", expectedPerson, resultPerson)
	}
}

// TestRows is a unit test function that tests the Rows method of the Matrix type in the Go program.
//
// It tests the behavior of the Rows method by providing different test cases, including single row,
// multiple rows, and empty rows, and comparing the actual output with the expected output.
//
// Parameters:
//   - t: The testing.T type object for running the test cases and reporting any failures.
//
// Return: None.
func TestRows(t *testing.T) {
	matrix := Matrix[int]{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}

	testCases := []struct {
		name     string
		args     []int
		expected Matrix[int]
	}{
		{
			name:     "Single Row",
			args:     []int{1},
			expected: Matrix[int]{{4, 5, 6}},
		},
		{
			name:     "Multiple Rows",
			args:     []int{0, 2},
			expected: Matrix[int]{{1, 2, 3}, {7, 8, 9}},
		},
		{
			name:     "Empty Rows",
			args:     []int{},
			expected: Matrix[int]{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := matrix.Rows(tc.args...)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("got %v, want %v", got, tc.expected)
			}
		})
	}
}

// TestDim is a function that tests the `Dim` method of the `Matrix` struct.
//
// It tests the behavior of the `Dim` method using multiple test cases:
//  1. Test case: nil matrix
//  2. Test case: empty matrix
//  3. Test case: matrix with zero columns
//  4. Test case: non-empty matrix
//  5. Test case: generic matrix
//
// The function checks the dimensions returned by the `Dim` method and
// compares them with the expected dimensions for each test case. If the
// returned dimensions do not match the expected dimensions, it logs an
// error message using the `t.Errorf` function.
//
// Return type: void.
func TestDim(t *testing.T) {
	// Test case: nil matrix
	matrix := (*Matrix[int])(nil)
	rows, cols := matrix.Dim()
	if rows != 0 || cols != 0 {
		t.Errorf("Expected dimensions (0, 0), got (%d, %d)", rows, cols)
	}

	// Test case: empty matrix
	matrix = &Matrix[int]{}
	rows, cols = matrix.Dim()
	if rows != 0 || cols != 0 {
		t.Errorf("Expected dimensions (0, 0), got (%d, %d)", rows, cols)
	}

	// Test case: matrix with zero columns
	matrix = &Matrix[int]{{}}
	rows, cols = matrix.Dim()
	if rows != 1 || cols != 0 {
		t.Errorf("Expected dimensions (1, 0), got (%d, %d)", rows, cols)
	}

	// Test case: non-empty matrix
	matrix = &Matrix[int]{
		{1, 2, 3},
		{4, 5, 6},
	}
	rows, cols = matrix.Dim()
	if rows != 2 || cols != 3 {
		t.Errorf("Expected dimensions (2, 3), got (%d, %d)", rows, cols)
	}

	// Test case: generic matrix
	genericMatrix := reflect.New(reflect.TypeOf(Matrix[int]{})).Interface()
	rows, cols = genericMatrix.(*Matrix[int]).Dim()
	if rows != 0 || cols != 0 {
		t.Errorf("Expected dimensions (0, 0), got (%d, %d)", rows, cols)
	}
}
