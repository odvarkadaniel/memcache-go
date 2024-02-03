package memcache

import (
	"bufio"
	"net"
	"sync"
)

type ConnectionPool interface {
	// This gets an active connection from the connection pool if exists,
	// else create a new connection.
	Get() (net.Conn, error)

	// This releases an active connection back to the connection pool.
	Put(conn net.Conn) error

	// This discards an active connection from the connection pool and
	// closes the connection.
	Close(conn net.Conn) error
}

type ConnPool struct {
	mu       sync.Mutex
	capacity uint
	used     uint
	idle     []*net.Conn
	rw       *bufio.ReadWriter
	client   *Client
}

func (cp *ConnPool) Get() (net.Conn, error) {
	return nil, nil
}

func (cp *ConnPool) Put(conn net.Conn) error {
	return nil
}

func (cp *ConnPool) Close(conn net.Conn) error {
	return nil
}

func NewConnPool() *ConnPool {
	return nil
}
