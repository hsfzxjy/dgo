package pin

import (
	"sync/atomic"
	"unsafe"

	"github.com/hsfzxjy/dgo/go/pin/bitset"
	"github.com/hsfzxjy/dgo/go/pin/pcop"
)

const (
	detached uint32 = iota
	accessing
	attached_not_intable

	intable uint32 = 0x1000_0000

	attached_intable = intable | attached_not_intable
)

type Meta struct {
	_       noCopy
	flag    atomic.Uint32
	version uint16
	nchans  uint8
	_pad    [1]byte
	lids    bitset.Bitset64
	ops     chan pcop.Op
}

//lint:ignore U1000 go:linkname
func metaPin(m *Meta, nchans uint8, workerfn func(chan pcop.Op)) (success bool) {
LOAD_FLAG:
	flag := m.flag.Load()
	switch flag {
	case accessing:
		goto LOAD_FLAG
	case detached:
		var ops chan pcop.Op
		runtime_procPin()
		if !m.flag.CompareAndSwap(flag, accessing) {
			runtime_procUnpin()
			goto LOAD_FLAG
		}
		m.version = uint16(pinTable.nextVersion.Add(1))
		m.lids = 0
		m.nchans = nchans
		if nchans > 0 {
			ops = pcop.NewOpChan()
			m.ops = ops
		}
		pinTable.m.Store(uintptr(m.key()), m)
		m.flag.Store(attached_intable)
		runtime_procUnpin()

		if nchans > 0 {
			go workerfn(ops)
		}

		return true
	case attached_not_intable:
		if !m.flag.CompareAndSwap(flag, attached_intable) {
			goto LOAD_FLAG
		}
	}
	return false
}

func (m *Meta) Unpin() (success bool) {
	var ops chan pcop.Op
LOAD_FLAG:
	flag := m.flag.Load()
	switch flag {
	case accessing:
		goto LOAD_FLAG
	case detached:
	case attached_not_intable, attached_intable:
		runtime_procPin()
		if !m.flag.CompareAndSwap(flag, accessing) {
			runtime_procUnpin()
			goto LOAD_FLAG
		}
		if m.lids.IsEmpty() {
			m.lids = 0
			pinTable.m.Delete(m.key())
			if m.nchans > 0 {
				ops = m.ops
				m.ops = nil
			}
			m.flag.Store(detached)
		} else {
			m.flag.Store(attached_not_intable)
		}
		runtime_procUnpin()

		if ops != nil {
			ops <- pcop.Op{Kind: pcop.META_DETACHED}
		}

		return true
	}
	return false
}

func (m *Meta) key() uintptr { return uintptr(unsafe.Pointer(m)) }

func (m *Meta) decref(version uint16, lid uint8) {
	var ops chan pcop.Op
LOAD_FLAG:
	flag := m.flag.Load()
	switch flag {
	case accessing:
		goto LOAD_FLAG
	case detached:
		return
	case attached_not_intable, attached_intable:
		runtime_procPin()
		if !m.flag.CompareAndSwap(flag, accessing) {
			runtime_procUnpin()
			goto LOAD_FLAG
		}
		if m.version != version ||
			!m.lids.Test(lid) {
			m.flag.Store(flag)
			runtime_procUnpin()
			return
		}
		m.lids.Clear(lid)
		if m.lids.IsEmpty() && flag&intable == 0 {
			m.lids = 0
			pinTable.m.Delete(uintptr(m.key()))
			if m.nchans > 0 {
				ops = m.ops
				m.ops = nil
			}
			m.flag.Store(detached)
		} else {
			m.flag.Store(flag)
		}
		runtime_procUnpin()
	}
	if ops != nil {
		ops <- pcop.Op{Kind: pcop.META_DETACHED}
	}
}

//lint:ignore U1000 go:linkname
func metaNewToken(m *Meta) untypedToken {
	var lid uint8
	var version uint16
LOAD_FLAG:
	flag := m.flag.Load()
	switch flag {
	case accessing:
		goto LOAD_FLAG
	case detached:
		panic("dgo:go: cannot call NewToken() on an unpinned object")
	case attached_intable, attached_not_intable:
		runtime_procPin()
		if !m.flag.CompareAndSwap(flag, accessing) {
			runtime_procUnpin()
			goto LOAD_FLAG
		}
		var ok bool
		lid, ok = m.lids.PickSet()
		if !ok {
			m.flag.Store(flag)
			runtime_procUnpin()
			panic("dgo:go: too many allocated lids")
		}
		version = m.version
		m.flag.Store(flag)
		runtime_procUnpin()
	}
	return newToken(m, version, lid)
}
