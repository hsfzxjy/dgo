package dgo

//#include <stdint.h>
import "C"
import (
	"sync"
	"unsafe"
)

type rawToken struct {
	version uint16
	_pad    [4]byte
	meta    *PinMeta
}

type PinToken[T any] struct {
	*rawToken
	//lint:ingore U1000 unexportable marker
	_ struct{}
}

var rawTokenPool sync.Pool

func newToken[T any](meta *PinMeta) PinToken[T] {
	rt, ok := rawTokenPool.Get().(*rawToken)
	if !ok {
		rt = new(rawToken)
	}
	rt.version = meta.version
	rt.meta = meta
	return PinToken[T]{rawToken: rt}
}

// t should be dropped after Dispose() invoked
func (t *PinToken[T]) Dispose() (success bool) {
	if t.IsEmpty() {
		return false
	}
	t.meta.decref(t.version)
	untypedTokenLeak(untypedToken(*t))
	return true
}

func (t PinToken[T]) Data() *T {
	if t.IsEmpty() {
		panic("dgo:go: Data() called on an empty Token")
	}
	return (*T)(unsafe.Pointer(t.meta))
}

func (t *PinToken[T]) IsEmpty() bool { return t.rawToken == nil || t.rawToken.meta == nil }

type untypedToken = PinToken[struct{}]

//lint:ignore U1000 go:linkname
func untypedTokenFromRaw(version uint16, data uintptr) (ret untypedToken) {
	meta, ok := pinTable.m.Load(data)
	if !ok {
		return
	}

LOAD_FLAG:
	flag := meta.flag.Load()
	switch flag {
	case accessing:
		goto LOAD_FLAG
	case attached:
		runtime_procPin()
		if !meta.flag.CompareAndSwap(flag, accessing) {
			runtime_procUnpin()
			goto LOAD_FLAG
		}
		defer meta.flag.Store(flag)
		if meta.version == version {
			ret = newToken[struct{}](meta)
			ret.version = version
		}
		// else: the version is mismatched, we return an empty token
		runtime_procUnpin()
		return
	case detached:
		return
	}

	return
}

//lint:ignore U1000 go:linkname
func untypedTokenLeak(token untypedToken) {
	token.version = 0
	token.meta = nil
	rawTokenPool.Put(token.meta)
}

//lint:ignore U1000 go:linkname
func untypedTokenExtract(token untypedToken) (version uint16, data uintptr) {
	version = token.version
	data = uintptr(unsafe.Pointer(token.meta))
	return version, data
}

//export dgo_DisposeToken
func dgo_DisposeToken(version C.uint16_t, data C.uintptr_t) {
	token := untypedTokenFromRaw(uint16(version), uintptr(data))
	token.Dispose()
}
