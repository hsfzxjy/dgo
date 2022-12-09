package pin

import (
	"sync/atomic"

	"github.com/puzpuzpuz/xsync/v2"
)

type _pinTable struct {
	m           *xsync.MapOf[uintptr, *Meta]
	nextVersion atomic.Uint32
}

var pinTable _pinTable

func init() {
	pinTable.m = xsync.NewIntegerMapOf[uintptr, *Meta]()
}
