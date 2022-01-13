package obiapat

/*
#cgo CFLAGS: -g -Wall
#include <stdlib.h>
#include "obiapat.h"
*/
import "C"
import (
	"errors"
	"unsafe"

	"git.metabarcoding.org/lecasofts/go/oa2/pkg/obiseq"
)

var MAX_PAT_LEN = int(C.MAX_PAT_LEN)

type ApatPattern struct {
	pointer *C.Pattern
}

type ApatSequence struct {
	pointer *C.Seq
}

var NilApatPattern = ApatPattern{nil}
var NilApatSequence = ApatSequence{nil}

func MakeApatPattern(pattern string, errormax int) (ApatPattern, error) {
	cpattern := C.CString(pattern)
	defer C.free(unsafe.Pointer(cpattern))
	cerrormax := C.int32_t(errormax)
	var errno C.int32_t
	var errmsg *C.char

	ap := C.buildPattern(cpattern, cerrormax, &errno, &errmsg)

	if ap == nil {
		message := C.GoString(errmsg)
		C.free(unsafe.Pointer(errmsg))
		return NilApatPattern, errors.New(message)
	}

	return ApatPattern{pointer: ap}, nil
}

func (pattern ApatPattern) ReverseComplement() (ApatPattern, error) {
	var errno C.int32_t
	var errmsg *C.char
	ap := C.complementPattern((*C.Pattern)(pattern.pointer), &errno, &errmsg)

	if ap == nil {
		message := C.GoString(errmsg)
		C.free(unsafe.Pointer(errmsg))
		return ApatPattern{nil}, errors.New(message)
	}

	return ApatPattern{pointer: ap}, nil
}

func (pattern ApatPattern) String() string {
	return C.GoString(pattern.pointer.cpat)
}

func (pattern ApatPattern) Length() int {
	return int(pattern.pointer.patlen)
}

func (pattern ApatPattern) Free() {
	C.free(unsafe.Pointer(pattern.pointer))
}

func (pattern ApatPattern) Print() {
	C.PrintDebugPattern(C.PatternPtr(pattern.pointer))
}

func MakeApatSequence(sequence obiseq.BioSequence, circular bool, recycle ...ApatSequence) (ApatSequence, error) {
	var errno C.int32_t
	var errmsg *C.char
	seqlen := sequence.Length()
	p := C.malloc(C.size_t(seqlen) + 1)
	ic := 0
	if circular {
		ic = 1
	}

	// copy the data into the buffer, by converting it to a Go array
	cBuf := (*[1 << 30]byte)(p)
	copy(cBuf[:], sequence.Sequence())
	cBuf[sequence.Length()] = 0

	var out *C.Seq

	if len(recycle) > 0 {
		out = recycle[0].pointer
	} else {
		out = nil
	}

	pseq := C.new_apatseq((*C.char)(p), C.int32_t(ic), C.int32_t(seqlen),
		(*C.Seq)(out),
		&errno, &errmsg)

	if pseq == nil {
		message := C.GoString(errmsg)
		C.free(unsafe.Pointer(errmsg))
		return NilApatSequence, errors.New(message)
	}

	seq := ApatSequence{pointer: pseq}

	//log.Println(C.GoString(pseq.cseq))
	// runtime.SetFinalizer(&seq, __free_apat_sequence__)

	return seq, nil
}

func (sequence ApatSequence) Length() int {
	return int(sequence.pointer.seqlen)
}

func (sequence ApatSequence) Free() {
	var errno C.int32_t
	var errmsg *C.char

	C.delete_apatseq(sequence.pointer,
		&errno, &errmsg)

	sequence.pointer = nil
}

func (pattern ApatPattern) FindAllIndex(sequence ApatSequence, limits ...int) (loc [][3]int) {
	begin := 0
	length := sequence.Length()

	if len(limits) > 0 {
		begin = limits[0]
	}

	if len(limits) > 1 {
		length = limits[1]
	}

	nhits := int(C.ManberAll(sequence.pointer,
		pattern.pointer,
		0,
		C.int32_t(begin),
		C.int32_t(length+C.MAX_PAT_LEN)))

	//log.Printf("match count : %d\n", nhits)

	if nhits == 0 {
		return nil
	}

	stktmp := (*[1 << 30]int32)(unsafe.Pointer(sequence.pointer.hitpos[0].val))
	errtmp := (*[1 << 30]int32)(unsafe.Pointer(sequence.pointer.hiterr[0].val))
	patlen := int(pattern.pointer.patlen)

	for i := 0; i < nhits; i++ {
		start := int(stktmp[i])
		err := int(errtmp[i])

		loc = append(loc, [3]int{start, start + patlen, err})
	}

	return loc
}
