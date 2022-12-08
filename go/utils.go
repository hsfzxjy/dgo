package dgo

import _ "unsafe"

//go:linkname runtime_procPin sync/atomic.runtime_procPin
func runtime_procPin()

//go:linkname runtime_procUnpin sync/atomic.runtime_procUnpin
func runtime_procUnpin()
