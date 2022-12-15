package pchan

import (
	"math/bits"

	dgo "github.com/hsfzxjy/dgo/go"
	"github.com/hsfzxjy/dgo/go/pin/bitset"
	"github.com/hsfzxjy/dgo/go/pin/pcop"
)

type Manager struct {
	backing  imanager
	finished bitset.Bitset64
	heads    []*listener
	lidsets  []bitset.Bitset64
}

func NewManager(n uint8) Manager {
	if n > 64 {
		panic("expect n <= 64")
	}
	i := 8 - bits.LeadingZeros8(n-1)
	return managerPools[i].Get().(imanager).as_chanman(n)
}

func (m *Manager) Recycle() {
	m.backing.recycle()
	m.backing = nil
}

func (m *Manager) do_chan_listen(op pcop.Op) bool {
	Chid := op.Chid
	if m.finished.Test(op.Chid) {
		return false
	}
	m.lidsets[Chid].Set(op.Lid)
	head := &m.heads[Chid]
	listener := listenerNew()
	listener.chid = op.Chid
	listener.lid = op.Lid
	listener.dcb = op.Dcb
	listener.port = (*dgo.Port)(op.Port)
	listener.next = *head
	*head = listener
	return true
}

func (m *Manager) do_chan_cancel_listen(op pcop.Op) {
	m.lidsets[op.Chid].Clear(op.Lid)
}

func (m *Manager) do_token_dispose(op pcop.Op) {
	Lid := op.Lid
	//todo: vectorize
	for i := range m.lidsets {
		m.lidsets[i].Clear(Lid)
	}
}

// For op == CHAN_LISTEN, result == true if the listener was successfully registered;
// for op == META_DETACHED, result is always true;
// for other values of op, result is meaningless.
func (m *Manager) Handle(op pcop.Op) (result bool) {
	switch op.Kind {
	case pcop.CHAN_LISTEN:
		return m.do_chan_listen(op)
	case pcop.CHAN_CANCEL_LISTEN:
		m.do_chan_cancel_listen(op)
	case pcop.TOKEN_DISPOSE:
		m.do_token_dispose(op)
	case pcop.META_DETACHED:
		return true
	}
	return false
}

func (m *Manager) GetLids(finished bool, chid uint8, result *[]dgo.CallableDartCallback) (samePort bool) {
	var visited bitset.Bitset64
	head := m.heads[chid]
	var q *listener
	p := head
	lids := &m.lidsets[chid]

	var port *dgo.Port
	if p != nil {
		port = p.port
	}

	var flag = dgo.CF.Fallible().PackArray().WithContext()
	if finished {
		flag = flag.Pop()
	}

	samePort = true
	for p != nil {
		lid := p.lid
		if visited.Test(lid) {
			goto REMOVE
		}
		visited.Set(lid)
		if !lids.Test(lid) {
			goto REMOVE
		}
		*result = append(*result, dgo.WrapDartCallback(p.dcb, p.port).Flag(flag))
		if p.port != port {
			samePort = false
		}
		q = p
		p = p.next
		continue
	REMOVE:
		switch {
		case q == nil:
			// p is head
			head = p.next
			p.recycle()
			p = head
		default:
			// p is not head
			q.next = p.next
			p.recycle()
			p = q.next
		}
	}

	if finished {
		head.free()
		m.heads[chid] = nil
		m.finished.Set(chid)
	} else {
		m.heads[chid] = head
	}

	return
}
