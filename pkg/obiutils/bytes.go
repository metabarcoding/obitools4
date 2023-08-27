package obiutils

// InPlaceToLower converts all uppercase letters in the input byte slice to lowercase in place.
//
// It takes a single parameter:
//   - data: a byte slice representing the input data
//
// It returns the modified byte slice.
func InPlaceToLower(data []byte) []byte {
	for i, l := range data {
		if l >= 'A' && l <= 'Z' {
			data[i] |= 32
		}
	}

	return data
}
