package dgo

type MethodCallImplFunc func(*Port, MethodCallFlag, []*Dart_CObject)

type MethodCallId uint32

type MethodCallFlag uint16

func (f MethodCallFlag) IsPinned() bool { return f&0b1 != 0 }

var methodCallMap = map[MethodCallId]MethodCallImplFunc{}

//lint:ignore U1000 Use go:linkname to call this function
func methodCallRegister(funcId MethodCallId, impl MethodCallImplFunc) {
	methodCallMap[funcId] = impl
}

type invokingGoMethod struct {
	payload uint64
	port    *Port
}

func (m invokingGoMethod) id() MethodCallId     { return MethodCallId(m.payload & (1<<32 - 1)) }
func (m invokingGoMethod) flag() MethodCallFlag { return MethodCallFlag(m.payload >> 32) }

func (m invokingGoMethod) specialInt() {}
func (m invokingGoMethod) handleCObjects(objs []*Dart_CObject) {
	methodCallMap[m.id()](m.port, m.flag(), objs)
}
