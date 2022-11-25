package dgo

//lint:file-ignore U1000 go:linkname

func lookupPort(key PortKey) *Port { return portMap.GetByPortKey(key) }

func isPortGone(key PortKey, isDefault bool) bool {
	portMap.mu.Lock()
	defer portMap.mu.Unlock()

	_, exists := portMap.key2port[key]
	if !isDefault {
		return !exists
	} else {
		return !exists &&
			(portMap.defaultPort == nil ||
				portMap.defaultPort.key() != key)
	}
}

func testPostToDartPort(port *Port) bool {
	return port.postInt(0, false)
}

func testPostToGoPort(port *Port) bool {
	return (&Port{sendPortKey: port.receivePortKey}).postInt(0, false)
}
