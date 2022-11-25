package dgo

import (
	"fmt"
	"reflect"
	"sync/atomic"

	xsync "github.com/puzpuzpuz/xsync/v2"
)

type Port struct {
	sendPortKey    PortKey
	receivePortKey PortKey
	goCallbacks    *xsync.MapOf[uint64, reflect.Value]
	nextCallbackId atomic.Uint32
	isClosed       atomic.Bool
}

func newPort(sendPortKey PortKey, receivePortKey PortKey) *Port {
	return &Port{
		sendPortKey:    sendPortKey,
		receivePortKey: receivePortKey,
		goCallbacks:    xsync.NewIntegerMapOf[uint64, reflect.Value](),
	}
}

func (p *Port) key() PortKey { return p.receivePortKey }

func (p *Port) String() string {
	return fmt.Sprintf("Port[S=%016X, R=%016X]", p.sendPortKey, p.receivePortKey)
}
