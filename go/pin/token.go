package pin

import (
	"sync"
	"unsafe"

	"github.com/hsfzxjy/dgo/go/pin/pcop"
)

type rawToken struct {
	version uint16
	lid     uint8
	_pad    [1]byte
	meta    *Meta
}

type Token[T any] struct {
	*rawToken
	//lint:ingore U1000 unexportable marker
	_ struct{}
}

var rawTokenPool = sync.Pool{
	New: func() any { return new(rawToken) },
}

func newToken(meta *Meta, version uint16, lid uint8) untypedToken {
	rt := rawTokenPool.Get().(*rawToken)
	rt.version = version
	rt.meta = meta
	rt.lid = lid
	return Token[struct{}]{rawToken: rt}
}

// t should be dropped after Dispose() invoked
func (t *Token[T]) Dispose() (success bool) {
	if t.IsEmpty() {
		return false
	}
	t.meta.decref(t.version, t.lid)
	untypedTokenLeak(untypedToken(*t))
	return true
}

func (t Token[T]) Data() *T {
	if t.IsEmpty() {
		panic("dgo:go: Data() called on an empty Token")
	}
	return (*T)(unsafe.Pointer(t.meta))
}

func (t *Token[T]) IsEmpty() bool { return t.rawToken == nil || t.rawToken.meta == nil }

type untypedToken = Token[struct{}]

//lint:ignore U1000 go:linkname
func untypedTokenFromRaw(version uint16, lid uint8, data uintptr) (ret untypedToken) {
	meta, ok := pinTable.m.Load(data)
	if !ok {
		return
	}

LOAD_FLAG:
	flag := meta.flag.Load()
	switch flag {
	case accessing:
		goto LOAD_FLAG
	case attached_not_intable, attached_intable:
		runtime_procPin()
		if !meta.flag.CompareAndSwap(flag, accessing) {
			runtime_procUnpin()
			goto LOAD_FLAG
		}
		if meta.version == version &&
			meta.lids.Test(lid) {
			ret = newToken(meta, version, lid)
		}
		// else: the version is mismatched or lid is invalid, we return an empty token
		meta.flag.Store(flag)
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
	token.lid = 0
	token.meta = nil
	rawTokenPool.Put(token.rawToken)
}

//lint:ignore U1000 go:linkname
func untypedTokenExtract(token untypedToken) (version uint16, lid uint8, data uintptr) {
	version = token.version
	lid = token.lid
	data = uintptr(unsafe.Pointer(token.meta))
	return version, lid, data
}

//go:linkname pin_TokenDispose github.com/hsfzxjy/dgo/go.pin_TokenDispose
func pin_TokenDispose(version uint16, lid uint8, data uintptr) {
	token := untypedTokenFromRaw(uint16(version), uint8(lid), uintptr(data))
	token.Dispose()
}

//go:linkname pin_ChanListen github.com/hsfzxjy/dgo/go.pin_ChanListen
func pin_ChanListen(version uint16, lid uint8, data uintptr, chid uint8, dcb uint32, port unsafe.Pointer) {
	token := untypedTokenFromRaw(uint16(version), uint8(lid), uintptr(data))
	if token.IsEmpty() {
		return
	}
	token.meta.ops <- pcop.Op{Kind: pcop.CHAN_LISTEN, Lid: lid, Chid: chid, Dcb: dcb, Port: port}
}

//go:linkname pin_ChanCancelListen github.com/hsfzxjy/dgo/go.pin_ChanCancelListen
func pin_ChanCancelListen(version uint16, lid uint8, data uintptr, chid uint8) {
	token := untypedTokenFromRaw(uint16(version), uint8(lid), uintptr(data))
	if token.IsEmpty() {
		return
	}
	token.meta.ops <- pcop.Op{Kind: pcop.CHAN_CANCEL_LISTEN, Lid: lid, Chid: chid}
}

var _ = pin_ChanCancelListen
