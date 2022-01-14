package cutils

import "C"

import (
	"reflect"
	"unsafe"
)

// ByteSlice builds a byte Slice over a C pointer returned by a
// C function.
// The C pointer has to be passed as an unsafe.Pointer
// and the size of the pointed memery area has to be
// indicated through the size parameter.
// No memory allocation is done by that function. It
// just consists in mapping a Go object too a C one.
func ByteSlice(pointer unsafe.Pointer, size int) []byte {
	var s []byte

	h := (*reflect.SliceHeader)((unsafe.Pointer(&s)))
	h.Cap = size
	h.Len = size
	h.Data = uintptr(pointer)

	return s
}
