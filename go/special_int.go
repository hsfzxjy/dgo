package dgo

type _SpecialIntKind uint64

const (
	sikDartCallback _SpecialIntKind = iota
	sikGoCallback
	sikGoObject
	sikGoMethod
	sikDartFutureCallback
	sikDartStreamCallback
	sikDartCallbackGroup
	sikPreservedGoCall
)

const (
	specialIntMagic uint64 = 0b0111_1111_1111_1
)

type _SpecialInt interface {
	specialInt()
}

type _Serializable interface {
	_SpecialInt
	getPayload() uint64
	getKind() _SpecialIntKind
}

type _Handlable interface {
	_SpecialInt
	handleCObjects(objs []*Dart_CObject)
}

func _buildSpecialInt(kind _SpecialIntKind, port *Port, payload uint64) _SpecialInt {
	switch kind {
	case sikDartCallback:
		return DartCallback{payload, port}
	case sikDartFutureCallback:
		return DartFutureCallback{payload, port}
	case sikDartStreamCallback:
		return DartStreamCallback{payload, port}
	default:
		panic("unreachable")
	}
}

func _buildHandlable(kind _SpecialIntKind, port *Port, payload uint64) _Handlable {
	switch kind {
	case sikGoCallback:
		return invokingGoCallback{payload, port}
	case sikGoMethod:
		return invokingGoMethod{payload, port}
	case sikPreservedGoCall:
		return _PreservedGoCall{payload, port}
	default:
		panic("unreachable")
	}
}

func parseSpecialInt(port *Port, value uint64, handlable bool) (result _SpecialInt, success bool) {
	if value>>(64-13) != specialIntMagic {
		return nil, false
	}
	kind := _SpecialIntKind((value >> 48) & 0b111)
	payload := value & ((1 << 48) - 1)

	if handlable {
		result = _buildHandlable(kind, port, payload)
	} else {
		result = _buildSpecialInt(kind, port, payload)
	}
	return result, true
}

func serialize(kind _SpecialIntKind, payload uint64) uint64 {
	return (specialIntMagic << (64 - 13)) |
		(uint64(kind) << 48) |
		payload
}

func serializeSpecialInt(i _Serializable) uint64 {
	return serialize(i.getKind(), i.getPayload())
}
