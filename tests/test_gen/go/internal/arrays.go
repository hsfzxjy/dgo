package internal

import (
	dgo "github.com/hsfzxjy/dgo/go"
	"github.com/hsfzxjy/dgo/tests/test_gen/go/internal/subpack"
)

//dgo:export
type ArrayPositiveSized [8]string

//dgo:export
type ArrayZeroSized [0]int

//dgo:export
type ArrayNested [8][9]string

//dgo:export
type ArrayOfOptional [8]dgo.Optional[int]

//dgo:export
type ArrayOfStruct [8]subpack.StructInSubpack
