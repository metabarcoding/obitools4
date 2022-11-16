package obiseq

import (
	"log"
	"sync"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/goutils"
)

var _BioSequenceByteSlicePool = sync.Pool{
	New: func() interface{} {
		bs := make([]byte, 0, 300)
		return &bs
	},
}

func RecycleSlice(s *[]byte) {
	if s != nil && cap(*s) > 0 {
		*s = (*s)[:0]
		if cap(*s) == 0 {
			log.Panicln("trying to store a NIL slice in the pool", s == nil, *s == nil, cap(*s))
		}
		_BioSequenceByteSlicePool.Put(s)
	}
}

// It returns a slice of bytes from a pool of slices.
//
// the slice can be prefilled with the provided values
func GetSlice(capacity int) []byte {
	p := _BioSequenceByteSlicePool.Get().(*[]byte)

	if p == nil || *p == nil || cap(*p) < capacity {
		s := make([]byte, 0, capacity)
		p = &s
	}
	s := *p

	return s
}

func CopySlice(src []byte) []byte {
	sl := GetSlice(len(src))[0:len(src)]

	copy(sl, src)

	return sl
}

var BioSequenceAnnotationPool = sync.Pool{
	New: func() interface{} {
		bs := make(Annotation, 5)
		return &bs
	},
}

func RecycleAnnotation(a *Annotation) {
	if a != nil {
		for k := range *a {
			delete(*a, k)
		}
		BioSequenceAnnotationPool.Put(a)
	}
}

// It returns a new Annotation object, initialized with the values from the first argument
func GetAnnotation(values ...Annotation) Annotation {
	a := Annotation(nil)

	for a == nil {
		a = *(BioSequenceAnnotationPool.Get().(*Annotation))
	}

	if len(values) > 0 {
		goutils.MustFillMap(a, values[0])
	}

	return a
}
