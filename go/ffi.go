package dgo

/*
#include "ffi.h"
*/
import "C"
import (
	"fmt"
	"reflect"
	"runtime"
	"runtime/cgo"
	"unsafe"
)

//export dgo__GoFinalizer
func dgo__GoFinalizer(callback_data C.uintptr_t, peer C.uintptr_t) {
	handle := cgo.Handle(peer)
	handle.Value().(hasCallback).call()
	handle.Delete()
}

//export dgo__HandleNativeMessage
func dgo__HandleNativeMessage(port C.Dart_Port_DL, msg *C.Dart_CObject) {
	objs := cobjectUnpack(msg)
	if len(objs) == 0 {
		panic("dgo:go empty cobject array")
	}
	var gcb GoCallback
	if first, ok := objs[0].(int64); !ok {
		panic("dgo:go expect 1-st argument to be int")
	} else {
		gcb = GoCallback(first)
	}
	gcb.handle(objs[1:])
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

const _MAX_ARRAY_LEN = 1<<30 - 1

func cobjectUnpack(cobjArr *C.Dart_CObject) []any {
	pValue := unsafe.Pointer(&cobjArr.value)
	switch cobjArr._type {
	case C.Dart_CObject_kArray:
		break
	case C.Dart_CObject_kInt32:
		return []any{int64(*(*C.int32_t)(pValue))}
	case C.Dart_CObject_kInt64:
		return []any{int64(*(*C.int64_t)(pValue))}
	default:
		panic(fmt.Sprintf("dgo:go expect cobject with kArray, got kind=%d", cobjArr._type))
	}

	pArr := (*C.dgo__Dart_CObject_AsArray)(pValue)
	length := pArr.length
	pCobjs := (*[_MAX_ARRAY_LEN]*C.Dart_CObject)(unsafe.Pointer(pArr.values))

	rets := make([]any, length)
	for i, cobj := range pCobjs[:length] {
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
		pArr := *(**[_MAX_ARRAY_LEN]byte)(pValue)
		length := 0
		for length < _MAX_ARRAY_LEN && pArr[length] != '\x00' {
			length++
		}
		if pArr[length] != '\x00' {
			panic("dgo:go string too long")
		}
		bytes := make([]byte, length)
		copy(bytes[:length], pArr[:length])
		return string(bytes)
	case C.Dart_CObject_kTypedData:
		pStruct := (*C.dgo__Dart_CObject_AsTypedData)(pValue)
		length := pStruct.length
		pAddr := unsafe.Pointer(pStruct.values)
		switch pStruct._type {
		case C.Dart_TypedData_kUint8:
			return unsafe.Slice((*uint8)(pAddr), length)
		default:
			panic(fmt.Sprintf("dgo:go unsupported typed data, kind=%d", pStruct._type))
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
		pValue._type = C.Dart_TypedData_kUint8
		pValue.length = C.intptr_t(header.Len)
		pValue.values = (*C.uint8_t)(unsafe.Pointer(header.Data))
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
		pValue._type = tkind
		pValue.length = C.intptr_t(header.Len)
		pValue.data = (*C.uint8_t)(unsafe.Pointer(header.Data))
		pValue.peer = C.uintptr_t(x.asCgoHandle())
		pValue.callback = C.dgo__pGoFinalizer
	default:
		panic(fmt.Sprintf("dgo:go cannot convert %#[1]v to Dart_CObject", x))
	}
	return
}
