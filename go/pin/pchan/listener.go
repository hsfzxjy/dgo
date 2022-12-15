package pchan

import (
	"sync"

	dgo "github.com/hsfzxjy/dgo/go"
)

type listener struct {
	chid uint8
	lid  uint8
	dcb  uint32
	port *dgo.Port

	next *listener
}

func (l *listener) free() {
	p := l
	for p != nil {
		q := p.next
		p.recycle()
		p = q
	}
}

var listenerPool = sync.Pool{
	New: func() any { return new(listener) },
}

func listenerNew() *listener { return listenerPool.Get().(*listener) }
func (l *listener) recycle() {
	l.next = nil
	l.port = nil
	listenerPool.Put(l)
}
