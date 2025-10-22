package obiseq

import (
	"errors"
	"fmt"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obilog"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

// BioSequenceSlice represents a collection or a set of BioSequence.
//
// BioSequenceSlice is used to define BioSequenceBatch
// a memory pool of BioSequenceSlice is managed to limit allocations.
type BioSequenceSlice []*BioSequence

// NewBioSequenceSlice returns a new BioSequenceSlice with the specified size.
//
// The size parameter is optional. If provided, the returned slice will be
// resized accordingly.
//
// Returns a pointer to the newly created BioSequenceSlice.
func NewBioSequenceSlice(size ...int) *BioSequenceSlice {
	capacity := 0
	if len(size) > 0 {
		capacity = size[0]
	}

	slice := make(BioSequenceSlice, capacity)

	return &slice
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
				obilog.Warnf("cannot allocate a Biosequence Slice of size %d (only %d from %d)", capacity, c, old_c)
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

func (s BioSequenceSlice) AttributeKeys(skip_map, skip_definition bool) obiutils.Set[string] {
	keys := obiutils.MakeSet[string]()

	for _, k := range s {
		keys = keys.Union(k.AttributeKeys(skip_map, skip_definition))
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

func (s *BioSequenceSlice) ExtractTaxonomy(taxonomy *obitax.Taxonomy, seqAsTaxa bool) (*obitax.Taxonomy, error) {
	var err error

	for _, s := range *s {
		path := s.Path()
		if seqAsTaxa {
			if len(path) == 0 {
				return nil, fmt.Errorf("sequence %v has no path", s.Id())
			}
			last := path[len(path)-1]
			taxname, _ := obiutils.SplitInTwo(last, ':')
			if idx, ok := s.GetIntAttribute("seq_number"); !ok {
				return nil, errors.New("sequences are not numbered")
			} else {
				path = append(path, fmt.Sprintf("%s:SEQ%010d [%s]@sequence", taxname, idx, s.Id()))
			}

		}

		taxonomy, err = taxonomy.InsertPathString(path)

		if err != nil {
			return nil, err
		}

	}

	return taxonomy, nil
}
