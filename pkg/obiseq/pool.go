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

func RecycleSlice(s *[]byte) {
	if s != nil && *s != nil {
		*s = (*s)[:0]
		_BioSequenceByteSlicePool.Put(s)
	}
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

func GetAnnotation(values ...Annotation) Annotation {
	a := Annotation(nil)

	for a == nil {
		a = *(BioSequenceAnnotationPool.Get().(*Annotation))
	}

	if len(values) > 0 {
		goutils.CopyMap(a, values[0])
	}

	return a
}
