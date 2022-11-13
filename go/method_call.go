package dgo

import "unsafe"

type MethodCallImplFunc func([]*Dart_CObject)

type MethodCallId uint32

var methodCallMap map[MethodCallId]MethodCallImplFunc

//lint:ignore U1000 Use go:linkname to call this function
func methodCallRegister(funcId MethodCallId, impl MethodCallImplFunc) {
	methodCallMap[funcId] = impl
}

func methodCallInvoke(first uint64, msg *Dart_CObject) {
	pValue := unsafe.Pointer(&msg.Value)
	pArr := (*Dart_CObject_AsTypedData)(pValue)
	arr := (*[MAX_ARRAY_LEN]*Dart_CObject)(unsafe.Pointer(pArr.Values))[1:]
	methodId := MethodCallId(first)
	methodCallMap[methodId](arr)
}
