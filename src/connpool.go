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

type Connection struct {
	mu     sync.Mutex
	owner  string
	conn   net.Conn
	rw     *bufio.ReadWriter
	client *Client
}

func (cp *Connection) Get() net.Conn {
	// Maybe use ticker to timeout?
	// fmt.Println(cp)
	// if cp == nil {
	// 	return nil
	// }

	// for {
	// 	if len(cp.idle) > 0 {
	// 		conn := cp.idle[0]

	// 		cp.idle = cp.idle[1:]

	// 		return conn
	// 	}
	// 	time.Sleep(time.Second)
	// 	fmt.Println("Sleeping:", len(cp.idle))
	// }
	return nil
}

func (cp *Connection) Put(conn net.Conn) error {
	return nil
}

func (cp *Connection) Close(conn net.Conn) error {
	return nil
}

// func NewConnPool() *Connection {
// 	return &Connection{
// 		capacity: 5,
// 		used:     0,
// 		idle:     make([]net.Conn, 0, 5),
// 		rw:       nil,
// 		client:   nil,
// 	}
// }
