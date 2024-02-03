package memcache

import (
	"errors"
	"net"
	"time"
)

type ConnType int8

const (
	TCP ConnType = iota
	UDP
	UNIX
)

var (
	ErrEstablishConnection = errors.New("failed to establish connection")
	ErrNoServers           = errors.New("no servers are currently connected")
	ErrNotStored           = errors.New("failed to store Value")
	ErrError               = errors.New("incorrect syntax or error while saving the Value")
	ErrClientError         = errors.New("failed to store Value while appending/prepending")
	ErrExists              = errors.New("someone else has modified the CAS Value since last fetch")
	ErrCacheMiss           = errors.New("key does not exist in the server")
)

type Item struct {
	Key        string
	Value      []byte
	Expiration time.Duration
	Flags      int32
	CAS        int64
}

type Client struct {
	router   *ServerList
	connPool map[string]*net.Conn

	// ...
}
