package internal

//dgo:export
type StructWithMapField struct {
	Map                 map[String][4]string
	MapWithDynamicValue map[int][]string
}
