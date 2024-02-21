package memcache

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

func New(addresses []string) *Client {
	sl := &ServerList{}
	if err := sl.addServer(addresses...); err != nil {
		return nil
	}

	cl := &Client{
		router: sl,
		// TODO: Change that to be configurable
		connCount: 1,
		connPool:  make(map[string][]*Connection),
	}

	cmp, err := cl.router.initializeConnectionPool(cl.connCount)
	if err != nil {
		return nil
	}

	cl.connPool = cmp

	return cl
}

func (c *Client) Set(item *Item) error {
	return c.set(item)
}

func (c *Client) set(item *Item) error {
	if ok := isKeyValid(item.Key); !ok {
		return fmt.Errorf("given key is not valid")
	}

	rw, err := c.createReadWriter2()
	if err != nil {
		return err
	}

	return c.storageFn("set", rw.rw, item)
}

func (c *Client) Add(item *Item) error {
	return c.add(item)
}

func (c *Client) add(item *Item) error {
	if ok := isKeyValid(item.Key); !ok {
		return fmt.Errorf("given key is not valid")
	}

	rw, err := c.createReadWriter()
	if err != nil {
		return err
	}

	return c.storageFn("add", rw, item)
}

func (c *Client) Replace(item *Item) error {
	return c.replace(item)
}

func (c *Client) replace(item *Item) error {
	if ok := isKeyValid(item.Key); !ok {
		return fmt.Errorf("given key is not valid")
	}

	rw, err := c.createReadWriter()
	if err != nil {
		return err
	}

	return c.storageFn("replace", rw, item)
}

func (c *Client) Append(item *Item) error {
	return c.append(item)
}

func (c *Client) append(item *Item) error {
	if ok := isKeyValid(item.Key); !ok {
		return fmt.Errorf("given key is not valid")
	}

	rw, err := c.createReadWriter()
	if err != nil {
		return err
	}

	return c.storageFn("append", rw, item)
}

func (c *Client) Prepend(item *Item) error {
	return c.prepend(item)
}

func (c *Client) prepend(item *Item) error {
	if ok := isKeyValid(item.Key); !ok {
		return fmt.Errorf("given key is not valid")
	}

	rw, err := c.createReadWriter()
	if err != nil {
		return err
	}

	return c.storageFn("prepend", rw, item)
}

func (c *Client) CompareAndSwap(item *Item) error {
	return c.compareAndSwap(item)
}

func (c *Client) compareAndSwap(item *Item) error {
	if ok := isKeyValid(item.Key); !ok {
		return fmt.Errorf("given key is not valid")
	}

	rw, err := c.createReadWriter()
	if err != nil {
		return err
	}

	return c.storageFn("cas", rw, item)
}

func (c *Client) Get(key string) (*Item, error) {
	return c.get(key)
}

func (c *Client) get(key string) (*Item, error) {
	if ok := isKeyValid(key); !ok {
		return nil, errors.New("given key is not valid")
	}

	rw, err := c.createReadWriter2()
	if err != nil {
		return nil, err
	}

	return c.retrieveFn("get", rw.rw, key)
}

func (c *Client) Gets() {
	panic("Not yet implemented")
}

func (c *Client) Delete(key string) error {
	return c.delete(key)
}

func (c *Client) delete(key string) error {
	if ok := isKeyValid(key); !ok {
		return errors.New("given key is not valid")
	}

	rw, err := c.createReadWriter()
	if err != nil {
		return err
	}

	return c.deleteFn("delete", rw, key)
}

func (c *Client) Incr(key string, delta uint64) (uint64, error) {
	return c.incr(key, delta)
}

func (c *Client) incr(key string, delta uint64) (uint64, error) {
	if ok := isKeyValid(key); !ok {
		return 0, errors.New("given key is not valid")
	}

	rw, err := c.createReadWriter()
	if err != nil {
		return 0, err
	}

	return c.incrDecrFn("incr", rw, key, delta)
}

func (c *Client) Decr(key string, delta uint64) (uint64, error) {
	return c.decr(key, delta)
}

func (c *Client) decr(key string, delta uint64) (uint64, error) {
	if ok := isKeyValid(key); !ok {
		return 0, errors.New("given key is not valid")
	}

	rw, err := c.createReadWriter()
	if err != nil {
		return 0, err
	}

	return c.incrDecrFn("decr", rw, key, delta)
}

func isKeyValid(key string) bool {
	if len(key) > 250 {
		return false
	}

	for _, c := range key {
		if c == ' ' || c == '\n' {
			return false
		}
	}

	return true
}

func (c *Client) storageFn(verb string, rw *bufio.ReadWriter, item *Item) error {
	var cmd string

	if verb == "cas" {
		cmd = fmt.Sprintf("%s %s %d %d %d %d\r\n",
			verb, item.Key, item.Flags, int(item.Expiration.Seconds()), len(item.Value), item.CAS)
	} else {
		cmd = fmt.Sprintf("%s %s %d %d %d\r\n",
			verb, item.Key, item.Flags, int(item.Expiration.Seconds()), len(item.Value))
	}

	if _, err := fmt.Fprint(rw, cmd); err != nil {
		return err
	}

	if _, err := rw.Write(item.Value); err != nil {
		return err
	}
	if _, err := rw.Write([]byte("\r\n")); err != nil {
		return err
	}
	if err := rw.Flush(); err != nil {
		return err
	}

	if err := parseStorageResponse(rw); err != nil {
		return err
	}

	return nil
}

func (c *Client) createReadWriter() (*bufio.ReadWriter, error) {
	addr, err := c.router.pickServer()
	if err != nil {
		return nil, err
	}

	// Look into cache for a connection
	if conn, ok := c.getFreeConn(addr.String()); ok {
		return bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)), nil
	}

	conn, err := net.Dial(addr.Network(), addr.String())
	if err != nil {
		return nil, err
	}

	// TODO: Proper new connpool
	cn := &Connection{
		conn: conn,
		rw:   bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
	}

	c.connPool[addr.String()] = append(c.connPool[addr.String()], cn)

	return bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)), nil
}

// servers = {
// 	1,
// 	2,
// 	3
// }

// clinet.cp[1] = connpool.idle[c1,c2,c3]
// clinet.cp[2] = connpool.idle[c1,c2,c3]
// clinet.cp[3] = connpool.idle[c1,c2,c3]

// Get -> grab first connection and add 1 to used
// Put -> append the connection to idle and substract 1 from used

// connpool = {
// 	mu       sync.Mutex
// 	capacity uint
// 	used     uint
// 	idle     []net.Conn
// 	rw       *bufio.ReadWriter
// 	client   *Client
// }

func (c *Client) createReadWriter2() (*Connection, error) {
	addr, err := c.router.pickServer()
	if err != nil {
		return nil, err
	}

	// Look into cache for a connection
	if conn := c.getFreeConn2(addr.String()); conn != nil {
		// We found the connection in connectionPool
		return conn, nil
	}

	// conn, err := net.Dial(addr.Network(), addr.String())
	// if err != nil {
	// 	return nil, err
	// }

	// cn := &Connections{
	// 	conn: conn,
	// 	rw:   bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
	// }

	// c.connPool[addr.String()] = append(c.connPool[addr.String()], cn)

	// return bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)), nil

	return nil, nil
}

func (c *Client) getFreeConn2(addr string) *Connection {
	for _, cn := range c.connPool[addr] {
		return cn
	}

	return nil
}

func (c *Client) getFreeConn(addr string) (net.Conn, bool) {
	if c.connPool[addr] == nil {
		fmt.Println("here")
		c.connPool = make(map[string][]*Connection)
		return nil, false
	}

	// TODO: Actual use of idle slice
	// conn := c.connPool[addr].Get()
	// if conn == nil {
	// 	return nil, false
	// }

	// conn := c.connPool[addr]
	// if conn.idle[0] != nil {
	// 	return conn.idle[0], true
	// }

	return nil, true
}

func parseStorageResponse(rw *bufio.ReadWriter) error {
	line, err := rw.ReadSlice('\n')
	if err != nil {
		return err
	}

	switch {
	case bytes.Equal(line, []byte("STORED\r\n")):
		return nil
	case bytes.Equal(line, []byte("ERROR\r\n")):
		return ErrError
	case bytes.Equal(line, []byte("NOT_STORED\r\n")):
		return ErrNotStored
	case bytes.Equal(line, []byte("CLIENT_ERROR\r\n")):
		return ErrClientError
	case bytes.Equal(line, []byte("EXISTS\r\n")):
		return ErrExists
	case bytes.Equal(line, []byte("NOT_FOUND\r\n")):
		return ErrCacheMiss
	default:
		// This should not happen.
		panic(string(line) + " is not a valid response")
	}
}

func writeFlushRead(rw *bufio.ReadWriter, cmd string) ([]byte, error) {
	if _, err := fmt.Fprint(rw, cmd); err != nil {
		return nil, err
	}

	if err := rw.Flush(); err != nil {
		return nil, err
	}

	line, err := rw.ReadSlice('\n')
	if err != nil {
		return nil, err
	}

	return line, nil
}

func (c *Client) incrDecrFn(verb string, rw *bufio.ReadWriter, key string, delta uint64) (uint64, error) {
	cmd := fmt.Sprintf("%s %s %d\r\n", verb, key, delta)

	line, err := writeFlushRead(rw, cmd)
	if err != nil {
		return 0, err
	}

	return parseIncrDecr(line)
}

func (c *Client) retrieveFn(verb string, rw *bufio.ReadWriter, key string) (*Item, error) {
	var cmd string

	if verb == "cas" {
		return nil, fmt.Errorf("TODO: CAS not yet implemented for retrieval commands")
	} else {
		cmd = fmt.Sprintf("%s %s\r\n", verb, key)
	}

	line, err := writeFlushRead(rw, cmd)
	if err != nil {
		return nil, err
	}

	it, err := parseGet(line)
	if err != nil && errors.Is(err, ErrCacheMiss) {
		return nil, err
	}

	val, err := rw.ReadSlice('\n')
	if err != nil {
		return nil, err
	}

	// To get rid of the CRLF
	it.Value = val[:len(val)-2]

	return it, nil
}

func (c *Client) deleteFn(verb string, rw *bufio.ReadWriter, key string) error {
	cmd := fmt.Sprintf("%s %s\r\n", verb, key)

	line, err := writeFlushRead(rw, cmd)
	if err != nil {
		return err
	}

	if err := parseDelete(line); err != nil {
		return err
	}

	return nil
}

func parseDelete(resp []byte) error {
	switch {
	case bytes.Equal(resp, []byte("DELETED\r\n")):
		return nil
	case bytes.Equal(resp, []byte("END\r\n")):
		return ErrError
	case bytes.Equal(resp, []byte("NOT_FOUND\r\n")):
		return ErrCacheMiss
	default:
		// This should not happen.
		panic(string(resp) + " is not a valid response")
	}
}

func parseIncrDecr(resp []byte) (uint64, error) {
	switch {
	case bytes.Equal(resp, []byte("NOT_FOUND\r\n")):
		return 0, ErrCacheMiss
	case bytes.HasPrefix(resp, []byte("CLIENT_ERROR ")):
		return 0, fmt.Errorf(string(resp[13:]))
	}

	return strconv.ParseUint(string(resp[:len(resp)-2]), 10, 64)
}

func parseGet(resp []byte) (*Item, error) {
	if bytes.Equal(resp, []byte("END\r\n")) {
		return nil, ErrCacheMiss
	}

	splitResp := strings.Split(string(resp), " ")

	flags, _ := strconv.Atoi(splitResp[2])

	return &Item{
		Key:   splitResp[1],
		Flags: int32(flags),
	}, nil
}
