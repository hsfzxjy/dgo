package pcop

import (
	"sync"
	"unsafe"
)

type OpKind uint8

const (
	CHAN_LISTEN OpKind = iota
	CHAN_CANCEL_LISTEN
	TOKEN_DISPOSE
	META_DETACHED
)

type Op struct {
	Kind OpKind
	Chid uint8
	Lid  uint8
	Dcb  uint32
	Port unsafe.Pointer
}

var opchanPool = sync.Pool{
	New: func() any { return make(chan Op, 1) },
}

func NewOpChan() chan Op       { return opchanPool.Get().(chan Op) }
func RecycleOpChan(ch chan Op) { opchanPool.Put(ch) }
