package internal

//dgo:export ,enum
type MyEnum uint64

const (
	ME_UNKNOWN MyEnum = 1 << (63 - iota) //dgo:export unknown
	ME_A                                 //dgo:export -
	ME_B                                 //dgo:export b
)

const B = 1
