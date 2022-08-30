package main

import (
	"github.com/google/go-cmp/cmp"
)

type extSliceWrapper[T any] struct {
	value []T
}

func (w extSliceWrapper[T]) Equals(x any) bool {
	var xx []T
	xx, ok := x.([]T)
	if !ok {
		return false
	}
	return cmp.Equal(w.value, xx)
}
func (w extSliceWrapper[T]) SendToDart(f func(...any)) { f(w.value) }

func eslice[T any](value []T) extSliceWrapper[T] { return extSliceWrapper[T]{value} }
