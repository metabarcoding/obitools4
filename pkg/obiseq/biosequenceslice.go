package obiseq

import (
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

// BioSequenceSlice represents a collection or a set of BioSequence.
//
// BioSequenceSlice is used to define BioSequenceBatch
// a memory pool of BioSequenceSlice is managed to limit allocations.
type BioSequenceSlice []*BioSequence

var _BioSequenceSlicePool = sync.Pool{
	New: func() interface{} {
		bs := make(BioSequenceSlice, 0, 10)
		return &bs
	},
}

// NewBioSequenceSlice returns a new BioSequenceSlice with the specified size.
//
// The size parameter is optional. If provided, the returned slice will be
// resized accordingly.
//
// Returns a pointer to the newly created BioSequenceSlice.
func NewBioSequenceSlice(size ...int) *BioSequenceSlice {
	slice := _BioSequenceSlicePool.Get().(*BioSequenceSlice)
	if len(size) > 0 {
		s := size[0]
		slice = slice.EnsureCapacity(s)
		(*slice) = (*slice)[0:s]
	}
	return slice
}

// MakeBioSequenceSlice creates a new BioSequenceSlice with the specified size(s).
//
// Parameters:
// - size: The size(s) of the BioSequenceSlice to create (optional).
//
// Return:
// A new BioSequenceSlice with the specified size(s).
func MakeBioSequenceSlice(size ...int) BioSequenceSlice {
	return *NewBioSequenceSlice(size...)
}

// Recycle cleans up the BioSequenceSlice by recycling its elements and resetting its length.
//
// If including_seq is true, each element of the BioSequenceSlice is recycled using the Recycle method,
// and then set to nil. If including_seq is false, each element is simply set to nil.
//
// The function does not return anything.
func (s *BioSequenceSlice) Recycle(including_seq bool) {
	if s == nil {
		log.Panicln("Trying too recycle a nil pointer")
	}

	// Code added to potentially limit memory leaks
	if including_seq {
		for i := range *s {
			(*s)[i].Recycle()
			(*s)[i] = nil
		}

	} else {
		for i := range *s {
			(*s)[i] = nil
		}
	}

	*s = (*s)[:0]
	_BioSequenceSlicePool.Put(s)
}

// EnsureCapacity ensures that the BioSequenceSlice has a minimum capacity
//
// It takes an integer `capacity` as a parameter, which represents the desired minimum capacity of the BioSequenceSlice.
// It returns a pointer to the BioSequenceSlice.
func (s *BioSequenceSlice) EnsureCapacity(capacity int) *BioSequenceSlice {
	var c int
	if s != nil {
		c = cap(*s)
	} else {
		c = 0
	}

	n := 0
	for capacity > c {
		old_c := c
		*s = slices.Grow(*s, capacity)
		c = cap(*s)
		if c < capacity {
			n++
			if n < 4 {
				log.Warnf("cannot allocate a Biosequence Slice of size %d (only %d from %d)", capacity, c, old_c)
			} else {
				log.Panicf("cannot allocate a Biosequence Slice of size %d (only %d from %d)", capacity, c, old_c)
			}
		}
	}

	return s
}

// Push appends a BioSequence to the BioSequenceSlice.
//
// It takes a pointer to a BioSequenceSlice and a BioSequence as parameters.
// It does not return anything.
func (s *BioSequenceSlice) Push(sequence *BioSequence) {
	*s = append(*s, sequence)
}

// Pop returns and removes the last element from the BioSequenceSlice.
//
// It does not take any parameters.
// It returns *BioSequence, the last element of the slice.
func (s *BioSequenceSlice) Pop() *BioSequence {
	// Get the length of the slice
	length := len(*s)

	// If the slice is empty, return nil
	if length == 0 {
		return nil
	}

	// Get the last element of the slice
	lastElement := (*s)[length-1]

	// Set the last element to nil
	(*s)[length-1] = nil

	// Remove the last element from the slice
	*s = (*s)[:length-1]

	// Return the last element
	return lastElement
}

// Pop0 returns and removes the first element of the BioSequenceSlice.
//
// It does not take any parameters.
// It returns a pointer to a BioSequence object.
func (s *BioSequenceSlice) Pop0() *BioSequence {
	if len(*s) == 0 {
		return nil
	}
	firstElement := (*s)[0]
	(*s)[0] = nil
	*s = (*s)[1:]
	return firstElement
}

// NotEmpty checks if the BioSequenceSlice is not empty.
//
// No parameters.
// Returns a boolean value indicating if the BioSequenceSlice is not empty.
func (s BioSequenceSlice) NotEmpty() bool {
	return len(s) > 0
}

// Len returns the length of the BioSequenceSlice.
//
// It has no parameters.
// It returns an integer.
func (s BioSequenceSlice) Len() int {
	return len(s)
}

// Size returns the total size of the BioSequenceSlice.
//
// It calculates the size by iterating over each BioSequence in the slice
// and summing up their lengths.
//
// Returns an integer representing the total size of the BioSequenceSlice.
func (s BioSequenceSlice) Size() int {
	size := 0

	for _, s := range s {
		size += s.Len()
	}

	return size
}

func (s BioSequenceSlice) AttributeKeys(skip_map bool) obiutils.Set[string] {
	keys := obiutils.MakeSet[string]()

	for _, k := range s {
		keys = keys.Union(k.AttributeKeys(skip_map))
	}

	return keys
}

func (s *BioSequenceSlice) SortOnCount(reverse bool) {
	slices.SortFunc(*s, func(a, b *BioSequence) int {
		if reverse {
			return b.Count() - a.Count()
		} else {
			return a.Count() - b.Count()
		}
	})
}

func (s *BioSequenceSlice) SortOnLength(reverse bool) {
	slices.SortFunc(*s, func(a, b *BioSequence) int {
		if reverse {
			return b.Len() - a.Len()
		} else {
			return a.Len() - b.Len()
		}
	})
}
