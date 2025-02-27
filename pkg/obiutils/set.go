package obiutils

import "fmt"

type Set[E comparable] map[E]struct{}

// MakeSet creates a new Set with the provided values.
//
// The function takes a variadic parameter `vals` of type `E` which represents the values
// that will be added to the Set.
//
// The function returns a Set of type `Set[E]` which contains the provided values.
func MakeSet[E comparable](vals ...E) Set[E] {
	s := Set[E]{}
	for _, v := range vals {
		s[v] = struct{}{}
	}
	return s
}

// NewSet creates a new Set with the given values.
//
// It takes a variadic parameter of type E, where E is a comparable type.
// It returns a pointer to a Set of type E.
func NewSet[E comparable](vals ...E) *Set[E] {
	s := MakeSet(vals...)
	return &s
}

// Add adds the given values to the set.
//
// It takes a variadic parameter `vals` of type `E`.
// There is no return type for this function.
func (s Set[E]) Add(vals ...E) {
	for _, v := range vals {
		s[v] = struct{}{}
	}
}

// Contains checks if the set contains a given element.
//
// Parameters:
// - v: the element to check for presence in the set.
//
// Returns:
// - bool: true if the set contains the given element, false otherwise.
func (s Set[E]) Contains(v E) bool {
	_, ok := s[v]
	return ok
}

// Members returns a slice of all the elements in the set.
//
// It does not modify the original set.
// It returns a slice of type []E.
func (s Set[E]) Members() []E {
	result := make([]E, 0, len(s))
	for v := range s {
		result = append(result, v)
	}
	return result
}

// String returns a string representation of the set.
//
// It returns a string representation of the set by formatting the set's members using the fmt.Sprintf function.
// The resulting string is then returned.
func (s Set[E]) String() string {
	return fmt.Sprintf("%v", s.Members())
}

// Union returns a new set that is the union of the current set and the specified set.
//
// Parameters:
// - s2: the set to be unioned with the current set.
//
// Return:
// - Set[E]: the resulting set after the union operation.
func (s Set[E]) Union(s2 Set[E]) Set[E] {
	result := MakeSet(s.Members()...)
	result.Add(s2.Members()...)
	return result
}

// Intersection returns a new set that contains the common elements between the current set and another set.
//
// Parameter:
// - s2: the other set to compare with.
// Return:
// - Set[E]: a new set that contains the common elements.
func (s Set[E]) Intersection(s2 Set[E]) Set[E] {
	result := MakeSet[E]()
	for _, v := range s.Members() {
		if s2.Contains(v) {
			result.Add(v)
		}
	}
	return result
}
