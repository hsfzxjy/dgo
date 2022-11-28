package internal

import "github.com/hsfzxjy/dgo/tests/test_gen/go/internal/subpack"

//dgo:export
type StructUsingExternalType struct {
	field    subpack.StructInSubpack
}
