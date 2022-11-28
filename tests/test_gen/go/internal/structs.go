package internal

import dgo "github.com/hsfzxjy/dgo/go"

// lint:file-ignore U1000 No reasons.

//dgo:export
type StructWithSimpleTypes struct {
	fieldInt     int
	fieldInt8    int8
	fieldInt16   int16
	fieldInt32   int32
	fieldInt64   int64
	fieldUint    uint
	fieldUint8   uint8
	fieldUint16  uint16
	fieldUint32  uint32
	fieldUint64  uint64
	fieldUintptr uintptr

	fieldFloat32 float32
	fieldFloat64 float64

	fieldBool bool

	fieldString string
}

//dgo:export
type StructWithArrayFields struct {
	fieldIntArray        [8]int
	fieldZeroLengthArray [0]int
}

//dgo:export
type StructWithOptionalFields struct {
	fieldOptionalInt dgo.Optional[int]
}

//dgo:export
type StructEmpty struct{}

//dgo:export
type StructwithAliasedSimpleTypeField struct {
	fieldUint Uint
}
