// MIT License
//
// Copyright (c) 2024 Odvarka Daniel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package memcache

import (
	"bufio"
	"hash/crc32"
	"net"
	"strings"
	"sync"
)

// ServerList holds the list of all server addresses.
// It is concurrent-safe.
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

// InitializeConnectionPool creates connections for server addresses.
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
