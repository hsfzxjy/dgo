package pin

import (
	"math/bits"
	"sync/atomic"
	"unsafe"
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
	_pad    [2]byte
	lids    uint64
	ops     chan chanop
}

func (m *Meta) lidcnt() int { return bits.OnesCount64(m.lids) }
func (m *Meta) lidtest(lid uint8) bool {
	return lid < 64 && m.lids&(uint64(1)<<lid) != 0
}
func (m *Meta) lidclear(lid uint8) { m.lids &^= 1 << lid }
func (m *Meta) lidset() (lid uint8, success bool) {
	const full = ^uint64(0)
	lids := m.lids
	if lids == full {
		return 0, false
	}
	x := bits.TrailingZeros64(^lids)
	m.lids |= 1 << x
	return uint8(x), true
}

func (m *Meta) Pin() (success bool) {
LOAD_FLAG:
	flag := m.flag.Load()
	switch flag {
	case accessing:
		goto LOAD_FLAG
	case detached:
		runtime_procPin()
		if !m.flag.CompareAndSwap(flag, accessing) {
			runtime_procUnpin()
			goto LOAD_FLAG
		}
		m.version = uint16(pinTable.nextVersion.Add(1))
		m.lids = 0
		pinTable.m.Store(uintptr(m.key()), m)
		m.flag.Store(attached_intable)
		runtime_procUnpin()
		return true
	case attached_not_intable:
		if !m.flag.CompareAndSwap(flag, attached_intable) {
			goto LOAD_FLAG
		}
	}
	return false
}

func (m *Meta) Unpin() (success bool) {
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
		if m.lidcnt() == 0 {
			m.lids = 0
			pinTable.m.Delete(m.key())
			m.flag.Store(detached)
		} else {
			m.flag.Store(attached_not_intable)
		}
		runtime_procUnpin()
		return true
	}
	return false
}

func (m *Meta) key() uintptr { return uintptr(unsafe.Pointer(m)) }

func (m *Meta) decref(version uint16, lid uint8) {
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
			!m.lidtest(lid) {
			m.flag.Store(flag)
			runtime_procUnpin()
			return
		}
		m.lidclear(lid)
		if m.lidcnt() == 0 && flag&intable == 0 {
			m.lids = 0
			pinTable.m.Delete(uintptr(m.key()))
			m.flag.Store(detached)
		} else {
			m.flag.Store(flag)
		}
		runtime_procUnpin()
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
		lid, ok = m.lidset()
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
