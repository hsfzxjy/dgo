package codec

import (
	"unsafe"

	dgo "github.com/hsfzxjy/dgo/go"
)

type integer interface {
	~int8 | ~int16 | ~int32 | ~int64 | ~int |
		~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uint | ~uintptr
}

func DecodeInteger[T integer](obj *dgo.Dart_CObject) T {
	if obj.Type == dgo.Dart_CObject_kInt32 {
		return T(*(*int32)(unsafe.Pointer(&obj.Value)))
	} else {
		return T(*(*int64)(unsafe.Pointer(&obj.Value)))
	}
}

func DecodeUint64(obj *dgo.Dart_CObject) uint64 {
	return uint64(*(*int64)(unsafe.Pointer(&obj.Value)))
}

func EncodeInteger(value int64, obj *dgo.Dart_CObject) {
	obj.Type = dgo.Dart_CObject_kInt64
	*(*int64)(unsafe.Pointer(&obj.Value)) = value
}

type float interface {
	~float32 | ~float64
}

func DecodeFloat[T float](obj *dgo.Dart_CObject) T {
	return T(*(*float64)(unsafe.Pointer(&obj.Value)))
}

func EncodeFloat(value float64, obj *dgo.Dart_CObject) {
	obj.Type = dgo.Dart_CObject_kDouble
	*(*float64)(unsafe.Pointer(&obj.Value)) = value
}

func DecodeBool(obj *dgo.Dart_CObject) bool {
	return *(*bool)(unsafe.Pointer(&obj.Value))
}

func EncodeBool(value bool, obj *dgo.Dart_CObject) {
	obj.Type = dgo.Dart_CObject_kBool
	*(*bool)(unsafe.Pointer(&obj.Value)) = value
}
