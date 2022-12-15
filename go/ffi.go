package dgo

/*
#include "ffi.h"
*/
import "C"
import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"runtime"
	"runtime/cgo"
	"sync"
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
	Dart_CObject_kArray     = C.Dart_CObject_kArray
	Dart_CObject_kTypedData = C.Dart_CObject_kTypedData
)

//go:nosplit
func noescape(p *Dart_CObject) *Dart_CObject {
	x := uintptr(unsafe.Pointer(p))
	return (*Dart_CObject)(unsafe.Pointer(x ^ 0))
}

type PortKey C.Dart_Port_DL

func dgo__PostCObjects(port *Port, n int, arr *Dart_CObject) bool {
	return bool(C.dgo__PostCObjects(C.Dart_Port_DL(port.sendPortKey), C.int(n), noescape(arr)))
}

/* Port Methods */

func (p *Port) panicPostFailure() {
	panic("dgo:go: fail to post via " + p.String())
}

func (p *Port) postInt(value int64, raises bool) bool {
	ret := bool(C.dgo__PostInt(C.Dart_Port_DL(p.sendPortKey), C.int64_t(value)))
	if !ret && raises {
		p.panicPostFailure()
	}
	return ret
}

func (p *Port) postCObject(obj *Dart_CObject, raises bool) bool {
	ret := bool(C.dgo__PostCObject(C.Dart_Port_DL(p.sendPortKey), noescape(obj)))
	runtime.KeepAlive(obj)
	if !ret && raises {
		p.panicPostFailure()
	}
	return ret
}

func (p *Port) postCObjects(objs []Dart_CObject, keepAlive any, raises bool) bool {
	var pobjs *Dart_CObject
	if len(objs) != 0 {
		pobjs = &objs[0]
	}
	ret := bool(C.dgo__PostCObjects(
		C.Dart_Port_DL(p.sendPortKey),
		C.int(len(objs)),
		noescape(pobjs)))
	runtime.KeepAlive(objs)
	runtime.KeepAlive(keepAlive)
	if !ret && raises {
		p.panicPostFailure()
	}
	return ret
}

func (p *Port) postCObjects2(objs1, objs2 []Dart_CObject, keepAlive any, raises bool) bool {
	var p1, p2 *Dart_CObject
	if len(objs1) != 0 {
		p1 = &objs1[0]
	}
	if len(objs2) != 0 {
		p2 = &objs2[0]
	}
	ret := bool(C.dgo__PostCObjects2(
		C.Dart_Port_DL(p.sendPortKey),
		C.int(len(objs1)),
		noescape(p1),
		C.int(len(objs2)),
		noescape(p2)))
	runtime.KeepAlive(objs1)
	runtime.KeepAlive(objs2)
	runtime.KeepAlive(keepAlive)
	if !ret && raises {
		p.panicPostFailure()
	}
	return ret
}

func (p *Port) close() {
	if !p.isClosed.CompareAndSwap(false, true) {
		return
	}

	p.postCObject(&Dart_CObject{Type: Dart_CObject_kNull}, false)

	portMap.removePort(p)
	C.dgo__CloseNativePort(C.Dart_Port_DL(p.receivePortKey))
}

/* _PortMap Methods */

func (m *_PortMap) initPort(sendPortKey PortKey) PortKey {
	receivePortKey := C.dgo__InitPort(C.Dart_Port_DL(sendPortKey))
	if receivePortKey == C.ILLEGAL_PORT {
		panic("dgo:go: fail to create a new receive port")
	}
	return PortKey(receivePortKey)
}

/* CallableDartCallback Methods */

func (dcb CallableDartCallback) Call(args ...any) bool {
	cf := CallbackFlag(dcb.payload)
	port := dcb.port
	if cf.HasFast() {
		if len(args) != 0 {
			panic("dgo:go: expect no argument when CF_FAST set")
		}
		return port.postInt(int64(dcb.serialize()), true)
	}

	n := len(args)
	cobjs := make([]Dart_CObject, n+1)
	cobjectLoadFromValue(&cobjs[0], dcb.serialize())
	for i, arg := range args {
		args[i] = cobjectLoadFromValue(&cobjs[i+1], arg)
	}
	return port.postCObjects(cobjs[:n+1], args, true)
}

func (dcb CallableDartCallback) callRaw(args []Dart_CObject) bool {
	var head [1]Dart_CObject
	head[0].Type = Dart_CObject_kInt64
	*(*C.int64_t)(unsafe.Pointer(&head[0].Value)) = C.int64_t(dcb.serialize())
	return dcb.port.postCObjects2(head[:], args, nil, false)
}

/* dartCallbackGroup Methods */

func (g dartCallbackGroup) callRaw(args []Dart_CObject) bool {
	head := make([]Dart_CObject, 1+len(g))
	head[0].Type = Dart_CObject_kInt64
	*(*C.int64_t)(unsafe.Pointer(&head[0].Value)) = C.int64_t(g.serialize())
	for i, dcb := range g {
		s := math.Float64frombits(dcb.serialize())
		head[i+1].Type = Dart_CObject_kDouble
		*(*C.double)(unsafe.Pointer(&head[i+1].Value)) = C.double(s)
	}
	return g[0].port.postCObjects2(head, args, nil, false)
}

/* EXPORTS */

var apiDLOnce sync.Once

//export dgo_InitPort
func dgo_InitPort(apiDLData *C.void, sendPortKey C.Dart_Port_DL, isDefault C.bool) {
	apiDLOnce.Do(func() {
		C.dgo__InitFFI(unsafe.Pointer(apiDLData))
	})
	portMap.addPort(PortKey(sendPortKey), bool(isDefault))
}

//export dgo__GoFinalizer
func dgo__GoFinalizer(callback_data C.uintptr_t, peer C.uintptr_t) {
	handle := cgo.Handle(peer)
	handle.Value().(hasCallback).call()
	handle.Delete()
}

//export dgo__HandleNativeMessage
func dgo__HandleNativeMessage(portKey C.Dart_Port_DL, msg *Dart_CObject) {
	var handler _SpecialInt
	var port *Port
	var success bool

	var firstArg uint64
	var cobjs []*Dart_CObject
	var firstObj = msg

	var atTopLevel = true

SWITCH:
	pValue := unsafe.Pointer(&firstObj.Value)
	switch firstObj.Type {
	case Dart_CObject_kInt32:
		firstArg = uint64(*(*C.int32_t)(pValue))
	case Dart_CObject_kInt64:
		firstArg = uint64(*(*C.int64_t)(pValue))
	case Dart_CObject_kDouble:
		firstArg = math.Float64bits(*(*float64)(pValue))
	case Dart_CObject_kNull:
		if atTopLevel {
			goto CLOSE_PORT
		}
	case Dart_CObject_kArray:
		pValue := (*Dart_CObject_AsArray)(pValue)
		if pValue.Length == 0 {
			panic("dgo:go: received empty array")
		}
		if atTopLevel {
			atTopLevel = false
			cobjs = (*[MAX_ARRAY_LEN]*Dart_CObject)(unsafe.Pointer(pValue.Values))[1:pValue.Length]
			firstObj = (*Dart_CObject)(unsafe.Pointer(*pValue.Values))
			goto SWITCH
		}
		goto BAD_FIRST_ARGUMENT
	}

	if firstArg == 0 {
		return
	}

	port = portMap.GetByPortKey(PortKey(portKey))
	handler, success = parseSpecialInt(port, firstArg, true)
	if !success {
		goto DESERIALIZE_FAILURE
	}
	handler.(_Handlable).handleCObjects(cobjs)
	return

CLOSE_PORT:
	portMap.GetByPortKey(PortKey(portKey)).close()
	return

DESERIALIZE_FAILURE:
	panic(fmt.Sprintf(
		"dgo:go: cannot deserialize the first argument %064b as handler",
		firstArg))

BAD_FIRST_ARGUMENT:
	panic(fmt.Sprintf(
		"dgo:go: bad cobject kind=%d for the first argument",
		firstObj.Type))
}

const MAX_ARRAY_LEN = 1<<30 - 1

/* Dart_COBject Parsing & Serializing */

func cobjectParseInt(cobj *Dart_CObject) int64 {
	pValue := unsafe.Pointer(&cobj.Value)
	switch cobj.Type {
	case C.Dart_CObject_kInt32:
		return int64(*(*C.int32_t)(pValue))
	case C.Dart_CObject_kInt64:
		return int64(*(*C.int64_t)(pValue))
	default:
		panic(fmt.Sprintf("dgo:go: cannot parse cobject into int, kind=%d", cobj.Type))
	}
}

func cobjectParse(port *Port, cobj *Dart_CObject) any {
	pValue := unsafe.Pointer(&cobj.Value)
	switch cobj.Type {
	case C.Dart_CObject_kNull:
		return nil
	case C.Dart_CObject_kBool:
		return bool(*(*C.bool)(pValue))
	case C.Dart_CObject_kInt32:
		return int64(*(*C.int32_t)(pValue))
	case C.Dart_CObject_kInt64:
		return int64(*(*C.int64_t)(pValue))
	case C.Dart_CObject_kDouble:
		value := float64(*(*C.double)(pValue))
		si, success := parseSpecialInt(port, math.Float64bits(value), false)
		if success {
			return si
		} else {
			return value
		}
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
		panic(fmt.Sprintf("dgo:go unsupport cobject, kind=%d", cobj.Type))
	}
}

func cobjectLoadFromValue(cobj *Dart_CObject, xx any) (mod any) {
	mod = xx
	pValue := unsafe.Pointer(&cobj.Value)
	switch x := xx.(type) {
	case nil:
		cobj.Type = C.Dart_CObject_kNull
	case bool:
		cobj.Type = C.Dart_CObject_kBool
		*(*C.bool)(pValue) = C.bool(x)

	// parse integer types
	case int:
		cobj.Type = C.Dart_CObject_kInt64
		*(*C.int64_t)(pValue) = C.int64_t(x)
	case int8:
		cobj.Type = C.Dart_CObject_kInt64
		*(*C.int64_t)(pValue) = C.int64_t(x)
	case int16:
		cobj.Type = C.Dart_CObject_kInt64
		*(*C.int64_t)(pValue) = C.int64_t(x)
	case int32:
		cobj.Type = C.Dart_CObject_kInt64
		*(*C.int64_t)(pValue) = C.int64_t(x)
	case int64:
		cobj.Type = C.Dart_CObject_kInt64
		*(*C.int64_t)(pValue) = C.int64_t(x)
	case uint:
		cobj.Type = C.Dart_CObject_kInt64
		*(*C.int64_t)(pValue) = C.int64_t(x)
	case uint8:
		cobj.Type = C.Dart_CObject_kInt64
		*(*C.int64_t)(pValue) = C.int64_t(x)
	case uint16:
		cobj.Type = C.Dart_CObject_kInt64
		*(*C.int64_t)(pValue) = C.int64_t(x)
	case uint32:
		cobj.Type = C.Dart_CObject_kInt64
		*(*C.int64_t)(pValue) = C.int64_t(x)
	case uint64:
		cobj.Type = C.Dart_CObject_kInt64
		*(*C.int64_t)(pValue) = C.int64_t(x)

	// parse float types
	case float32:
		cobj.Type = C.Dart_CObject_kDouble
		*(*C.double)(pValue) = C.double(canonicalizeNAN(float64(x)))
	case float64:
		cobj.Type = C.Dart_CObject_kDouble
		*(*C.double)(pValue) = C.double(canonicalizeNAN(x))

	case []byte:
		header := (*reflect.SliceHeader)(unsafe.Pointer(&x))
		cobj.Type = C.Dart_CObject_kTypedData
		pValue := (*C.dgo__Dart_CObject_AsTypedData)(pValue)
		pValue.Type = C.Dart_TypedData_kUint8
		pValue.Length = C.intptr_t(header.Len)
		pValue.Values = (*C.uint8_t)(unsafe.Pointer(header.Data))
	case string:
		if len(x) == 0 || x[len(x)-1] != '\x00' {
			x = x + "\x00"
			mod = x
		}
		header := (*reflect.StringHeader)(unsafe.Pointer(&x))
		cobj.Type = C.Dart_CObject_kString
		*(*uintptr)(pValue) = header.Data
	case ExtSlice:
		var tkind C.Dart_TypedData_Type
		var header *reflect.SliceHeader
		switch slice := x.slice.(type) {
		case []byte:
			tkind = C.Dart_TypedData_kUint8
			header = (*reflect.SliceHeader)(unsafe.Pointer(&slice))
		default:
			panic(fmt.Sprintf("dgo:go: cannot use %#v as typed object", x.slice))
		}
		cobj.Type = C.Dart_CObject_kExternalTypedData
		pValue := (*C.dgo__Dart_Cobject_AsExternalTypedData)(pValue)
		pValue.Type = tkind
		pValue.Length = C.intptr_t(header.Len)
		pValue.Data = (*C.uint8_t)(unsafe.Pointer(header.Data))
		pValue.Peer = C.uintptr_t(x.asCgoHandle())
		pValue.Callback = C.dgo__pGoFinalizer
	case _Serializable:
		cobj.Type = Dart_CObject_kDouble
		serialized := math.Float64frombits(serializeSpecialInt(x))
		*(*C.double)(pValue) = C.double(serialized)
	default:
		panic(fmt.Sprintf("dgo:go cannot convert %#[1]v to Dart_CObject", x))
	}
	return
}
