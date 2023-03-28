package obiseq

import (
	log "github.com/sirupsen/logrus"
	"sync"
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

// > This function returns a pointer to a new `BioSequenceSlice` object
func NewBioSequenceSlice(size ...int) *BioSequenceSlice {
	slice := _BioSequenceSlicePool.Get().(*BioSequenceSlice)
	if len(size) > 0 {
		s := size[0]
		slice = slice.InsureCapacity(s)
		(*slice)=(*slice)[0:s]
	}
	return slice
}

// `MakeBioSequenceSlice()` returns a pointer to a new `BioSequenceSlice` struct
func MakeBioSequenceSlice(size ...int) BioSequenceSlice {
	return *NewBioSequenceSlice(size...)
}

func (s *BioSequenceSlice) Recycle(including_seq bool) {
	if s == nil {
		log.Panicln("Trying too recycle a nil pointer")
	}

	// Code added to potentially limit memory leaks
	if including_seq {
		for i := range *s {
			(*s)[i] .Recycle()
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

// Making sure that the slice has enough capacity to hold the number of elements that are being added
// to it.
func (s *BioSequenceSlice) InsureCapacity(capacity int) *BioSequenceSlice {
	var c int
	if s != nil {
		c = cap(*s)	
	} else {
		c = 0
	}

	if c < capacity {
		sl := make(BioSequenceSlice, 0,capacity)
		s = &sl
	}

	return s
}

// Appending the sequence to the slice.
func (s *BioSequenceSlice) Push(sequence *BioSequence) {
	*s = append(*s, sequence)
}

// Returning the last element of the slice and removing it from the slice.
func (s *BioSequenceSlice) Pop() *BioSequence {
	_s := (*s)[len(*s)-1]
	(*s)[len(*s)-1] = nil
	*s = (*s)[:len(*s)-1]
	return _s
}

// Returning the first element of the slice and removing it from the slice.
func (s *BioSequenceSlice) Pop0() *BioSequence {
	_s := (*s)[0]
	(*s)[0] = nil
	*s = (*s)[1:]
	return _s
}

// Test that a slice of sequences contains at least a sequence.
func (s BioSequenceSlice) NotEmpty() bool {
	return len(s) > 0
}
