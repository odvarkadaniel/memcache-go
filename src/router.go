package memcache

import (
	"net"
	"sync"
)

type ServerList struct {
	// TODO: What about RWMutex?
	mu    sync.Mutex
	addrs []net.Addr
}

// [127.0.0.1:1234, 127.0.0.1:1235, ...] addresses come in
func (sl *ServerList) addServer(connType ConnType, addresses ...string) error {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	// Establish connection with the addresses
	// TODO: Support UDP and UNIX sockets
	switch connType {
	case TCP:
		for _, server := range addresses {
			udpAddr, err := net.ResolveTCPAddr("tcp", server)
			if err != nil {
				return ErrEstablishConnection
			}
			sl.addrs = append(sl.addrs, udpAddr)
		}
	case UDP:
		panic("udp not supported yet")
	case UNIX:
		panic("unix socks not supported yet")
	default:
		panic("unknown connection type")
	}

	return nil
}

// TODO: Proper server selection.
func (sl *ServerList) pickServer() (net.Addr, error) {
	if len(sl.addrs) == 0 {
		return nil, ErrNoServers
	}
	return sl.addrs[0], nil
}
