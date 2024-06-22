package obiseq

import (
	"sync"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

var _BioSequenceByteSlicePool = sync.Pool{
	New: func() interface{} {
		bs := make([]byte, 0, 300)
		return &bs
	},
}

// RecycleSlice recycles a byte slice by clearing its contents and returning it
// to a pool if it is small enough.
//
// Parameters: - s: a pointer to a byte slice that will be recycled.
//
// This function first checks if the input slice is not nil and has a non-zero
// capacity. If so, it clears the contents of the slice by setting its length to
// 0. Then, it checks if the capacity of the slice is less than or equal to
// 1024. If it is, the function puts the slice into a pool for reuse. If the
// capacity is 0 or greater than 1024, the function does nothing. If the input
// slice is nil or has a zero capacity, the function logs a panic message.
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

// GetSlice returns a byte slice with the specified capacity.
//
// The function first checks if the capacity is less than or equal to 1024. If it is,
// it retrieves a byte slice from the _BioSequenceByteSlicePool. If the retrieved
// slice is nil, has a nil underlying array, or has a capacity less than the
// specified capacity, a new byte slice is created with the specified capacity.
// If the capacity is greater than 1024, a new byte slice is created with the
// specified capacity.
//
// The function returns the byte slice.
//
// Parameters:
// - capacity: the desired capacity of the byte slice.
//
// Return type:
// - []byte: the byte slice with the specified capacity.
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
		bs := make(Annotation, 1)
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
	a := (*Annotation)(nil)

	for a == nil || (*a == nil) {
		a = BioSequenceAnnotationPool.Get().(*Annotation)
	}

	annot := *a

	if len(values) > 0 {
		obiutils.MustFillMap(annot, values[0])
	}

	return annot
}
