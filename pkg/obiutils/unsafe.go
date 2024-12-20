package obiutils

import "unsafe"

// UnsafeBytes converts a string into a byte slice without making a copy of the data.
// This function is considered unsafe because it directly manipulates memory and does not
// perform any checks on the string's contents. It should be used with caution.
//
// Parameters:
//   - str: The input string to be converted into a byte slice.
//
// Returns:
//
//	A byte slice representation of the input string. The returned slice shares the same
//	underlying data as the original string, so modifications to the byte slice may affect
//	the original string and vice versa.
func UnsafeBytes(str string) []byte {
	d := unsafe.StringData(str)
	b := unsafe.Slice(d, len(str))

	return b
}

func UnsafeString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
