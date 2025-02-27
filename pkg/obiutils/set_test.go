package obiutils

import (
	"reflect"
	"sort"
	"testing"
)

// TestMakeSet tests the MakeSet function.
//
// The function is tested with an empty set, integer values, and string values.
// It checks if the returned set matches the expected set for each case.
// If any test fails, an error message is printed with the expected and actual sets.
func TestMakeSet(t *testing.T) {
	// Testing with an empty set
	set := MakeSet[int]()
	if len(set) != 0 {
		t.Errorf("Expected an empty set, but got a set with length %d", len(set))
	}

	// Testing with integer values
	intSet := MakeSet(1, 2, 3)
	expectedIntSet := Set[int]{1: {}, 2: {}, 3: {}}
	if !reflect.DeepEqual(intSet, expectedIntSet) {
		t.Errorf("Expected set %v, but got %v", expectedIntSet, intSet)
	}

	// Testing with string values
	strSet := MakeSet("apple", "banana", "orange")
	expectedStrSet := Set[string]{"apple": {}, "banana": {}, "orange": {}}
	if !reflect.DeepEqual(strSet, expectedStrSet) {
		t.Errorf("Expected set %v, but got %v", expectedStrSet, strSet)
	}
}

// TestNewSet tests the NewSet function.
//
// Test Case 1: Creating a set with no values.
// Test Case 2: Creating a set with multiple values.
//
// Parameters:
// - t: *testing.T - the testing object.
//
// Return type: void.
func TestNewSet(t *testing.T) {
	// Test Case 1: Creating a set with no values
	set1 := NewSet[int]()
	if len(*set1) != 0 {
		t.Errorf("Expected size to be 0, but got %d", len(*set1))
	}

	// Test Case 2: Creating a set with multiple values
	set2 := NewSet("apple", "banana", "cherry")
	if len(*set2) != 3 {
		t.Errorf("Expected size to be 3, but got %d", len(*set2))
	}
	if !set2.Contains("apple") {
		t.Errorf("Expected set to contain 'apple', but it didn't")
	}
	if !set2.Contains("banana") {
		t.Errorf("Expected set to contain 'banana', but it didn't")
	}
	if !set2.Contains("cherry") {
		t.Errorf("Expected set to contain 'cherry', but it didn't")
	}
}

// TestSet_Add tests the Add method of the Set type.
//
// It verifies that the Add method properly adds a single value to the set
// and that it correctly adds multiple values as well.
// The function takes a testing.T parameter for error reporting.
func TestSet_Add(t *testing.T) {
	// Test adding a single value
	s := MakeSet[int]()
	s.Add(1)
	if !s.Contains(1) {
		t.Errorf("Expected value 1 to be added to the set")
	}

	// Test adding multiple values
	s.Add(2, 3, 4)
	if !s.Contains(2) {
		t.Errorf("Expected value 2 to be added to the set")
	}
	if !s.Contains(3) {
		t.Errorf("Expected value 3 to be added to the set")
	}
	if !s.Contains(4) {
		t.Errorf("Expected value 4 to be added to the set")
	}
}

// TestSetContains tests the Contains method of the Set type.
//
// It checks whether the element is present in the set or not.
// The function takes a testing.T parameter and does not return any value.
func TestSetContains(t *testing.T) {
	// Test case 1: Element is present in the set
	setInt := NewSet[int]()
	setInt.Add(1)
	setInt.Add(2)
	if !setInt.Contains(1) {
		t.Error("Expected set to contain element 1")
	}

	// Test case 2: Element is not present in the set
	setString := NewSet[string]()
	setString.Add("a")
	setString.Add("b")
	if setString.Contains("c") {
		t.Error("Expected set to not contain element c")
	}
}

// TestMembers tests the Members method of the Set struct in the Go code.
//
// This function includes two test cases:
// 1. Test case 1: Empty set
//   - It creates an empty set using the MakeSet function.
//   - It defines the expected and actual values as empty slices of integers.
//   - It calls the Members method on the set and sorts the returned slice.
//   - It checks if the actual slice is equal to the expected slice using reflect.DeepEqual.
//   - If the actual and expected slices are not equal, it reports an error.
//
// 2. Test case 2: Set with multiple elements
//   - It creates a set with the elements 1, 2, and 3 using the MakeSet function.
//   - It defines the expected and actual values as slices of integers.
//   - It calls the Members method on the set and sorts the returned slice.
//   - It checks if the actual slice is equal to the expected slice using reflect.DeepEqual.
//   - If the actual and expected slices are not equal, it reports an error.
//
// Parameters:
// - t: The testing.T pointer for running the tests and reporting errors.
//
// Return type:
// This function does not return anything.
func TestMembers(t *testing.T) {
	// Test case 1: Empty set
	set := MakeSet[int]()
	expected := []int{}
	actual := set.Members()
	sort.Ints(actual)
	sort.Ints(expected)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Test case 1 failed. Expected %v but got %v", expected, actual)
	}

	// Test case 2: Set with multiple elements
	set = MakeSet(1, 2, 3)
	expected = []int{1, 2, 3}
	actual = set.Members()
	sort.Ints(actual)
	sort.Ints(expected)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Test case 2 failed. Expected %v but got %v", expected, actual)
	}
}

// TestSetString tests the String method of the Set type in Go.
//
// This function checks the string representation of different sets and compares them with the expected values.
// It verifies the correctness of the String method for empty sets, sets with a single member, and sets with multiple members.
// If any of the string representations do not match the expected values, an error message is printed.
func TestSetString(t *testing.T) {
	// Test empty set
	emptySet := NewSet[int]()
	emptySetString := emptySet.String()
	expectedEmptySetString := "[]"
	if emptySetString != expectedEmptySetString {
		t.Errorf("String representation of empty set is incorrect, got: %s, want: %s", emptySetString, expectedEmptySetString)
	}

	// Test set with single member
	singleMemberSet := NewSet(42)
	singleMemberSetString := singleMemberSet.String()
	expectedSingleMemberSetString := "[42]"
	if singleMemberSetString != expectedSingleMemberSetString {
		t.Errorf("String representation of set with single member is incorrect, got: %s, want: %s", singleMemberSetString, expectedSingleMemberSetString)
	}

	// Test set with multiple members
	multipleMembersSet := NewSet(1, 2, 3)
	multipleMembersSetString := multipleMembersSet.String()
	expectedMultipleMembersSetString := "[1 2 3]"
	if multipleMembersSetString != expectedMultipleMembersSetString {
		t.Errorf("String representation of set with multiple members is incorrect, got: %s, want: %s", multipleMembersSetString, expectedMultipleMembersSetString)
	}
}

// TestUnion tests the Union method of the Set struct.
//
// The function checks different test cases for the Union method:
// 1. Union of two empty sets should return an empty set.
// 2. Union of an empty set and a non-empty set should return the non-empty set.
// 3. Union of two non-empty sets with common elements should return a set with unique elements.
// 4. Union of two non-empty sets with no common elements should return a set with all elements.
//
// Parameters:
// - t: The testing.T object for running test cases and reporting failures.
//
// Return type:
// None.
func TestUnion(t *testing.T) {
	// Test case 1: Union of two empty sets should return an empty set
	set1 := MakeSet[int]()
	set2 := MakeSet[int]()
	expected := MakeSet[int]()
	result := set1.Union(set2)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}

	// Test case 2: Union of an empty set and a non-empty set should return the non-empty set
	set1 = MakeSet[int]()
	set2 = MakeSet(1, 2, 3)
	expected = MakeSet(1, 2, 3)
	result = set1.Union(set2)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}

	// Test case 3: Union of two non-empty sets with common elements should return a set with unique elements
	set1 = MakeSet(1, 2, 3)
	set2 = MakeSet(2, 3, 4)
	expected = MakeSet(1, 2, 3, 4)
	result = set1.Union(set2)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}

	// Test case 4: Union of two non-empty sets with no common elements should return a set with all elements
	set1 = MakeSet(1, 2, 3)
	set2 = MakeSet(4, 5, 6)
	expected = MakeSet(1, 2, 3, 4, 5, 6)
	result = set1.Union(set2)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}
}

// TestIntersection is a test function that checks the Intersection method of the Set type.
//
// It tests the following scenarios:
// - Test case 1: Intersection of two empty sets should return an empty set
// - Test case 2: Intersection of an empty set and a non-empty set should return an empty set
// - Test case 3: Intersection of two sets with common elements should return a set with those common elements
// - Test case 4: Intersection of two sets with no common elements should return an empty set
//
// Parameters:
// - t: A testing.T object to report test failures
//
// Return type: None
func TestIntersection(t *testing.T) {
	// Test case 1: Intersection of two empty sets should return an empty set
	emptySet1 := MakeSet[int]()
	emptySet2 := MakeSet[int]()
	expectedResult1 := MakeSet[int]()
	result1 := emptySet1.Intersection(emptySet2)
	if !reflect.DeepEqual(result1, expectedResult1) {
		t.Errorf("Intersection of two empty sets returned incorrect result. Expected: %v, got: %v", expectedResult1, result1)
	}

	// Test case 2: Intersection of an empty set and a non-empty set should return an empty set
	emptySet3 := MakeSet[string]()
	nonEmptySet1 := MakeSet[string]()
	nonEmptySet1.Add("a")
	expectedResult2 := MakeSet[string]()
	result2 := emptySet3.Intersection(nonEmptySet1)
	if !reflect.DeepEqual(result2, expectedResult2) {
		t.Errorf("Intersection of an empty set and a non-empty set returned incorrect result. Expected: %v, got: %v", expectedResult2, result2)
	}

	// Test case 3: Intersection of two sets with common elements should return a set with those common elements
	set1 := MakeSet[int]()
	set1.Add(1)
	set1.Add(2)
	set1.Add(3)
	set2 := MakeSet[int]()
	set2.Add(3)
	set2.Add(4)
	expectedResult3 := MakeSet[int]()
	expectedResult3.Add(3)
	result3 := set1.Intersection(set2)
	if !reflect.DeepEqual(result3, expectedResult3) {
		t.Errorf("Intersection of two sets with common elements returned incorrect result. Expected: %v, got: %v", expectedResult3, result3)
	}

	// Test case 4: Intersection of two sets with no common elements should return an empty set
	set3 := MakeSet[string]()
	set3.Add("a")
	set3.Add("b")
	set3.Add("c")
	set4 := MakeSet[string]()
	set4.Add("x")
	set4.Add("y")
	expectedResult4 := MakeSet[string]()
	result4 := set3.Intersection(set4)
	if !reflect.DeepEqual(result4, expectedResult4) {
		t.Errorf("Intersection of two sets with no common elements returned incorrect result. Expected: %v, got: %v", expectedResult4, result4)
	}
}
