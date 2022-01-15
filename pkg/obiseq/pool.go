package obiseq

import (
	"sync"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/goutils"
)

var _BioSequenceByteSlicePool = sync.Pool{
	New: func() interface{} {
		bs := make([]byte, 0, 300)
		return &bs
	},
}

func RecycleSlice(s []byte) {
	s0 := s[:0]
	_BioSequenceByteSlicePool.Put(&s0)
}

func GetSlice(values ...byte) []byte {
	s := *(_BioSequenceByteSlicePool.Get().(*[]byte))

	if len(values) > 0 {
		s = append(s, values...)
	}

	return s
}

var BioSequenceAnnotationPool = sync.Pool{
	New: func() interface{} {
		bs := make(Annotation, 100)
		return &bs
	},
}

func RecycleAnnotation(a Annotation) {
	for k := range a {
		delete(a, k)
	}
	BioSequenceAnnotationPool.Put(&(a))
}

func GetAnnotation(values ...Annotation) Annotation {
	a := *(BioSequenceAnnotationPool.Get().(*Annotation))

	if len(values) > 0 {
		goutils.CopyMap(a, values[0])
	}

	return a
}

// var __bioseq__pool__ = sync.Pool{
// 	New: func() interface{} {
// 		var bs _BioSequence
// 		bs.annotations = make(Annotation, 50)
// 		return &bs
// 	},
// }

// func MakeEmptyBioSequence() BioSequence {
// 	bs := BioSequence{__bioseq__pool__.Get().(*_BioSequence)}
// 	return bs
// }

// func MakeBioSequence(id string,
// 	sequence []byte,
// 	definition string) BioSequence {
// 	bs := MakeEmptyBioSequence()
// 	bs.SetId(id)
// 	bs.Write(sequence)
// 	bs.SetDefinition(definition)
// 	return bs
// }

// func (sequence *BioSequence) Recycle() {
// 	sequence.Reset()
// 	__bioseq__pool__.Put(sequence.sequence)
// 	sequence.sequence = nil
// }
