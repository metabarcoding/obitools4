package obiseq

import (
	"sync"
)

type BioSequenceSlice []*BioSequence

var _BioSequenceSlicePool = sync.Pool{
	New: func() interface{} {
		bs := make(BioSequenceSlice, 0, 10)
		return &bs
	},
}

func NewBioSequenceSlice() *BioSequenceSlice {
	return _BioSequenceSlicePool.Get().(*BioSequenceSlice)
}

func MakeBioSequenceSlice() BioSequenceSlice {
	return *NewBioSequenceSlice()
}

func (s *BioSequenceSlice) Recycle() {
	// if s == nil {
	// 	log.Panicln("Trying too recycle a nil pointer")
	// }

	// // Code added to potentially limit memory leaks
	// for i := range *s {
	// 	(*s)[i] = nil
	// }

	// *s = (*s)[:0]
	// _BioSequenceSlicePool.Put(s)
}

func (s *BioSequenceSlice) Push(sequence *BioSequence) {
	*s = append(*s, sequence)
}

func (s *BioSequenceSlice) Pop() *BioSequence {
	_s := (*s)[len(*s)-1]
	(*s)[len(*s)-1] = nil
	*s = (*s)[:len(*s)-1]
	return _s
}

func (s *BioSequenceSlice) Pop0() *BioSequence {
	_s := (*s)[0]
	(*s)[0] = nil
	*s = (*s)[1:]
	return _s
}

func (s BioSequenceSlice) NotEmpty() bool {
	return len(s) > 0
}
