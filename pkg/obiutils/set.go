package obiutils

import (
	"fmt"
	"reflect"
	"slices"
)

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

// compareMembers orders two set members: numbers by value, strings
// alphabetically, anything else by its printed form. It is only used to give
// the string representation of a set a stable order.
func compareMembers(a, b any) int {
	va, vb := reflect.ValueOf(a), reflect.ValueOf(b)
	switch {
	case va.CanInt() && vb.CanInt():
		return cmpOrdered(va.Int(), vb.Int())
	case va.CanUint() && vb.CanUint():
		return cmpOrdered(va.Uint(), vb.Uint())
	case va.CanFloat() && vb.CanFloat():
		return cmpOrdered(va.Float(), vb.Float())
	case va.Kind() == reflect.String && vb.Kind() == reflect.String:
		return cmpOrdered(va.String(), vb.String())
	}
	return cmpOrdered(fmt.Sprint(a), fmt.Sprint(b))
}

func cmpOrdered[T int64 | uint64 | float64 | string](a, b T) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

// String returns a string representation of the set.
//
// The members are formatted with fmt, in increasing order: a Go map has no
// order, and a representation changing from one call to the next is of no use
// to print, log or compare.
func (s Set[E]) String() string {
	members := s.Members()
	slices.SortFunc(members, func(a, b E) int { return compareMembers(a, b) })
	return fmt.Sprintf("%v", members)
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
