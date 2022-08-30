package dgo

import (
	"runtime/cgo"
)

type hasCallback interface{ call() }

type ExtSlice struct {
	slice     any
	peer      any
	finalizer func(any)
}

func Ext[T any](slice []T, finalizer func(any), peer any) ExtSlice {
	return ExtSlice{slice, peer, finalizer}
}

func (s ExtSlice) asCgoHandle() cgo.Handle { return cgo.NewHandle(s) }
func (s ExtSlice) call() {
	if s.finalizer != nil {
		s.finalizer(s.peer)
	}
}
