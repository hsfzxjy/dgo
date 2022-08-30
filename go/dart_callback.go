package dgo

import (
	"sync/atomic"
)

const (
	dartVersionBits         = 16
	dartVersionDelta uint64 = 1 << (64 - dartVersionBits)
	dartVersionMask  uint64 = ^uint64(0) - (dartVersionDelta - 1)
)

var dartVersion uint64 = 0

func dartVersionInc() {
	atomic.AddUint64(&dartVersion, dartVersionDelta)
}

type RawDartCallback interface{ ~uint32 }

type DartCallback uint64

func WrapDartCallback[T RawDartCallback](rcb T) DartCallback {
	return DartCallback(atomic.LoadUint64(&dartVersion) + uint64(rcb))
}

func (dcb DartCallback) IsIllegal() bool {
	return atomic.LoadUint64(&dartVersion) != (uint64(dcb) & dartVersionMask)
}

func (dcb DartCallback) Flag(cf CallbackFlag) CallableDartCallback {
	if dcb.IsIllegal() {
		return illegalCallableDartCallback
	}
	return CallableDartCallback((uint64(dcb) & ^dartVersionMask) | uint64(cf))
}

type CallableDartCallback uint64

func (dcb CallableDartCallback) IsIllegal() bool {
	return dcb == illegalCallableDartCallback
}

const (
	illegalCallableDartCallback = ^CallableDartCallback(0)
)
