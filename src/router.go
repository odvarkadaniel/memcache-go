package memcache

import (
	"net"
	"strings"
	"sync"
)

type ServerList struct {
	// TODO: What about RWMutex?
	mu    sync.Mutex
	addrs []net.Addr
}

func (sl *ServerList) addServer(addresses ...string) error {
	addrs := make([]net.Addr, len(addresses))

	// Establish connection with the addresses
	for i, server := range addresses {
		if strings.Contains(server, "/") {
			addr, err := net.ResolveUnixAddr("unix", server)
			if err != nil {
				return ErrEstablishConnection
			}
			addrs[i] = addr
		} else {
			addr, err := net.ResolveTCPAddr("tcp", server)
			if err != nil {
				return ErrEstablishConnection
			}
			addrs[i] = addr
		}
	}

	sl.mu.Lock()
	sl.addrs = addrs
	sl.mu.Unlock()

	return nil
}

// TODO: Proper server selection.
func (sl *ServerList) pickServer() (net.Addr, error) {
	if len(sl.addrs) == 0 {
		return nil, ErrNoServers
	}
	return sl.addrs[0], nil
}
