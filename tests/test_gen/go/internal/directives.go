package internal

//dgo:export
type StructWithNoDartField struct {
	field       string `dgo:",!dart"`
	normalField int
}

//dgo:export
type StructWithNoGoField struct {
	field string `dgo:",!go"`
}

//dgo:export
type StructWithRenamedField struct {
	renamedField       string `dgo:"Field"`
	renamedNoDartField string `dgo:"FieldNoGo,!dart"`
	renamedNoGoField   string `dgo:"FieldNoDart,!go"`
}
