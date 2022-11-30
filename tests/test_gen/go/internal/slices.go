package internal

import dgo "github.com/hsfzxjy/dgo/go"

//dgo:export
type StructWithSliceField struct {
	Slice          []int
	StringField    string
	NestedSlice    [][]int
	NestedSlice2   [][2][]int
	OptionalSlice  dgo.Optional[[]int]
	PointerToSlice *[]string
}

//dgo:export
type SliceAliased []string
