package obiutils

// Contains checks if the given element is present in the given array.
//
// Parameters:
// - arr: The array to search in.
// - x: The element to search for.
//
// Return type:
// - bool: Returns true if the element is found, false otherwise.
func Contains[T comparable](arr []T, x T) bool {
	for _, v := range arr {
		if v == x {
			return true
		}
	}
	return false
}

// LookFor searches for the first occurrence of a given element in an array and returns its index.
//
// Parameters:
// - arr: the array to be searched
// - x: the element to search for
//
// Return:
// - int: the index of the first occurrence of the element in the array, or -1 if not found
func LookFor[T comparable](arr []T, x T) int {
	for i, v := range arr {
		if v == x {
			return i
		}
	}
	return -1
}

// RemoveIndex removes an element at a specified index from a slice.
//
// Parameters:
// - s: The slice from which the element will be removed.
// - index: The index of the element to be removed.
//
// Returns:
// A new slice with the element removed.
func RemoveIndex[T comparable](s []T, index int) []T {
	return append(s[:index], s[index+1:]...)
}

// Reverse reverses the elements of a slice.
//
// The function takes a slice `s` and a boolean `inplace` parameter. If `inplace`
// is `true`, the function modifies the input slice directly. If `inplace` is
// `false`, the function creates a new slice `c` and copies the elements of `s`
// into `c`. The function then reverses the elements of `s` in-place or `c`
// depending on the `inplace` parameter.
//
// The function returns the reversed slice.
func Reverse[S ~[]E, E any](s S, inplace bool) S {
	if !inplace {
		c := make([]E, len(s))
		copy(c, s)
		s = c
	}
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

	return s
}
