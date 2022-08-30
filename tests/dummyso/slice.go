package main

import (
	"github.com/google/go-cmp/cmp"
)

type sliceWrapper[T any] struct {
	value []T
}

func (w sliceWrapper[T]) Equals(x any) bool {
	var xx []T
	xx, ok := x.([]T)
	if !ok {
		return false
	}
	return cmp.Equal(w.value, xx)
}
func (w sliceWrapper[T]) SendToDart(f func(...any)) { f(w.value) }

func slice[T any](value []T) sliceWrapper[T] { return sliceWrapper[T]{value} }
