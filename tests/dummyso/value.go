package main

import (
	"reflect"
)

type valueWrapper[T, R comparable] struct {
	value  T
	target R
}

func (w valueWrapper[T, R]) Equals(x any) bool {
	if xx, ok := x.(R); !ok {
		return false
	} else {
		return w.target == xx
	}
}

func (w valueWrapper[T, R]) SendToDart(f func(...any)) { f(w.value) }

func as[T, R any](x T) R {
	var _dummy R
	RType := reflect.TypeOf(_dummy)
	y := reflect.ValueOf(x).Convert(RType)
	return y.Interface().(R)
}

func vs[T comparable](x T) valueWrapper[T, T] { return valueWrapper[T, T]{x, x} }
func vc[T, R comparable](x T) valueWrapper[T, R] {
	return valueWrapper[T, R]{x, as[T, R](x)}
}

type tupleWrapper[T1, T2, R1, R2 comparable] struct {
	value1  T1
	value2  T2
	target1 R1
	target2 R2
}

func (w tupleWrapper[T1, T2, R1, R2]) Equals(xs any) bool {
	arr := xs.([]any)
	x1, x2 := arr[0], arr[1]
	if xx1, ok := x1.(R1); !ok {
		return false
	} else if xx2, ok := x2.(R2); !ok {
		return false
	} else {
		return w.target1 == xx1 && w.target2 == xx2
	}
}
func (w tupleWrapper[T1, T2, R1, R2]) SendToDart(f func(...any)) { f(w.value1, w.value2) }

func ts[T1, T2 comparable](x1 T1, x2 T2) tupleWrapper[T1, T2, T1, T2] {
	return tupleWrapper[T1, T2, T1, T2]{x1, x2, x1, x2}
}
func tc[T1, T2, R1, R2 comparable](x1 T1, x2 T2) tupleWrapper[T1, T2, R1, R2] {
	return tupleWrapper[T1, T2, R1, R2]{x1, x2, as[T1, R1](x1), as[T2, R2](x2)}
}
