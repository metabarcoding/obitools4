package obiiter

import "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"

type BioSequenceBatch struct {
	source string
	slice  obiseq.BioSequenceSlice
	order  int
}

var NilBioSequenceBatch = BioSequenceBatch{"", nil, -1}

// MakeBioSequenceBatch creates a new BioSequenceBatch with the given source, order, and sequences.
//
// Parameters:
// - source: The source of the BioSequenceBatch.
// - order: The order of the BioSequenceBatch.
// - sequences: The slice of BioSequence.
//
// Returns:
// - BioSequenceBatch: The newly created BioSequenceBatch.
func MakeBioSequenceBatch(
	source string,
	order int,
	sequences obiseq.BioSequenceSlice) BioSequenceBatch {

	return BioSequenceBatch{
		source: source,
		slice:  sequences,
		order:  order,
	}
}

// Order returns the order of the BioSequenceBatch.
//
// Returns:
// - int: The order of the BioSequenceBatch.
func (batch BioSequenceBatch) Order() int {
	return batch.order
}

// Source returns the source of the BioSequenceBatch.
//
// Returns:
// - string: The source of the BioSequenceBatch.
func (batch BioSequenceBatch) Source() string {
	return batch.source
}

// Reorder updates the order of the BioSequenceBatch and returns the updated batch.
//
// Parameters:
// - newOrder: The new order value to assign to the BioSequenceBatch.
//
// Returns:
// - BioSequenceBatch: The updated BioSequenceBatch with the new order value.
func (batch BioSequenceBatch) Reorder(newOrder int) BioSequenceBatch {
	batch.order = newOrder
	return batch
}

// Slice returns the BioSequenceSlice contained within the BioSequenceBatch.
//
// Returns:
// - obiseq.BioSequenceSlice: The BioSequenceSlice contained within the BioSequenceBatch.
func (batch BioSequenceBatch) Slice() obiseq.BioSequenceSlice {
	return batch.slice
}

// Len returns the number of BioSequence elements in the given BioSequenceBatch.
//
// Parameters:
// - batch: The BioSequenceBatch to get the length from.
//
// Return type:
// - int: The number of BioSequence elements in the BioSequenceBatch.
func (batch BioSequenceBatch) Len() int {
	return len(batch.slice)
}

// NotEmpty returns whether the BioSequenceBatch is empty or not.
//
// It checks if the BioSequenceSlice contained within the BioSequenceBatch is not empty.
//
// Returns:
// - bool: True if the BioSequenceBatch is not empty, false otherwise.
func (batch BioSequenceBatch) NotEmpty() bool {
	return batch.slice.NotEmpty()
}

// Pop0 returns and removes the first element of the BioSequenceBatch.
//
// It does not take any parameters.
// It returns a pointer to a BioSequence object.
func (batch BioSequenceBatch) Pop0() *obiseq.BioSequence {
	return batch.slice.Pop0()
}

// IsNil checks if the BioSequenceBatch's slice is nil.
//
// This function takes a BioSequenceBatch as a parameter and returns a boolean value indicating whether the slice of the BioSequenceBatch is nil or not.
//
// Parameters:
// - batch: The BioSequenceBatch to check for nil slice.
//
// Returns:
// - bool: True if the BioSequenceBatch's slice is nil, false otherwise.
func (batch BioSequenceBatch) IsNil() bool {
	return batch.slice == nil
}

// Recycle cleans up the BioSequenceBatch by recycling its elements and resetting its slice.
//
// If including_seq is true, each element of the BioSequenceBatch's slice is recycled using the Recycle method,
// and then set to nil. If including_seq is false, each element is simply set to nil.
//
// This function does not return anything.
func (batch BioSequenceBatch) Recycle(including_seq bool) {
	batch.slice.Recycle(including_seq)
	batch.slice = nil
}
