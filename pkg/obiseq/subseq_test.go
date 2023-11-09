package obiseq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSubsequence tests the Subsequence function.
//
// The function tests various cases of the Subsequence method of a BioSequence object.
// It checks different scenarios of subsequence slicing, including both valid and invalid parameters.
// The function is designed for unit testing purposes and uses the Go testing package.
// It asserts that the expected subsequence is returned for each test case and checks for any errors.
// The function also verifies the correctness of the subsequence qualities, if applicable.
// The test cases cover both non-circular and circular subsequence slicing.
// It ensures that the function handles different scenarios such as when `from` is greater than `to`,
// `from` or `to` is out of bounds, and normal subsequence slicing cases.
//
// TestSubsequence does not return any value.
func TestSubsequence(t *testing.T) {
	// Test case 1: Subsequence with valid parameters and non-circular
	seq := NewBioSequence("ID1", []byte("ATCG"), "")
	sub, err := seq.Subsequence(1, 3, false)
	assert.NoError(t, err)
	assert.Equal(t, []byte("tc"), sub.Sequence())

	seq = NewBioSequenceWithQualities("ID1", []byte("ATCG"), "", []byte{40, 30, 20, 10})
	sub, err = seq.Subsequence(1, 3, false)
	assert.NoError(t, err)
	assert.Equal(t, []byte("tc"), sub.Sequence())
	assert.Equal(t, Quality([]byte{30, 20}), sub.Qualities())

	// Test case 2: Subsequence with valid parameters and circular
	seq2 := NewBioSequence("ID1", []byte("ATCG"), "")
	sub2, err2 := seq2.Subsequence(3, 2, true)
	assert.NoError(t, err2)
	assert.Equal(t, []byte("gat"), sub2.Sequence())

	seq = NewBioSequenceWithQualities("ID1", []byte("ATCG"), "", []byte{40, 30, 20, 10})
	sub, err = seq.Subsequence(3, 2, true)
	assert.NoError(t, err)
	assert.Equal(t, []byte("gat"), sub.Sequence())
	assert.Equal(t, Quality([]byte{10, 40, 30}), sub.Qualities())

	// Test case 3: Subsequence with from greater than to and non-circular
	seq3 := NewBioSequence("ID1", []byte("ATCG"), "")
	_, err3 := seq3.Subsequence(3, 1, false)
	assert.EqualError(t, err3, "from greater than to")

	// Test case 4: Subsequence with from out of bounds
	seq4 := NewBioSequence("ID1", []byte("ATCG"), "")
	_, err4 := seq4.Subsequence(-1, 2, false)
	assert.EqualError(t, err4, "from out of bounds")

	// Test case 5: Subsequence with to out of bounds
	seq5 := NewBioSequence("ID1", []byte("ATCG"), "")
	_, err5 := seq5.Subsequence(2, 5, false)
	assert.EqualError(t, err5, "to out of bounds")
}
