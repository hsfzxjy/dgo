package dgo

type MethodCallImplFunc func([]*Dart_CObject)

type MethodCallId uint32

var methodCallMap map[MethodCallId]MethodCallImplFunc

//lint:ignore U1000 Use go:linkname to call this function
func methodCallRegister(funcId MethodCallId, impl MethodCallImplFunc) {
	methodCallMap[funcId] = impl
}

type invokingGoMethod struct {
	id   MethodCallId
	port *Port
}

func (m invokingGoMethod) specialInt() {}
func (m invokingGoMethod) handleCObjects(objs []*Dart_CObject) {
	methodCallMap[m.id](objs)
}
