package dgo

import "C"

//export dgo_InitGo
func dgo_InitGo() {
	dartVersionInc()
}
