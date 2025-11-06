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
	"strings"
	"unsafe"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obialign"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

var _MaxPatLen = int(C.MAX_PAT_LEN)

// var _AllocatedApaSequences = int64(0)
var _AllocatedApaPattern = 0

// ApatPattern stores a regular pattern usable by the
// Apat algorithm functions and methods
type _ApatPattern struct {
	pointer C.PatternPtr
	pattern string
}

type ApatPattern struct {
	pointer *_ApatPattern
}

// ApatSequence stores sequence in structure usable by the
// Apat algorithm functions and methods
type _ApatSequence struct {
	pointer   *C.Seq
	reference *obiseq.BioSequence
}

type ApatSequence struct {
	pointer *_ApatSequence
}

// NilApatPattern is the nil instance of the BuildAlignArena
// type.
var NilApatPattern = ApatPattern{nil}

// NilApatSequence is the nil instance of the ApatSequence
// type.
var NilApatSequence = ApatSequence{nil}

// MakeApatPattern creates an ApatPattern object based on the given pattern, error maximum, and allowsIndel flag.
//
// Parameters:
// The pattern is a short DNA sequence (up to 64 symbols).
// Ambiguities can be represented or using UIPAC symboles,
// or using the [...] classical in regular pattern grammar.
// For example, the ambiguity A/T can be indicated using W
// or [AT]. A nucleotide can be negated by preceding it with
// a '!'. The pattern is converted to uppercase.
// Some positions can be marked as not
// allowed for mismatches. They have to be signaled using a '#'
// sign after the corresponding nucleotide.
//
// errormax is the maximum number of errors allowed in the pattern.
//
// allowsIndel is a flag indicating whether indels are allowed in the pattern.
//
// Returns an ApatPattern object and an error if the pattern is invalid.
func MakeApatPattern(pattern string, errormax int, allowsIndel bool) (ApatPattern, error) {

	cpattern := C.CString(pattern)
	defer C.free(unsafe.Pointer(cpattern))
	cerrormax := C.int32_t(errormax)

	callosindel := C.uint8_t(0)
	if allowsIndel {
		callosindel = C.uint8_t(1)
	}

	var errno C.int32_t
	var errmsg *C.char

	apc := C.buildPattern(cpattern, cerrormax, callosindel, &errno, &errmsg)

	if apc == nil {
		message := C.GoString(errmsg)
		C.free(unsafe.Pointer(errmsg))
		return NilApatPattern, errors.New(message)
	}

	ap := _ApatPattern{apc, pattern}

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
	spat := strings.ToLower(C.GoString(apc.cpat))
	ap := _ApatPattern{apc, spat}

	runtime.SetFinalizer(&ap, func(p *_ApatPattern) {
		// log.Printf("Finaliser called on %s\n", C.GoString(p.pointer.cpat))
		C.free(unsafe.Pointer(p.pointer))
	})

	return ApatPattern{pointer: &ap}, nil
}

// String method casts the ApatPattern to a Go String.
func (pattern ApatPattern) String() string {
	return pattern.pointer.pattern
	//return C.GoString(pattern.pointer.pointer.cpat)
}

// Len method returns the length of the matched pattern.
func (pattern ApatPattern) Len() int {
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
func (pattern ApatPattern) Free() {
	// log.Printf("Free called on %s\n", C.GoString(pattern.pointer.pointer.cpat))
	if pattern.pointer != nil {
		C.free(unsafe.Pointer(pattern.pointer.pointer))
		runtime.SetFinalizer(pattern.pointer, nil)

		pattern.pointer = nil
	}
}

// Print method prints the ApatPattern to the standard output.
// This is mainly a debug method.
func (pattern ApatPattern) Print() {
	C.PrintDebugPattern((*C.Pattern)(pattern.pointer.pointer))
}

// MakeApatSequence casts an obiseq.BioSequence to an ApatSequence.
// The circular parameter indicates the topology of the sequence.
// if sequence is circular (ciruclar = true), the match can occurs
// at the junction. To limit memory allocation, it is possible to provide
// an already allocated ApatSequence to recycle its allocated memory.
// The provided sequence is no more usable after the call.
func MakeApatSequence(sequence *obiseq.BioSequence, circular bool, recycle ...ApatSequence) (ApatSequence, error) {
	var errno C.int32_t
	var errmsg *C.char
	seqlen := sequence.Len()

	ic := 0
	if circular {
		ic = 1
	}

	var out *C.Seq

	if len(recycle) > 0 {
		out = recycle[0].pointer.pointer
	} else {
		out = nil
	}

	// copy the data into the buffer, by converting it to a Go array
	p := unsafe.Pointer(unsafe.SliceData(sequence.Sequence()))
	pseqc := C.new_apatseq((*C.char)(p), C.int32_t(ic), C.int32_t(seqlen),
		(*C.Seq)(out),
		&errno, &errmsg)

	if pseqc == nil {
		message := C.GoString(errmsg)
		C.free(unsafe.Pointer(errmsg))
		if p != nil {
			C.free(p)
			// atomic.AddInt64(&_AllocatedApaSequences, -1)
		}

		return NilApatSequence, errors.New(message)
	}

	if out == nil {
		// log.Printf("Make ApatSeq called on %p -> %p\n", out, pseqc)
		seq := _ApatSequence{pointer: pseqc, reference: sequence}

		runtime.SetFinalizer(&seq, func(apat_p *_ApatSequence) {
			var errno C.int32_t
			var errmsg *C.char

			if apat_p != nil && apat_p.pointer != nil {
				// log.Debugf("Finaliser called on %p\n", apat_p.pointer)
				C.delete_apatseq(apat_p.pointer, &errno, &errmsg)
			}
		})

		return ApatSequence{&seq}, nil
	}

	recycle[0].pointer.pointer = pseqc
	recycle[0].pointer.reference = sequence

	//log.Println(C.GoString(pseq.cseq))

	return ApatSequence{recycle[0].pointer}, nil
}

// Len method returns the length of the ApatSequence.
func (sequence ApatSequence) Len() int {
	return int(sequence.pointer.pointer.seqlen)
}

// Free method ensure that the C structure wrapped is
// desallocated
func (sequence ApatSequence) Free() {
	var errno C.int32_t
	var errmsg *C.char

	log.Debugf("Free called on %p\n", sequence.pointer.pointer)

	if sequence.pointer != nil && sequence.pointer.pointer != nil {
		C.delete_apatseq(sequence.pointer.pointer,
			&errno, &errmsg)

		runtime.SetFinalizer(sequence.pointer, nil)

		sequence.pointer = nil
	}
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

func (pattern ApatPattern) FindAllIndex(sequence ApatSequence, begin, length int) (loc [][3]int) {
	if begin < 0 {
		begin = 0
	}

	if length < 0 {
		length = sequence.Len()
	}

	nhits := int(C.ManberAll(sequence.pointer.pointer,
		pattern.pointer.pointer,
		0,
		C.int32_t(begin),
		C.int32_t(length+C.MAX_PAT_LEN)))

	if nhits == 0 {
		return nil
	}

	stktmp := (*[1 << 30]int32)(unsafe.Pointer(sequence.pointer.pointer.hitpos[0].val))
	errtmp := (*[1 << 30]int32)(unsafe.Pointer(sequence.pointer.pointer.hiterr[0].val))
	patlen := int(pattern.pointer.pointer.patlen)

	for i := 0; i < nhits; i++ {
		start := int(stktmp[i])
		err := int(errtmp[i])
		//log.Debugln(C.GoString(pattern.pointer.pointer.cpat), start, err)
		loc = append(loc, [3]int{start, start + patlen, err})
	}

	//log.Debugln("------------")
	return loc
}

func (pattern ApatPattern) IsMatching(sequence ApatSequence, begin, length int) bool {
	if begin < 0 {
		begin = 0
	}

	if length < 0 {
		length = sequence.Len()
	}

	nhits := int(C.ManberAll(sequence.pointer.pointer,
		pattern.pointer.pointer,
		0,
		C.int32_t(begin),
		C.int32_t(length+C.MAX_PAT_LEN)))

	return nhits > 0
}

// BestMatch finds the best match of a given pattern in a sequence.
//
// THe function identify the first occurrence of the pattern in the sequence.
// The search can be limited to a portion of the sequence using the begin and
// length parameters to find the next occurrences.
//
// The BestMatch methood ins
// It takes the following parameters:
// - pattern: the pattern to search for (ApatPattern).
// - sequence: the sequence to search in (ApatSequence).
// - begin: the starting index of the search (int).
// - length: the length of the search (int).
//
// It returns the following values:
// - start: the starting index of the best match (int).
// - end: the ending index of the best match (int).
// - nerr: the number of errors in the best match (int).
// - matched: a boolean indicating whether a match was found (bool).
func (pattern ApatPattern) BestMatch(sequence ApatSequence, begin, length int) (start int, end int, nerr int, matched bool) {
	res := pattern.FindAllIndex(sequence, begin, length)

	if len(res) == 0 {
		matched = false
		return
	}

	matched = true

	best := [3]int{0, 0, 10000}
	for _, m := range res {
		if m[2] < best[2] {
			best = m
			log.Debugln(best)
		}
	}

	nerr = best[2]
	end = best[1]

	if best[0] < 0 || best[1] > sequence.Len() {
		matched = false
		return
	}

	if nerr == 0 || !pattern.pointer.pointer.hasIndel {
		start = best[0]
		log.Debugln("No nws ", start, nerr)
		return
	}

	start = best[0] - nerr
	end = best[0] + int(pattern.pointer.pointer.patlen) + nerr
	start = max(start, 0)
	end = min(end, sequence.Len())

	cpattern := (*[1 << 30]byte)(unsafe.Pointer(pattern.pointer.pointer.cpat))
	frg := sequence.pointer.reference.Sequence()[start:end]

	log.Debugln(
		string(frg),
		string((*cpattern)[0:int(pattern.pointer.pointer.patlen)]),
		best[0], nerr, int(pattern.pointer.pointer.patlen),
		sequence.Len(), start, end)

	from, to, score := obialign.LocatePattern(sequence.pointer.reference.Id(),
		(*cpattern)[0:int(pattern.pointer.pointer.patlen)],
		frg)

	// olderr := m[2]

	nerr = score
	start = start + from
	end = start + to
	log.Debugf("BestMatch on %s : score=%d [%d..%d]", sequence.pointer.reference.Id(), score, start, nerr)
	return
}

// FilterBestMatch filters the best non overlapping matches of a given pattern in a sequence.
//
// It takes the following parameters:
// - pattern: the pattern to search for (ApatPattern).
// - sequence: the sequence to search in (ApatSequence).
// - begin: the starting index of the search (int).
// - length: the length of the search (int).
//
// It returns a slice of [3]int representing the locations of all non-overlapping matches in the sequence.
// The two firsts values of the [3]int indicate respectively the start and the end position of
// the match. Following the GO convention the end position is not included in the
// match. The third value indicates the number of error detected for this occurrence.
func (pattern ApatPattern) FilterBestMatch(sequence ApatSequence, begin, length int) (loc [][3]int) {
	res := pattern.FindAllIndex(sequence, begin, length)
	filtered := make([][3]int, 0, len(res))

	best := [3]int{0, 0, 10000}
	for _, m := range res {
		// obilog.Warnf("Current : Begin : %d End : %d Err : %d", m[0], m[1], m[2])
		// obilog.Warnf("Best    : Begin : %d End : %d Err : %d", best[0], best[1], best[2])
		if (m[0] - m[2]) < best[1]+best[2] {
			// match are overlapping
			// log.Warnln("overlap")
			if m[2] < best[2] {
				// log.Warnln("better")
				best = m
			}
		} else if best[2] < 10000 {
			// no overlap
			// log.Warnln("no overlap")
			filtered = append(filtered, best)
			best = m
		}
	}

	if best[2] < 10000 {
		filtered = append(filtered, best)
	}

	if len(filtered) == 0 {
		return nil
	}

	return filtered
}

// tagaacaggctcctctag
// func AllocatedApaSequences() int {
// 	return int(_AllocatedApaSequences)
// }

// AllMatches finds all matches of a given pattern in a sequence.
//
// It only works if the parttern is a pure IUPAC sequence without
// supplementary characters normally allowed : ! and #.
//
// It takes the following parameters:
// - pattern: the pattern to search for (ApatPattern).
// - sequence: the sequence to search in (ApatSequence).
// - begin: the starting index of the search (int).
// - length: the length of the search (int).
//
// It returns a slice of [3]int representing the locations of all matches in the sequence.
// The AllMatches methood returns return a slice of [3]int. The two firsts
// values of the [3]int indicate respectively the start and the end position of
// the match. Following the GO convention the end position is not included in the
// match. The third value indicates the number of error detected for this occurrence.
func (pattern ApatPattern) AllMatches(sequence ApatSequence, begin, length int) (loc [][3]int) {
	res := pattern.FilterBestMatch(sequence, begin, length)

	j := 0
	for _, m := range res {
		// Recompute the start and end position of the match
		// when the pattern allows for indels
		if m[2] > 0 && pattern.pointer.pointer.hasIndel {
			// obilog.Warnf("Locating indel on sequence %s[%s]", sequence.pointer.reference.Id(), pattern.String())
			start := m[0] - m[2]*2
			start = max(start, 0)
			end := start + int(pattern.pointer.pointer.patlen) + 4*m[2]
			end = min(end, sequence.Len())
			// 1 << 30 = 1,073,741,824 = 1Gb
			// It's a virtual array mapping the sequence to the pattern
			// in the C code.
			// No allocations are done here.
			cpattern := (*[1 << 30]byte)(unsafe.Pointer(pattern.pointer.pointer.cpat))
			frg := sequence.pointer.reference.Sequence()[start:end]

			pb, pe, score := obialign.LocatePattern(
				sequence.pointer.reference.Id(),
				(*cpattern)[0:int(pattern.pointer.pointer.patlen)],
				frg)

			// olderr := m[2]
			m[2] = score
			m[0] = start + pb
			m[1] = start + pe

			// obilog.Warnf("seq[%d@%d:%d] %d: %s %d - %s:%s:%s", i, m[0], m[1], olderr, sequence.pointer.reference.Id(), score,
			// 	frg, (*cpattern)[0:int(pattern.pointer.pointer.patlen)], sequence.pointer.reference.Sequence()[m[0]:m[1]])
		}

		if int(pattern.pointer.pointer.maxerr) >= m[2] {
			res[j] = m
			j++
		}
	}
	return res[0:j]
}
