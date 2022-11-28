package internal

import "github.com/hsfzxjy/dgo/tests/test_gen/go/internal/subpack"

//dgo:export
type StructWithPointerToSingleType struct {
	ptrToInt      *int
	ptrToString   *string
	ptrToPtrToInt **int
}

//dgo:export
type StructWithPointerToCompoundTypes struct {
	ptrToStruct            *StructWithSimpleTypes
	ptrToPtrToStruct       **StructWithSimpleTypes
	ptrToArray             *[8]int
	ptrToPtrToArray        **[8]int
	ptrToExternalType      *subpack.StructInSubpack
	ptrToPtrToExternalType **subpack.StructInSubpack
}

//dgo:export
type ArrayWithPointerToSimpleType [8]*int

//dgo:export
type ArrayWithPointerToCompoundType [8]*subpack.StructInSubpack