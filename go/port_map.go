package dgo

import (
	"fmt"

	xsync "github.com/puzpuzpuz/xsync/v2"
)

type _PortMap struct {
	key2port    map[PortKey]*Port
	defaultPort *Port
	mu          *xsync.RBMutex
}

var portMap = newPortMap()

func newPortMap() *_PortMap {
	return &_PortMap{
		key2port: make(map[PortKey]*Port),
		mu:       xsync.NewRBMutex(),
	}
}

func (m *_PortMap) addPort(sendPortKey PortKey, isDefault bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if isDefault && m.defaultPort != nil {
		panic("dgo:go: default port is already set")
	}

	receivePortKey := m.initPort(sendPortKey)

	port := newPort(
		PortKey(sendPortKey),
		PortKey(receivePortKey),
	)
	m.key2port[port.key()] = port

	if isDefault {
		m.defaultPort = port
	}
}

func (m *_PortMap) removePort(port *Port) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := port.key()
	delete(m.key2port, key)
	if m.defaultPort == port {
		m.defaultPort = nil
	}
}

func (m *_PortMap) DefaultPort() *Port {
	rtok := m.mu.RLock()
	defer m.mu.RUnlock(rtok)
	return m.defaultPort
}

func (m *_PortMap) GetByPortKey(key PortKey) *Port {
	rtok := m.mu.RLock()
	defer m.mu.RUnlock(rtok)
	if port, ok := m.key2port[key]; ok {
		return port
	} else {
		panic(fmt.Sprintf("dgo:go: port [key=%d] not exists", key))
	}
}

func (m *_PortMap) ResolvePort(port *Port) *Port {
	if port != nil {
		return port
	}
	return m.DefaultPort()
}
