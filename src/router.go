package memcache

import (
	"bufio"
	"hash/crc32"
	"net"
	"strings"
	"sync"
)

type ServerList struct {
	mu    sync.RWMutex
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

func (sl *ServerList) InitializeConnectionPool(connCount int) (map[string][]*Connection, error) {
	mcp := make(map[string][]*Connection)

	sl.mu.RLock()
	defer sl.mu.RUnlock()

	for _, addr := range sl.addrs {
		for i := 0; i < connCount; i++ {

			cp := &Connection{}

			conn, err := net.Dial(addr.Network(), addr.String())
			if err != nil {
				return nil, err
			}
			cp.conn = conn

			cp.rw = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

			mcp[addr.String()] = append(mcp[addr.String()], cp)
		}
	}

	return mcp, nil
}

func (sl *ServerList) pickServer(key string) (net.Addr, error) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	if len(sl.addrs) == 0 {
		return nil, ErrNoServers
	}

	return sl.addrs[int(crc32.ChecksumIEEE([]byte(key)))%len(sl.addrs)], nil
}
