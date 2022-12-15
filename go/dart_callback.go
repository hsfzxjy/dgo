package dgo

const (
	callbackIdBits = 32
	callbackIdMask = (uint64(1) << 32) - 1
)

type rawDartCallback interface{ ~uint32 }

type DartCallback struct {
	id   uint64 // Id (32 bits)
	port *Port
}

func (dcb DartCallback) specialInt() {}

func WrapDartCallback[T rawDartCallback](rcb T, port *Port) DartCallback {
	port = portMap.ResolvePort(port)
	return DartCallback{uint64(rcb), port}
}

func (dcb DartCallback) Flag(cf CallbackFlag) CallableDartCallback {
	if dcb.port == nil {
		panic("dgo:go: dart callback with zero value found, maybe uninitialized")
	}
	return CallableDartCallback{uint64(cf) | dcb.id, dcb.port}
}

type CallableDartCallback struct {
	payload uint64 // Flag (16 bits) | Id (32 bits)
	port    *Port
}

func (dcb CallableDartCallback) specialInt()              {}
func (dcb CallableDartCallback) getPayload() uint64       { return dcb.payload }
func (dcb CallableDartCallback) getKind() _SpecialIntKind { return sikDartCallback }
func (dcb CallableDartCallback) serialize() uint64        { return serialize(dcb.getKind(), dcb.getPayload()) }
