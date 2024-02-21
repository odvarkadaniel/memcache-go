package memcache

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func New(addresses []string, connCount int) *Client {
	sl := &ServerList{}
	if err := sl.addServer(addresses...); err != nil {
		return nil
	}

	cl := &Client{
		router:        sl,
		idleConnCount: connCount,
		connPool:      make(map[string][]*Connection),
	}

	cmp, err := cl.router.InitializeConnectionPool(cl.idleConnCount)
	if err != nil {
		return nil
	}

	cl.connPool = cmp

	return cl
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var retErr error

	for _, conns := range c.connPool {
		for _, conn := range conns {
			if err := conn.conn.Close(); err != nil {
				retErr = err
			}
		}
	}

	return retErr
}

func (c *Client) Set(item *Item) error {
	return c.set(item)
}

func (c *Client) set(item *Item) error {
	if ok := isKeyValid(item.Key); !ok {
		return fmt.Errorf("given key is not valid")
	}

	cn, err := c.createReadWriter(item.Key)
	if err != nil {
		return err
	}

	return c.storageFn("set", cn, item)
}

func (c *Client) Add(item *Item) error {
	return c.add(item)
}

func (c *Client) add(item *Item) error {
	if ok := isKeyValid(item.Key); !ok {
		return fmt.Errorf("given key is not valid")
	}

	cn, err := c.createReadWriter(item.Key)
	if err != nil {
		return err
	}

	return c.storageFn("add", cn, item)
}

func (c *Client) Replace(item *Item) error {
	return c.replace(item)
}

func (c *Client) replace(item *Item) error {
	if ok := isKeyValid(item.Key); !ok {
		return fmt.Errorf("given key is not valid")
	}

	cn, err := c.createReadWriter(item.Key)
	if err != nil {
		return err
	}

	return c.storageFn("replace", cn, item)
}

func (c *Client) Append(item *Item) error {
	return c.append(item)
}

func (c *Client) append(item *Item) error {
	if ok := isKeyValid(item.Key); !ok {
		return fmt.Errorf("given key is not valid")
	}

	cn, err := c.createReadWriter(item.Key)
	if err != nil {
		return err
	}

	return c.storageFn("append", cn, item)
}

func (c *Client) Prepend(item *Item) error {
	return c.prepend(item)
}

func (c *Client) prepend(item *Item) error {
	if ok := isKeyValid(item.Key); !ok {
		return fmt.Errorf("given key is not valid")
	}

	cn, err := c.createReadWriter(item.Key)
	if err != nil {
		return err
	}

	return c.storageFn("prepend", cn, item)
}

func (c *Client) CompareAndSwap(item *Item) error {
	return c.compareAndSwap(item)
}

func (c *Client) compareAndSwap(item *Item) error {
	if ok := isKeyValid(item.Key); !ok {
		return fmt.Errorf("given key is not valid")
	}

	cn, err := c.createReadWriter(item.Key)
	if err != nil {
		return err
	}

	return c.storageFn("cas", cn, item)
}

func (c *Client) Get(key string) (*Item, error) {
	return c.get(key)
}

func (c *Client) get(key string) (*Item, error) {
	if ok := isKeyValid(key); !ok {
		return nil, errors.New("given key is not valid")
	}

	cn, err := c.createReadWriter(key)
	if err != nil {
		return nil, err
	}

	return c.retrieveFn("get", cn, key)
}

func (c *Client) Gets(key string) (*Item, error) {
	if ok := isKeyValid(key); !ok {
		return nil, errors.New("given key is not valid")
	}

	cn, err := c.createReadWriter(key)
	if err != nil {
		return nil, err
	}

	return c.retrieveFn("gets", cn, key)
}

func (c *Client) Delete(key string) error {
	return c.delete(key)
}

func (c *Client) delete(key string) error {
	if ok := isKeyValid(key); !ok {
		return errors.New("given key is not valid")
	}

	cn, err := c.createReadWriter(key)
	if err != nil {
		return err
	}

	return c.deleteFn("delete", cn, key)
}

func (c *Client) Incr(key string, delta uint64) (uint64, error) {
	return c.incr(key, delta)
}

func (c *Client) incr(key string, delta uint64) (uint64, error) {
	if ok := isKeyValid(key); !ok {
		return 0, errors.New("given key is not valid")
	}

	cn, err := c.createReadWriter(key)
	if err != nil {
		return 0, err
	}

	return c.incrDecrFn("incr", cn, key, delta)
}

func (c *Client) Decr(key string, delta uint64) (uint64, error) {
	return c.decr(key, delta)
}

func (c *Client) decr(key string, delta uint64) (uint64, error) {
	if ok := isKeyValid(key); !ok {
		return 0, errors.New("given key is not valid")
	}

	cn, err := c.createReadWriter(key)
	if err != nil {
		return 0, err
	}

	return c.incrDecrFn("decr", cn, key, delta)
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

func (c *Client) storageFn(verb string, cn *Connection, item *Item) error {
	var cmd string

	defer c.putBackConnection(cn)

	if verb == "cas" {
		cmd = fmt.Sprintf("%s %s %d %d %d %d\r\n",
			verb, item.Key, item.Flags, int(item.Expiration.Seconds()), len(item.Value), item.CAS)
	} else {
		cmd = fmt.Sprintf("%s %s %d %d %d\r\n",
			verb, item.Key, item.Flags, int(item.Expiration.Seconds()), len(item.Value))
	}

	if _, err := fmt.Fprint(cn.rw, cmd); err != nil {
		return err
	}

	if _, err := cn.rw.Write(item.Value); err != nil {
		return err
	}
	if _, err := cn.rw.Write([]byte("\r\n")); err != nil {
		return err
	}
	if err := cn.rw.Flush(); err != nil {
		return err
	}

	if err := parseStorageResponse(cn.rw); err != nil {
		return err
	}

	return nil
}

func (c *Client) createReadWriter(key string) (*Connection, error) {
	addr, err := c.router.pickServer(key)
	if err != nil {
		return nil, err
	}

	// Look into cache for a connection
	if conn := c.getFreeConn(addr.String()); conn != nil {
		return conn, nil
	}

	return nil, nil
}

func (c *Client) getFreeConn(addr string) *Connection {
	c.mu.Lock()
	defer c.mu.Unlock()

	for {
		for i, cn := range c.connPool[addr] {
			c.connPool[addr] = c.connPool[addr][i+1:]
			cn.owner = addr

			// fmt.Println("Len 1cn:", len(c.connPool[cn.owner]))

			return cn
		}

		time.Sleep(10 * time.Millisecond)
	}
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

func (c *Client) incrDecrFn(verb string, cn *Connection, key string, delta uint64) (uint64, error) {
	defer c.putBackConnection(cn)

	cmd := fmt.Sprintf("%s %s %d\r\n", verb, key, delta)

	line, err := writeFlushRead(cn.rw, cmd)
	if err != nil {
		return 0, err
	}

	return parseIncrDecr(line)
}

func (c *Client) putBackConnection(cn *Connection) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.connPool[cn.owner] = append(c.connPool[cn.owner], cn)
	cn.owner = ""
}

func (c *Client) retrieveFn(verb string, cn *Connection, key string) (*Item, error) {
	var cmd string

	defer c.putBackConnection(cn)

	if verb == "gets" {
		cmd = fmt.Sprintf("%s %s\r\n", verb, key)
	} else {
		cmd = fmt.Sprintf("%s %s\r\n", verb, key)
	}

	line, err := writeFlushRead(cn.rw, cmd)
	if err != nil {
		return nil, err
	}

	it, err := parseGetResponse(line)
	if err != nil && errors.Is(err, ErrCacheMiss) {
		return nil, err
	}

	val, err := cn.rw.ReadSlice('\n')
	if err != nil {
		return nil, err
	}

	// To get rid of the CRLF
	it.Value = val[:len(val)-2]

	// Parse the final END\r\n
	if _, err := cn.rw.ReadSlice('\n'); err != nil {
		return nil, err
	}

	return it, nil
}

func (c *Client) deleteFn(verb string, cn *Connection, key string) error {
	cmd := fmt.Sprintf("%s %s\r\n", verb, key)

	defer c.putBackConnection(cn)

	line, err := writeFlushRead(cn.rw, cmd)
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

func parseGetResponse(resp []byte) (*Item, error) {
	if bytes.Equal(resp, []byte("END\r\n")) {
		return nil, ErrCacheMiss
	}

	splitResp := strings.Split(string(resp), " ")

	flags, _ := strconv.Atoi(splitResp[2])

	if len(splitResp) == 5 {
		cas, _ := strconv.Atoi(splitResp[4])

		return &Item{
			Key:   splitResp[1],
			Flags: int32(flags),
			CAS:   int64(cas),
		}, nil
	}

	return &Item{
		Key:   splitResp[1],
		Flags: int32(flags),
	}, nil
}
