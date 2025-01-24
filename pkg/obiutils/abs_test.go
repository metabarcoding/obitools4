package obiutils_test

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	"testing"
)

func TestAbs(t *testing.T) {
	// Test cases for positive integers
	testCasesPositive := []struct {
		input    int
		expected int
	}{
		{0, 0},
		{1, 1},
		{5, 5},
		{10, 10},
	}

	for _, tc := range testCasesPositive {
		result := obiutils.Abs(tc.input)
		if result != tc.expected {
			t.Errorf("Abs(%d) = %d; want %d", tc.input, result, tc.expected)
		}
	}

	// Test cases for negative integers
	testCasesNegative := []struct {
		input    int
		expected int
	}{
		{-1, 1},
		{-5, 5},
		{-10, 10},
	}

	for _, tc := range testCasesNegative {
		result := obiutils.Abs(tc.input)
		if result != tc.expected {
			t.Errorf("Abs(%d) = %d; want %d", tc.input, result, tc.expected)
		}
	}
}
