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

	return &Client{router: sl}
}

func (c *Client) Set(item *Item) error {
	return c.set(item)
}

func (c *Client) set(item *Item) error {
	if ok := isKeyValid(item.Key); !ok {
		return fmt.Errorf("given key is not valid")
	}

	rw, err := c.createReadWriter()
	if err != nil {
		return err
	}

	return c.storageFn("set", rw, item)
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

	rw, err := c.createReadWriter()
	if err != nil {
		return nil, err
	}

	return c.retrieveFn("get", rw, key)
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
		return bufio.NewReadWriter(bufio.NewReader(*conn), bufio.NewWriter(*conn)), nil
	}

	conn, err := net.Dial(addr.Network(), addr.String())
	if err != nil {
		return nil, err
	}

	x := &ConnPool{
		idle:   []*net.Conn{&conn},
		client: c,
		rw:     bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
	}

	c.connPool[addr.String()] = x

	return bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)), nil
}

func (c *Client) getFreeConn(addr string) (*net.Conn, bool) {
	if c.connPool == nil {
		c.connPool = make(map[string]*ConnPool)
		return nil, false
	}

	conn := c.connPool[addr]
	if conn.idle[0] != nil {
		return conn.idle[0], true
	}

	return nil, false
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
	if it == nil {
		if err != nil {
			return nil, err
		}

		return nil, ErrCacheMiss
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
		return nil, nil
	}

	splitResp := strings.Split(string(resp), " ")

	flags, _ := strconv.ParseInt(splitResp[2], 10, 32)

	it := &Item{
		Key:   splitResp[1],
		Flags: int32(flags),
	}

	return it, nil
}
