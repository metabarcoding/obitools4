package obiseq

import (
	"bytes"
	"reflect"
	"sync"
	"testing"
)

// TestNewEmptyBioSequence tests the NewEmptyBioSequence function.
//
// It checks the behavior of the function by creating different BioSequence instances with different preallocate values.
// The function verifies that the sequence is correctly preallocated or not preallocated based on the input value.
// It also checks that the length and capacity of the sequence are set correctly.
// The test fails if the function returns nil or if the sequence length or capacity is not as expected.
func TestNewEmptyBioSequence(t *testing.T) {
	// Test case: preallocate is 0, sequence should not be preallocated
	seq := NewEmptyBioSequence(0)
	if seq == nil {
		t.Errorf("NewEmptyBioSequence(0) returned nil")
	} else if len(seq.sequence) != 0 {
		t.Errorf("Expected sequence length to be 0, got %d", len(seq.sequence))
	}

	// Test case: preallocate is greater than 0, sequence should be preallocated
	seq = NewEmptyBioSequence(100)
	if seq == nil {
		t.Errorf("NewEmptyBioSequence(100) returned nil")
	} else if cap(seq.sequence) < 100 {
		t.Errorf("Expected sequence capacity to be at least 100, got %d", cap(seq.sequence))
	}

	// Test case: preallocate is negative, sequence should not be preallocated
	seq = NewEmptyBioSequence(-100)
	if seq == nil {
		t.Errorf("NewEmptyBioSequence(-100) returned nil")
	} else if len(seq.sequence) != 0 {
		t.Errorf("Expected sequence length to be 0, got %d", len(seq.sequence))
	}
}

// TestNewBioSequence tests the NewBioSequence function.
//
// It checks the correctness of the NewBioSequence function by validating that the BioSequence object
// created has the correct ID, sequence, and definition.
// The function performs two test cases:
//  1. Test case 1 checks if the BioSequence object created using the NewBioSequence function has
//     the expected ID, sequence, and definition when provided with valid inputs.
//  2. Test case 2 checks if the BioSequence object created using the NewBioSequence function has
//     the expected ID, sequence, and definition when provided with different valid inputs.
func TestNewBioSequence(t *testing.T) {
	// Test case 1:
	id := "seq1"
	sequence := []byte("ACGT")
	definition := "DNA sequence"
	expectedID := "seq1"
	expectedSequence := []byte("acgt")
	expectedDefinition := "DNA sequence"

	bs := NewBioSequence(id, sequence, definition)

	if bs.Id() != expectedID {
		t.Errorf("Expected ID to be %s, but got %s", expectedID, bs.Id())
	}

	if !bytes.Equal(bs.Sequence(), expectedSequence) {
		t.Errorf("Expected sequence to be %v, but got %v", expectedSequence, bs.Sequence())
	}

	if bs.Definition() != expectedDefinition {
		t.Errorf("Expected definition to be %s, but got %s", expectedDefinition, bs.Definition())
	}

	// Test case 2:
	id = "seq2"
	sequence = []byte("ATCG")
	definition = "RNA sequence"
	expectedID = "seq2"
	expectedSequence = []byte("atcg")
	expectedDefinition = "RNA sequence"

	bs = NewBioSequence(id, sequence, definition)

	if bs.Id() != expectedID {
		t.Errorf("Expected ID to be %s, but got %s", expectedID, bs.Id())
	}

	if !bytes.Equal(bs.Sequence(), expectedSequence) {
		t.Errorf("Expected sequence to be %v, but got %v", expectedSequence, bs.Sequence())
	}

	if bs.Definition() != expectedDefinition {
		t.Errorf("Expected definition to be %s, but got %s", expectedDefinition, bs.Definition())
	}
}

// TestNewBioSequenceWithQualities tests the NewBioSequenceWithQualities function.
//
// It tests that the BioSequence object is created with the correct id, sequence,
// definition, and qualities.
// Parameters:
// - t: A pointer to a testing.T object.
// Return type: None.
func TestNewBioSequenceWithQualities(t *testing.T) {
	id := "123"
	sequence := []byte("ATGC")
	definition := "DNA sequence"
	qualities := []byte("1234")

	bs := NewBioSequenceWithQualities(id, sequence, definition, qualities)

	// Test that the BioSequence object is created with the correct id
	if bs.Id() != id {
		t.Errorf("Expected id to be %s, but got %s", id, bs.Id())
	}

	// Test that the BioSequence object is created with the correct sequence
	if string(bs.Sequence()) != string(sequence) {
		t.Errorf("Expected sequence to be %s, but got %s", string(sequence), string(bs.Sequence()))
	}

	// Test that the BioSequence object is created with the correct definition
	if bs.Definition() != definition {
		t.Errorf("Expected definition to be %s, but got %s", definition, bs.Definition())
	}

	// Test that the BioSequence object is created with the correct qualities
	if string(bs.Qualities()) != string(qualities) {
		t.Errorf("Expected qualities to be %s, but got %s", string(qualities), string(bs.Qualities()))
	}
}

// TestBioSequence_Recycle tests the Recycle method of the BioSequence struct.
//
// Test case 1: Recycle a BioSequence object with non-nil slices and annotations.
// Test case 2: Recycle a nil BioSequence object.
// Test case 3: Recycle a BioSequence object with nil slices and annotations.
func TestBioSequence_Recycle(t *testing.T) {
	// Test case 1: Recycle a BioSequence object with non-nil slices and annotations
	sequence := &BioSequence{
		sequence:    []byte{'A', 'C', 'G', 'T'},
		feature:     []byte("..."),
		qualities:   []byte{30, 30, 30, 30},
		annotations: Annotation{"description": "Test"},
	}
	sequence.Recycle()

	if len(sequence.sequence) != 0 {
		t.Errorf("Expected sequence to be empty, got %v", sequence.sequence)
	}
	if len(sequence.feature) != 0 {
		t.Errorf("Expected feature to be empty, got %v", sequence.feature)
	}
	if len(sequence.qualities) != 0 {
		t.Errorf("Expected qualities to be empty, got %v", sequence.qualities)
	}
	if sequence.annotations != nil {
		t.Errorf("Expected annotations to be nil, got %v", sequence.annotations)
	}

	// Test case 2: Recycle a nil BioSequence object
	var nilSequence *BioSequence
	nilSequence.Recycle() // No panic expected

	// Test case 3: Recycle a BioSequence object with nil slices and annotations
	emptySequence := &BioSequence{}
	emptySequence.Recycle() // No panic expected
}

// TestCopy tests the Copy function of the BioSequence struct.
//
// It creates a new BioSequence and copies the fields from the original sequence
// to the new one. It then performs various tests to check if the fields were
// copied correctly.
//
// Parameters:
// - t: The testing.T object used for reporting test failures.
//
// Returns: None.
func TestCopy(t *testing.T) {
	seq := &BioSequence{
		id:        "test",
		sequence:  []byte("ATCG"),
		qualities: []byte("1234"),
		feature:   []byte("feature1...feature2"),
		annotations: Annotation{
			"annotation1": "value1",
			"annotation2": "value2",
		},
		annot_lock: &sync.Mutex{},
	}

	newSeq := seq.Copy()

	// Test if the id and definition fields are copied correctly
	if newSeq.id != seq.id {
		t.Errorf("Expected id to be %v, got %v", seq.id, newSeq.id)
	}
	// Test if the sequence, qualities, and feature fields are copied correctly
	if !reflect.DeepEqual(newSeq.sequence, seq.sequence) {
		t.Errorf("Expected sequence to be %v, got %v", seq.sequence, newSeq.sequence)
	}
	if !reflect.DeepEqual(newSeq.qualities, seq.qualities) {
		t.Errorf("Expected qualities to be %v, got %v", seq.qualities, newSeq.qualities)
	}
	if !reflect.DeepEqual(newSeq.feature, seq.feature) {
		t.Errorf("Expected feature to be %v, got %v", seq.feature, newSeq.feature)
	}

	// Test if the annotations are copied correctly
	if !reflect.DeepEqual(newSeq.annotations, seq.annotations) {
		t.Errorf("Expected annotations to be %v, got %v", seq.annotations, newSeq.annotations)
	}
}

// TestBioSequence_Id tests the Id method of the BioSequence struct.
//
// It initializes a BioSequence with an ID using the constructor and then
// verifies that the Id method returns the expected ID.
// The expected ID is "ABC123".
// The method takes no parameters and returns a string.
func TestBioSequence_Id(t *testing.T) {
	// Initialize BioSequence with an ID using the constructor
	bioSeq := NewBioSequence("ABC123", []byte(""), "")

	// Test case: ID is returned correctly
	expected := "ABC123"
	result := bioSeq.Id()
	if result != expected {
		t.Errorf("Expected ID to be %s, but got %s", expected, result)
	}
}

// TestBioSequenceDefinition tests the Definition() method of the BioSequence struct.
//
// This function verifies the behavior of the Definition() method in two test cases:
// 1. Empty definition: It creates a BioSequence object with an empty definition and verifies that the Definition() method returns an empty string.
// 2. Non-empty definition: It creates a BioSequence object with a non-empty definition and verifies that the Definition() method returns the expected definition.
func TestBioSequenceDefinition(t *testing.T) {
	// Test case 1: Empty definition
	seq1 := NewBioSequence("", []byte{}, "")
	expected1 := ""
	if got1 := seq1.Definition(); got1 != expected1 {
		t.Errorf("Expected %q, but got %q", expected1, got1)
	}

	// Test case 2: Non-empty definition
	seq2 := NewBioSequence("", []byte{}, "This is a definition")
	expected2 := "This is a definition"
	if got2 := seq2.Definition(); got2 != expected2 {
		t.Errorf("Expected %q, but got %q", expected2, got2)
	}
}

// TestBioSequenceSequence tests the Sequence() method of the BioSequence struct.
//
// It verifies the behavior of the Sequence() method under two scenarios:
// - Test case 1: Empty sequence
// - Test case 2: Non-empty sequence
//
// Parameter(s):
//   - t: The testing object provided by the testing framework.
//     It is used to report errors if the test fails.
//
// Return type(s):
// None.
func TestBioSequenceSequence(t *testing.T) {
	// Test case 1: Empty sequence
	seq := NewBioSequence("", []byte{}, "")
	expected := []byte{}
	actual := seq.Sequence()
	if !bytes.EqualFold(actual, expected) {
		t.Errorf("Expected %v, but got %v", expected, actual)
	}

	// Test case 2: Non-empty sequence
	seq = NewBioSequence("", []byte("atcg"), "")
	expected = []byte("atcg")
	actual = seq.Sequence()
	if !bytes.EqualFold(actual, expected) {
		t.Errorf("Expected %v, but got %v", expected, actual)
	}
}

// TestBioSequence_String tests the String method of the BioSequence struct.
//
// It includes two test cases:
//
//  1. Test case 1: Empty sequence
//     - Creates an empty BioSequence instance.
//     - Expects an empty string as the result of calling the String method on the BioSequence instance.
//
//  2. Test case 2: Non-empty sequence
//     - Creates a BioSequence instance with the sequence "acgt".
//     - Expects the sequence "acgt" as the result of calling the String method on the BioSequence instance.
//
// No parameters are required.
// No return types are specified.
func TestBioSequence_String(t *testing.T) {
	// Test case 1: Empty sequence
	seq1 := &BioSequence{}
	expected1 := ""
	if got1 := seq1.String(); got1 != expected1 {
		t.Errorf("Test case 1 failed: expected %s, got %s", expected1, got1)
	}

	// Test case 2: Non-empty sequence
	seq2 := &BioSequence{sequence: []byte("acgt")}
	expected2 := "acgt"
	if got2 := seq2.String(); got2 != expected2 {
		t.Errorf("Test case 2 failed: expected %s, got %s", expected2, got2)
	}
}

// TestBioSequence_Len tests the Len method of the BioSequence struct.
//
// It verifies the behavior of the method by performing multiple test cases.
// Each test case creates a BioSequence instance with a specific sequence and
// compares the actual length returned by the Len method with the expected
// length.
//
// Test 1: Empty sequence
//   - Create a BioSequence instance with an empty sequence.
//   - The expected length is 0.
//   - Check if the actual length returned by the Len method matches the expected
//     length. If not, report an error.
//
// Test 2: Sequence with 5 characters
//   - Create a BioSequence instance with a sequence of 5 characters.
//   - The expected length is 5.
//   - Check if the actual length returned by the Len method matches the expected
//     length. If not, report an error.
//
// Test 3: Sequence with 10 characters
//   - Create a BioSequence instance with a sequence of 10 characters.
//   - The expected length is 10.
//   - Check if the actual length returned by the Len method matches the expected
//     length. If not, report an error.
func TestBioSequence_Len(t *testing.T) {
	// Test 1: Empty sequence
	s1 := NewBioSequence("", nil, "")
	expected1 := 0
	if len := s1.Len(); len != expected1 {
		t.Errorf("Expected length: %d, but got: %d", expected1, len)
	}

	// Test 2: Sequence with 5 characters
	s2 := NewBioSequence("", []byte("ATCGT"), "")
	expected2 := 5
	if len := s2.Len(); len != expected2 {
		t.Errorf("Expected length: %d, but got: %d", expected2, len)
	}

	// Test 3: Sequence with 10 characters
	s3 := NewBioSequence("", []byte("AGCTAGCTAG"), "")
	expected3 := 10
	if len := s3.Len(); len != expected3 {
		t.Errorf("Expected length: %d, but got: %d", expected3, len)
	}
}

// TestHasQualities tests the HasQualities method of the BioSequence struct.
//
// It includes two test cases:
//
//  1. Test case 1: BioSequence with empty qualities slice
//     - Creates a BioSequence instance with an empty qualities slice.
//     - Expects false as the result of calling the HasQualities method on the BioSequence instance.
//
//  2. Test case 2: BioSequence with non-empty qualities slice
//     - Creates a BioSequence instance with a non-empty qualities slice.
//     - Expects true as the result of calling the HasQualities method on the BioSequence instance.
//
// No parameters are required.
// No return types are specified.
func TestHasQualities(t *testing.T) {
	// Test case 1: BioSequence with empty qualities slice
	seq1 := NewBioSequence("", []byte(""), "")
	seq1.qualities = []byte{}
	if seq1.HasQualities() != false {
		t.Errorf("Test case 1 failed: expected false, got true")
	}

	// Test case 2: BioSequence with non-empty qualities slice
	seq2 := NewBioSequence("", []byte(""), "")
	seq2.qualities = []byte{20, 30, 40}
	if seq2.HasQualities() != true {
		t.Errorf("Test case 2 failed: expected true, got false")
	}
}

// TestQualities tests the Qualities method of the BioSequence struct.
//
// It creates a BioSequence with a given sequence and qualities and sets them.
// Then it compares the returned qualities with the expected ones.
// If the qualities are not equal, it fails the test case.
//
// Test case 1: BioSequence has qualities
// - sequence: []byte("ATCG")
// - qualities: Quality{10, 20, 30, 40}
// - expected: Quality{10, 20, 30, 40}
//
// Test case 2: BioSequence does not have qualities
// - sequence: []byte("ATCG")
// - qualities: nil
// - expected: defaultQualities
//
// Parameters:
// - t: *testing.T - the testing struct for running test cases and reporting failures.
//
// Return type:
// None
func TestQualities(t *testing.T) {
	// Test case: BioSequence has qualities
	sequence := []byte("ATCG")
	qualities := Quality{10, 20, 30, 40}
	bioSeq := NewBioSequence("ABC123", sequence, "Test Sequence")
	bioSeq.SetQualities(qualities)

	result := bioSeq.Qualities()
	expected := qualities

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Test case failed: BioSequence has qualities")
	}

	// Test case: BioSequence does not have qualities
	defaultQualities := __make_default_qualities__(len(sequence))
	bioSeq = NewBioSequence("ABC123", sequence, "Test Sequence")
	bioSeq.SetQualities(nil)

	result = bioSeq.Qualities()
	expected = defaultQualities

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Test case failed: BioSequence does not have qualities")
	}
}

// TestBioSequence_Features tests the Features function of the BioSequence struct.
//
// It first tests the case when the feature string is empty. It creates a new BioSequence
// with an empty feature string and an empty byte slice. It expects an empty string as
// the result of calling the Features function on this BioSequence. If the result does
// not match the expected value, it prints an error message.
//
// It then tests the case when the feature string is non-empty. It creates a new BioSequence
// with an empty feature string and an empty byte slice. It sets the feature string to
// "test sequence" and expects "test sequence" as the result of calling the Features function
// on this BioSequence. If the result does not match the expected value, it prints an error message.
func TestBioSequence_Features(t *testing.T) {
	// Testing empty feature string
	seq := NewBioSequence("", []byte(""), "")
	expected := ""
	if got := seq.Features(); got != expected {
		t.Errorf("Expected %q, but got %q", expected, got)
	}

	// Testing non-empty feature string
	seq = NewBioSequence("", []byte(""), "")
	seq.feature = []byte("test sequence")
	expected = "test sequence"
	if got := seq.Features(); got != expected {
		t.Errorf("Expected %q, but got %q", expected, got)
	}
}

// TestHasAnnotation is a unit test function that tests the HasAnnotation method of the BioSequence struct.
//
// This function tests the behavior of the HasAnnotation method in different scenarios:
// - Test case: BioSequence with no annotations.
// - Test case: BioSequence with one annotation.
// - Test case: BioSequence with multiple annotations.
//
// The function verifies that the HasAnnotation method returns the expected boolean value for each test case.
// It uses the *testing.T parameter to report any test failures.
//
// No parameters.
// No return values.
func TestHasAnnotation(t *testing.T) {
	// Test case: BioSequence with no annotations
	seq := BioSequence{}
	expected := false
	if got := seq.HasAnnotation(); got != expected {
		t.Errorf("Expected %v, but got %v", expected, got)
	}

	// Test case: BioSequence with one annotation
	seq = BioSequence{annotations: map[string]interface{}{"annotation1": "value1"}}
	expected = true
	if got := seq.HasAnnotation(); got != expected {
		t.Errorf("Expected %v, but got %v", expected, got)
	}

	// Test case: BioSequence with multiple annotations
	seq = BioSequence{
		annotations: map[string]interface{}{
			"annotation1": "value1",
			"annotation2": "value2",
		},
	}
	expected = true
	if got := seq.HasAnnotation(); got != expected {
		t.Errorf("Expected %v, but got %v", expected, got)
	}
}

// TestBioSequenceAnnotations tests the Annotations method of the BioSequence struct.
//
// It verifies the behavior of the method when the `annotations` field of the BioSequence struct is nil and when it is not nil.
// The method should return the expected annotation values and fail the test if the returned annotations do not match the expected ones.
// The test cases cover both scenarios to ensure the correctness of the method.
func TestBioSequenceAnnotations(t *testing.T) {
	s := &BioSequence{}

	// Test case 1: Annotations is nil
	s.annotations = nil
	expected := GetAnnotation()
	actual := s.Annotations()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Test case 1 failed: Expected %v, but got %v", expected, actual)
	}

	// Test case 2: Annotations is not nil
	s.annotations = Annotation{}
	expected = s.annotations
	actual = s.Annotations()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Test case 2 failed: Expected %v, but got %v", expected, actual)
	}
}

func TestAnnotationsLock(t *testing.T) {
	// Test case 1: Lock the annotation of an empty BioSequence
	seq := NewEmptyBioSequence(0)
	seq.AnnotationsLock()

	// Test case 2: Lock the annotation of a BioSequence with existing annotations
	seq2 := NewEmptyBioSequence(0)
	seq2.annotations = map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}
	seq2.AnnotationsLock()
}

// TestBioSequence_MD5 tests the MD5 function of the BioSequence struct.
//
// It includes two test cases: one for an empty sequence and one for a non-empty sequence.
// Each test case creates a BioSequence instance with a specific sequence and compares the MD5 result with the expected value.
// If the result does not match the expected value, an error is reported using the t.Errorf function.
// The expected MD5 values are hardcoded in the test cases.
func TestBioSequence_MD5(t *testing.T) {
	// Test case 1: Empty sequence
	{
		s := &BioSequence{sequence: []byte("")}
		expected := [16]byte{
			0xd4, 0x1d, 0x8c, 0xd9, 0x8f, 0x00, 0xb2, 0x04,
			0xe9, 0x80, 0x09, 0x98, 0xec, 0xf8, 0x42, 0x7e,
		}
		result := s.MD5()
		if result != expected {
			t.Errorf("Test case 1 failed. Expected: %v, got: %v", expected, result)
		}
	}

	// Test case 2: Non-empty sequence
	{
		s := &BioSequence{sequence: []byte("ACGT")}
		expected := [16]byte{
			0xf1, 0xf8, 0xf4, 0xbf, 0x41, 0x3b, 0x16, 0xad,
			0x13, 0x57, 0x22, 0xaa, 0x45, 0x91, 0x04, 0x3e,
		}
		result := s.MD5()
		if result != expected {
			t.Errorf("Test case 2 failed. Expected: %v, got: %v", expected, result)
		}
	}
}

// TestBioSequence_Composition tests the Composition method of the BioSequence struct.
//
// It tests the method with three different test cases:
// 1. Empty sequence: It checks if the Composition method returns the expected composition when the sequence is empty.
// 2. Sequence with valid nucleotides: It checks if the Composition method returns the expected composition when the sequence contains valid nucleotides.
// 3. Sequence with invalid nucleotides: It checks if the Composition method returns the expected composition when the sequence contains invalid nucleotides.
//
// The expected composition for each test case is defined in a map where the keys are the nucleotides and the values are the expected counts.
// The Composition method is expected to return a map with the actual nucleotide counts.
//
// Parameters:
// - t: The testing.T object used for reporting test failures and logging.
//
// Return type: void.
func TestBioSequence_Composition(t *testing.T) {
	// Test case: Empty sequence
	seq1 := NewBioSequence("", []byte(""), "")
	expected1 := map[byte]int{'a': 0, 'c': 0, 'g': 0, 't': 0, 'o': 0}
	if result1 := seq1.Composition(); !reflect.DeepEqual(result1, expected1) {
		t.Errorf("Composition() returned incorrect result for empty sequence. Got %v, expected %v", result1, expected1)
	}

	// Test case: Sequence with valid nucleotides
	seq2 := NewBioSequence("", []byte("acgtACGT"), "")
	expected2 := map[byte]int{'a': 2, 'c': 2, 'g': 2, 't': 2, 'o': 0}
	if result2 := seq2.Composition(); !reflect.DeepEqual(result2, expected2) {
		t.Errorf("Composition() returned incorrect result for sequence with valid nucleotides. Got %v, expected %v", result2, expected2)
	}

	// Test case: Sequence with invalid nucleotides
	seq3 := NewBioSequence("", []byte("acgtACGT1234"), "")
	expected3 := map[byte]int{'a': 2, 'c': 2, 'g': 2, 't': 2, 'o': 4}
	if result3 := seq3.Composition(); !reflect.DeepEqual(result3, expected3) {
		t.Errorf("Composition() returned incorrect result for sequence with invalid nucleotides. Got %v, expected %v", result3, expected3)
	}
}
