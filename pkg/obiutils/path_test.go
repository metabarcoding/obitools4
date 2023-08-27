package obiutils

import (
	"testing"
)

// TestRemoveAllExt is a test function for the RemoveAllExt function.
//
// It tests the RemoveAllExt function by providing different test cases
// with different file paths and expected results. It ensures that the
// RemoveAllExt function correctly removes all file extensions and returns
// the file path without any extensions.
//
// Parameter(s):
// - t: A testing.T value representing the test framework.
//
// Return type(s): None.
func TestRemoveAllExt(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "No extensions",
			path:     "path/to/file",
			expected: "path/to/file",
		},
		{
			name:     "Single extension",
			path:     "path/to/file.txt",
			expected: "path/to/file",
		},
		{
			name:     "Multiple extensions",
			path:     "path/to/file.tar.gz",
			expected: "path/to/file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveAllExt(tt.path)
			if result != tt.expected {
				t.Errorf("Expected %q, but got %q", tt.expected, result)
			}
		})
	}
}
