package obiseq

import (
	"log"
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
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
		if cap(*s) <= 1024 {
			_BioSequenceByteSlicePool.Put(s)
		}
	}
}

// It returns a slice of bytes from a pool of slices.
//
// the slice can be prefilled with the provided values
func GetSlice(capacity int) []byte {
	p := (*[]byte)(nil)
	if capacity <= 1024 {
		p = _BioSequenceByteSlicePool.Get().(*[]byte)
	}

	if p == nil || *p == nil || cap(*p) < capacity {
		return make([]byte, 0, capacity)
	}

	s := *p

	if cap(s) < capacity {
		log.Panicln("Bizarre... j'aurai pourtant cru")
	}

	return s
}

func CopySlice(src []byte) []byte {
	sl := GetSlice(len(src))
	sl = sl[0:len(src)]

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

// GetAnnotation returns an Annotation from the BioSequenceAnnotationPool.
//
// It takes as argument O or 1 Annotation annotation object.
// If an annotation object is passed, it is copied into the new Annotation.
//
// It returns an Annotation.
func GetAnnotation(values ...Annotation) Annotation {
	a := Annotation(nil)

	for a == nil {
		a = *(BioSequenceAnnotationPool.Get().(*Annotation))
	}

	if len(values) > 0 {
		obiutils.MustFillMap(a, values[0])
	}

	return a
}
