package pin

import (
	"sync/atomic"
	"unsafe"

	"github.com/bits-and-blooms/bitset"
)

const (
	detached uint32 = iota
	accessing
	attached
)

type Meta struct {
	_       noCopy
	flag    atomic.Uint32
	intable bool
	version uint16
	lidcnt  uint16
	lids    *bitset.BitSet
	Extra   any
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
		m.intable = true
		m.lidcnt = 0
		m.lids = bitsetGet()
		pinTable.m.Store(uintptr(m.key()), m)
		m.flag.Store(attached)
		runtime_procUnpin()
		return true
	case attached:
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
	case attached:
		runtime_procPin()
		if !m.flag.CompareAndSwap(flag, accessing) {
			runtime_procUnpin()
			goto LOAD_FLAG
		}
		m.intable = false
		if m.lidcnt == 0 {
			bitsetRecycle(m.lids)
			m.lids = nil
			pinTable.m.Delete(m.key())
			m.flag.Store(detached)
		} else {
			m.flag.Store(attached)
		}
		runtime_procUnpin()
		return true
	}
	return false
}

func (m *Meta) key() uintptr { return uintptr(unsafe.Pointer(m)) }

func (m *Meta) decref(version uint16, lid uint16) {
LOAD_FLAG:
	flag := m.flag.Load()
	switch flag {
	case accessing:
		goto LOAD_FLAG
	case detached:
		return
	case attached:
		runtime_procPin()
		if !m.flag.CompareAndSwap(flag, accessing) {
			runtime_procUnpin()
			goto LOAD_FLAG
		}
		if m.version != version ||
			!m.lids.Test(uint(lid)) {
			m.flag.Store(flag)
			runtime_procUnpin()
			return
		}
		m.lids.Clear(uint(lid))
		m.lidcnt--
		if m.lidcnt == 0 && !m.intable {
			bitsetRecycle(m.lids)
			m.lids = nil
			pinTable.m.Delete(uintptr(m.key()))
			m.flag.Store(detached)
		} else {
			m.flag.Store(attached)
		}
		runtime_procUnpin()
	}
}

//lint:ignore U1000 go:linkname
func metaNewToken(m *Meta) untypedToken {
	var lid uint16
LOAD_FLAG:
	flag := m.flag.Load()
	switch flag {
	case accessing:
		goto LOAD_FLAG
	case detached:
		panic("dgo:go: cannot call NewToken() on an unpinned object")
	case attached:
		runtime_procPin()
		if !m.flag.CompareAndSwap(flag, accessing) {
			runtime_procUnpin()
			goto LOAD_FLAG
		}
		var ok bool
		lid, ok = bitsetFillFirstClear(m.lids)
		if !ok {
			m.flag.Store(flag)
			runtime_procUnpin()
			panic("dgo:go: too many allocated lids")
		}
		m.lidcnt++
		m.flag.Store(attached)
		runtime_procUnpin()
	}
	return newToken[struct{}](m, lid)
}
