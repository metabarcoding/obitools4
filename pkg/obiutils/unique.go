package obiutils

// Unique returns a new slice containing only unique values from the input slice.
// The order of elements in the output slice is not guaranteed to match the input order.
//
// Parameters:
// - slice: The input slice containing potentially duplicate values
//
// Returns:
// - A new slice containing only unique values
func Unique[T comparable](slice []T) []T {
	// Create a map to track unique values
	seen := Set[T]{}

	for _, v := range slice {
		seen.Add(v)
	}

	return seen.Members()
}
