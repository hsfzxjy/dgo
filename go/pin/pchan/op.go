package pchan

import (
	"sync"

	dgo "github.com/hsfzxjy/dgo/go"
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
	Port *dgo.Port
}

func (op *Op) AsDartCallback() dgo.DartCallback {
	return dgo.WrapDartCallback(op.Dcb, op.Port)
}

var opchanPool = sync.Pool{
	New: func() any { return make(chan Op, 1) },
}

func NewOpChan() chan Op       { return opchanPool.Get().(chan Op) }
func RecycleOpChan(ch chan Op) { opchanPool.Put(ch) }
