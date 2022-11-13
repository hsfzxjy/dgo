package dgo

/*
#include "ffi.h"
*/
import "C"
import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"runtime/cgo"
	"unsafe"
)

type (
	Dart_CObject                     = C.dgo__Dart_CObject
	Dart_CObject_Type                = C.Dart_CObject_Type
	Dart_CObject_AsArray             = C.dgo__Dart_CObject_AsArray
	Dart_CObject_AsTypedData         = C.dgo__Dart_CObject_AsTypedData
	Dart_Cobject_AsExternalTypedData = C.dgo__Dart_Cobject_AsExternalTypedData
)

const (
	Dart_CObject_kNull      = C.Dart_CObject_kNull
	Dart_CObject_kBool      = C.Dart_CObject_kBool
	Dart_CObject_kInt32     = C.Dart_CObject_kInt32
	Dart_CObject_kInt64     = C.Dart_CObject_kInt64
	Dart_CObject_kDouble    = C.Dart_CObject_kDouble
	Dart_CObject_kString    = C.Dart_CObject_kString
	Dart_CObject_kTypedData = C.Dart_CObject_kTypedData
)

func dgo__PostCObjects(arg0 int, arg1 *C.Dart_CObject) bool {
	return bool(C.dgo__PostCObjects(C.int(arg0), arg1))
}

//export dgo__GoFinalizer
func dgo__GoFinalizer(callback_data C.uintptr_t, peer C.uintptr_t) {
	handle := cgo.Handle(peer)
	handle.Value().(hasCallback).call()
	handle.Delete()
}

//export dgo__HandleNativeMessage
func dgo__HandleNativeMessage(port C.Dart_Port_DL, msg *Dart_CObject) {
	first := cobjectUnpackFirst(msg)
	if CallbackFlag(first).hasMethodCall() {
		methodCallInvoke(first, msg)
		return
	}
	objs := cobjectUnpackRest(msg)
	gcb := GoCallback(first)
	gcb.handle(objs)
}

func (dcb CallableDartCallback) Call(args ...any) bool {
	if dcb == illegalCallableDartCallback {
		return false
	}

	cf := CallbackFlag(dcb)
	if cf.HasFast() {
		if len(args) != 0 {
			panic("dgo:go expect no argument when CF_FAST set")
		}
		return bool(C.dgo__PostInt(C.int64_t(dcb)))
	}

	n := len(args)
	cobjs := make([]C.Dart_CObject, n+1)
	cobjectFrom(&cobjs[0], dcb)
	for i, arg := range args {
		args[i] = cobjectFrom(&cobjs[i+1], arg)
	}
	ret := C.dgo__PostCObjects(C.int(n+1), &cobjs[0])
	runtime.KeepAlive(cobjs)
	runtime.KeepAlive(args)
	return bool(ret)
}

const MAX_ARRAY_LEN = 1<<30 - 1

func cobjectUnpackFirst(cobj *Dart_CObject) uint64 {
BEGIN:
	pValue := unsafe.Pointer(&cobj.Value)
	switch cobj.Type {
	case C.Dart_CObject_kArray:
		pArr := (*C.dgo__Dart_CObject_AsArray)(pValue)
		if pArr.Length == 0 {
			panic("dgo:go empty cobject array")
		}
		pCobjs := (*[MAX_ARRAY_LEN]*Dart_CObject)(unsafe.Pointer(pArr.Values))
		cobj = pCobjs[0]
		goto BEGIN
	case C.Dart_CObject_kInt32:
		return uint64(*(*C.int32_t)(pValue))
	case C.Dart_CObject_kInt64:
		return uint64(*(*C.int64_t)(pValue))
	default:
		panic(fmt.Sprintf("dgo:go expect first argument to be int, got kind=%d", cobj.Type))
	}
}

func cobjectUnpackRest(cobjArr *Dart_CObject) []any {
	pValue := unsafe.Pointer(&cobjArr.Value)
	switch cobjArr.Type {
	case C.Dart_CObject_kArray:
		break
	case C.Dart_CObject_kInt32, C.Dart_CObject_kInt64:
		return nil
	default:
		panic(fmt.Sprintf("dgo:go expect cobject with kArray, got kind=%d", cobjArr.Type))
	}

	pArr := (*C.dgo__Dart_CObject_AsArray)(pValue)
	length := pArr.Length
	pCobjs := (*[MAX_ARRAY_LEN]*C.Dart_CObject)(unsafe.Pointer(pArr.Values))

	rets := make([]any, length-1)
	for i, cobj := range pCobjs[1:length] {
		rets[i] = cobjectAs(cobj)
	}

	return rets
}

func cobjectAs(cobj *C.Dart_CObject) any {
	pValue := unsafe.Pointer(&cobj.value)
	switch cobj._type {
	case C.Dart_CObject_kNull:
		return nil
	case C.Dart_CObject_kBool:
		return bool(*(*C.bool)(pValue))
	case C.Dart_CObject_kInt32:
		return int64(*(*C.int32_t)(pValue))
	case C.Dart_CObject_kInt64:
		return int64(*(*C.int64_t)(pValue))
	case C.Dart_CObject_kDouble:
		return float64(*(*C.double)(pValue))
	case C.Dart_CObject_kString:
		pArr := *(**[MAX_ARRAY_LEN]byte)(pValue)
		length := bytes.IndexByte(pArr[:], '\x00')
		if length < 0 || pArr[length] != '\x00' {
			panic("dgo:go string too long")
		}
		bytes := make([]byte, length)
		copy(bytes[:length], pArr[:length])
		return string(bytes)
	case C.Dart_CObject_kTypedData:
		pStruct := (*C.dgo__Dart_CObject_AsTypedData)(pValue)
		length := pStruct.Length
		pAddr := unsafe.Pointer(pStruct.Values)
		switch pStruct.Type {
		case C.Dart_TypedData_kUint8:
			return unsafe.Slice((*uint8)(pAddr), length)
		default:
			panic(fmt.Sprintf("dgo:go unsupported typed data, kind=%d", pStruct.Type))
		}
	default:
		panic(fmt.Sprintf("dgo:go unsupport cobject, kind=%d", cobj._type))
	}
}

func cobjectFrom(cobj *C.Dart_CObject, xx any) (mod any) {
	mod = xx
	if xx == nil {
		cobj._type = C.Dart_CObject_kNull
		return
	}
	switch x := xx.(type) {
	case bool:
		cobj._type = C.Dart_CObject_kBool
		*(*C.bool)(unsafe.Pointer(&cobj.value)) = C.bool(x)
	case int, int8, int16, int32, int64:
		cobj._type = C.Dart_CObject_kInt64
		*(*C.int64_t)(unsafe.Pointer(&cobj.value)) = C.int64_t(reflect.ValueOf(x).Int())
	case uint, uint8, uint16, uint32, uint64, GoCallback, CallableDartCallback:
		cobj._type = C.Dart_CObject_kInt64
		*(*C.int64_t)(unsafe.Pointer(&cobj.value)) = C.int64_t(reflect.ValueOf(x).Uint())
	case float32, float64:
		cobj._type = C.Dart_CObject_kDouble
		*(*C.double)(unsafe.Pointer(&cobj.value)) = C.double(reflect.ValueOf(x).Float())
	case []byte:
		header := (*reflect.SliceHeader)(unsafe.Pointer(&x))
		cobj._type = C.Dart_CObject_kTypedData
		pValue := (*C.dgo__Dart_CObject_AsTypedData)(unsafe.Pointer(&cobj.value))
		pValue.Type = C.Dart_TypedData_kUint8
		pValue.Length = C.intptr_t(header.Len)
		pValue.Values = (*C.uint8_t)(unsafe.Pointer(header.Data))
	case string:
		if len(x) == 0 || x[len(x)-1] != '\x00' {
			x = x + "\x00"
			mod = x
		}
		header := (*reflect.StringHeader)(unsafe.Pointer(&x))
		cobj._type = C.Dart_CObject_kString
		*(*uintptr)(unsafe.Pointer(&cobj.value)) = header.Data
	case ExtSlice:
		var tkind C.Dart_TypedData_Type
		var header *reflect.SliceHeader
		switch slice := x.slice.(type) {
		case []byte:
			tkind = C.Dart_TypedData_kUint8
			header = (*reflect.SliceHeader)(unsafe.Pointer(&slice))
		default:
			panic(fmt.Sprintf("dgo:go cannot use %#v as typed object", x.slice))
		}
		cobj._type = C.Dart_CObject_kExternalTypedData
		pValue := (*C.dgo__Dart_Cobject_AsExternalTypedData)(unsafe.Pointer(&cobj.value))
		pValue.Type = tkind
		pValue.Length = C.intptr_t(header.Len)
		pValue.Data = (*C.uint8_t)(unsafe.Pointer(header.Data))
		pValue.Peer = C.uintptr_t(x.asCgoHandle())
		pValue.Callback = C.dgo__pGoFinalizer
	default:
		panic(fmt.Sprintf("dgo:go cannot convert %#[1]v to Dart_CObject", x))
	}
	return
}
