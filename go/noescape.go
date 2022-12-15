package dgo

import "unsafe"

//go:nosplit
func noescape(p *Dart_CObject) *Dart_CObject {
	x := uintptr(unsafe.Pointer(p))
	return (*Dart_CObject)(unsafe.Pointer(x ^ 0))
}
