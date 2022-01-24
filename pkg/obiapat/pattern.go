package obiapat

/*
#cgo CFLAGS: -g -Wall
#include <stdlib.h>
#include "obiapat.h"
*/
import "C"
import (
	"errors"
	"runtime"
	"unsafe"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

var _MaxPatLen = int(C.MAX_PAT_LEN)

// ApatPattern stores a regular pattern usable by the
// Apat algorithm functions and methods
type _ApatPattern struct {
	pointer *C.Pattern
}

type ApatPattern struct {
	pointer *_ApatPattern
}

// ApatSequence stores sequence in structure usable by the
// Apat algorithm functions and methods
type ApatSequence struct {
	pointer *C.Seq
}

// NilApatPattern is the nil instance of the BuildAlignArena
// type.
var NilApatPattern = ApatPattern{nil}

// NilApatSequence is the nil instance of the ApatSequence
// type.
var NilApatSequence = ApatSequence{nil}

// MakeApatPattern builds a new ApatPattern.
// The created object wrap a C allocated structure.
// Do not forget to free it when it is no more needed
// to forbid memory leaks using the Free methode of the
// ApatPattern.
// The pattern is a short DNA sequence (up to 64 symboles).
// Ambiguities can be represented or using UIPAC symboles,
// or using the [...] classical in regular pattern grammar.
// For example, the ambiguity A/T can be indicated using W
// or [AT]. A nucleotide can be negated by preceding it with
// a '!'. The APAT algorithm allows for error during the
// matching process. The maximum number of tolerated error
// is indicated at the construction of the pattern using
// the errormax parameter. Some positions can be marked as not
// allowed for mismatches. They have to be signaled using a '#'
// sign after the corresponding nucleotide.
func MakeApatPattern(pattern string, errormax int) (ApatPattern, error) {
	cpattern := C.CString(pattern)
	defer C.free(unsafe.Pointer(cpattern))
	cerrormax := C.int32_t(errormax)
	var errno C.int32_t
	var errmsg *C.char

	apc := C.buildPattern(cpattern, cerrormax, &errno, &errmsg)

	if apc == nil {
		message := C.GoString(errmsg)
		C.free(unsafe.Pointer(errmsg))
		return NilApatPattern, errors.New(message)
	}

	ap := _ApatPattern{apc}

	runtime.SetFinalizer(&ap, func(p *_ApatPattern) {
		// log.Printf("Finaliser called on %s\n", C.GoString(p.pointer.cpat))
		C.free(unsafe.Pointer(p.pointer))
	})

	return ApatPattern{pointer: &ap}, nil
}

// ReverseComplement method builds a new ApatPattern
// matching the reverse complemented sequence of the original
// pattern.
func (pattern ApatPattern) ReverseComplement() (ApatPattern, error) {
	var errno C.int32_t
	var errmsg *C.char
	apc := C.complementPattern((*C.Pattern)(pattern.pointer.pointer), &errno, &errmsg)

	if apc == nil {
		message := C.GoString(errmsg)
		C.free(unsafe.Pointer(errmsg))
		return ApatPattern{nil}, errors.New(message)
	}

	ap := _ApatPattern{apc}

	runtime.SetFinalizer(&ap, func(p *_ApatPattern) {
		// log.Printf("Finaliser called on %s\n", C.GoString(p.pointer.cpat))
		C.free(unsafe.Pointer(p.pointer))
	})

	return ApatPattern{pointer: &ap}, nil
}

// String method casts the ApatPattern to a Go String.
func (pattern ApatPattern) String() string {
	return C.GoString(pattern.pointer.pointer.cpat)
}

// Length method returns the length of the matched pattern.
func (pattern ApatPattern) Length() int {
	return int(pattern.pointer.pointer.patlen)
}

// Release the C allocated memory of an ApatPattern instance.
//
// Thee method ensurse that the C structure wrapped in
// an ApatPattern instance is released. Normally this
// action is taken in charge by a finalizer and the call
// to the Free meethod is not mandatory. Nevertheless,
// If you choose to call this method, it will disconnect
// the finalizer associated to the ApatPattern instance
// to avoid double freeing.
//
func (pattern ApatPattern) Free() {
	// log.Printf("Free called on %s\n", C.GoString(pattern.pointer.pointer.cpat))
	C.free(unsafe.Pointer(pattern.pointer.pointer))
	runtime.SetFinalizer(pattern.pointer, nil)

	pattern.pointer = nil
}

// Print method prints the ApatPattern to the standard output.
// This is mainly a debug method.
func (pattern ApatPattern) Print() {
	C.PrintDebugPattern(C.PatternPtr(pattern.pointer.pointer))
}

// MakeApatSequence casts an obiseq.BioSequence to an ApatSequence.
// The circular parameter indicates the topology of the sequence.
// if sequence is circular (ciruclar = true), the match can occurs
// at the junction. To limit memory allocation, it is possible to provide
// an already allocated ApatSequence to recycle its allocated memory.
// The provided sequence is no more usable after the call.
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

// Length method returns the length of the ApatSequence.
func (sequence ApatSequence) Length() int {
	return int(sequence.pointer.seqlen)
}

// Free method ensure that the C structure wrapped is
// desallocated
func (sequence ApatSequence) Free() {
	var errno C.int32_t
	var errmsg *C.char

	C.delete_apatseq(sequence.pointer,
		&errno, &errmsg)

	sequence.pointer = nil
}

// FindAllIndex methood returns the position of every occurrences of the
// pattern on the provided sequences. The search can be limited
// to a portion of the sequence by adding one or two integer parameters
// when calling the FindAllIndex method. The fisrt optional argument indicates
// the starting point of the search. The first nucleotide of the sequence is
// indexed as 0. The second optional argument indicates the length of the region
// where the pattern is looked for.
// The FindAllIndex methood returns return a slice of [3]int. The two firsts
// values of the [3]int indicate respectively the start and the end position of
// the match. Following the GO convention the end position is not included in the
// match. The third value indicates the number of error detected for this occurrence.
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
		pattern.pointer.pointer,
		0,
		C.int32_t(begin),
		C.int32_t(length+C.MAX_PAT_LEN)))

	if nhits == 0 {
		return nil
	}

	stktmp := (*[1 << 30]int32)(unsafe.Pointer(sequence.pointer.hitpos[0].val))
	errtmp := (*[1 << 30]int32)(unsafe.Pointer(sequence.pointer.hiterr[0].val))
	patlen := int(pattern.pointer.pointer.patlen)

	for i := 0; i < nhits; i++ {
		start := int(stktmp[i])
		err := int(errtmp[i])

		loc = append(loc, [3]int{start, start + patlen, err})
	}

	return loc
}
