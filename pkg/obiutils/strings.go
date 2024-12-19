package obiutils

import (
	"fmt"
	"unsafe"
)

type AsciiSet [256]bool

var AsciiSpaceSet = AsciiSetFromString("\t\n\v\f\r ")
var AsciiDigitSet = AsciiSetFromString("0123456789")
var AsciiUpperSet = AsciiSetFromString("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
var AsciiLowerSet = AsciiSetFromString("abcdefghijklmnopqrstuvwxyz")
var AsciiAlphaSet = AsciiUpperSet.Union(AsciiLowerSet)
var AsciiAlphaNumSet = AsciiAlphaSet.Union(AsciiDigitSet)

// UnsafeStringFromBytes converts a byte slice into a string without making a copy of the data.
// This function is considered unsafe because it directly manipulates memory and does not
// perform any checks on the byte slice's contents. It should be used with caution.
//
// Parameters:
//   - data: A byte slice that contains the data to be converted to a string.
//
// Returns:
//
//	A string representation of the byte slice. If the input byte slice is empty,
//	an empty string is returned.
func UnsafeStringFromBytes(data []byte) string {
	if len(data) > 0 {
		// Convert the byte slice to a string using unsafe operations.
		s := unsafe.String(unsafe.SliceData(data), len(data))
		return s
	}

	return "" // Return an empty string if the input slice is empty.
}

func AsciiSetFromString(s string) AsciiSet {
	r := [256]bool{}
	for _, c := range s {
		r[c] = true
	}
	return r
}

func (r *AsciiSet) Contains(c byte) bool {
	return r[c]
}

func (r *AsciiSet) Union(s AsciiSet) AsciiSet {
	for i := 0; i < 256; i++ {
		s[i] = r[i] || s[i]
	}

	return s
}

func (r *AsciiSet) Intersect(s AsciiSet) AsciiSet {
	for i := 0; i < 256; i++ {
		s[i] = r[i] && s[i]
	}

	return s
}

// FirstWord extracts the first word from a given string.
// A word is defined as a sequence of non-space characters.
// It ignores leading whitespace and stops at the first whitespace character encountered.
//
// Parameters:
//   - s: The input string from which the first word is to be extracted.
//
// Returns:
//
//	A string containing the first word found in the input string. If the input string
//	is empty or contains only whitespace, an empty string is returned.
func FirstWord(s string) string {
	// Fast path for ASCII: look for the first ASCII non-space byte
	start := 0
	for ; start < len(s); start++ {
		c := s[start]
		if !AsciiSpaceSet.Contains(c) {
			break
		}
	}

	stop := start
	for ; stop < len(s); stop++ {
		c := s[stop]
		if AsciiSpaceSet.Contains(c) {
			break
		}
	}
	return s[start:stop]
}

// FirstRestrictedWord extracts the first word from a given string while enforcing character restrictions.
// A word is defined as a sequence of non-space characters. The function checks each character
// against the provided restriction array, which indicates whether a character is allowed.
//
// Parameters:
//   - s: The input string from which the first restricted word is to be extracted.
//   - restriction: A boolean array of size 256 where each index represents a character's ASCII value.
//     If restriction[c] is false, the character c is not allowed in the word.
//
// Returns:
//
//	A string containing the first word found in the input string that does not contain any restricted characters.
//	If a restricted character is found, an error is returned indicating the invalid character.
//	If the input string is empty or contains only whitespace, an empty string is returned with no error.
func (restriction *AsciiSet) FirstWord(s string) (string, error) {
	// Fast path for ASCII: look for the first ASCII non-space byte
	start := 0
	for ; start < len(s); start++ {
		c := s[start]
		if !AsciiSpaceSet.Contains(c) {
			break
		}
	}

	stop := start
	for ; stop < len(s); stop++ {
		c := s[stop]
		if AsciiSpaceSet.Contains(c) {
			break
		}
		if !restriction.Contains(c) {
			return "", fmt.Errorf("invalid character '%c' in string: %s", c, s)
		}
	}

	return s[start:stop], nil
}

func (r *AsciiSet) TrimLeft(s string) string {
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if !AsciiSpaceSet.Contains(c) {
			break
		}
	}
	return s[i:]
}

func SplitInTwo(s string, sep byte) (string, string) {
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c == sep {
			break
		}
	}
	if i == len(s) {
		return s, ""
	}
	return s[:i], s[i+1:]
}
