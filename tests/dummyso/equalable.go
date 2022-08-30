package main

type Equalable interface {
	Equals(any) bool
	SendToDart(func(...any))
}
