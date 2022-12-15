package dgo

type dartCallbackGroup []CallableDartCallback

func (dartCallbackGroup) specialInt()              {}
func (dartCallbackGroup) getKind() _SpecialIntKind { return sikDartCallbackGroup }
func (g dartCallbackGroup) getPayload() uint64     { return uint64(len(g)) }
func (g dartCallbackGroup) serialize() uint64      { return serialize(g.getKind(), g.getPayload()) }

//lint:ignore U1000 go:linkname
func callbackGroupCall(callbacks []CallableDartCallback, samePort bool, args []Dart_CObject) bool {
	switch len(callbacks) {
	case 0:
		return true
	case 1:
		return callbacks[0].callRaw(args)
	default:
	}
	if samePort {
		g := dartCallbackGroup(callbacks)
		return g.callRaw(args)
	}
	panic("todo: unimplemented")
}
