package cutils

import "C"

import (
	"reflect"
	"unsafe"
)

func ByteSlice(pointer unsafe.Pointer, size int) []byte {
	var s []byte

	h := (*reflect.SliceHeader)((unsafe.Pointer(&s)))
	h.Cap = size
	h.Len = size
	h.Data = uintptr(pointer)

	return s
}
