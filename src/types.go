package memcache

import (
	"errors"
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
	ErrNotStored           = errors.New("failed to store data")
	ErrError               = errors.New("incorrect syntax or error while saving the data")
	ErrClientError         = errors.New("failed to store data while appending/prepending")
	ErrExists              = errors.New("someone else has modified the CAS data since last fetch")
	ErrNotFound            = errors.New("key does not exist in the server")
)

type Item struct {
	Key        string
	Data       []byte
	Expiration time.Duration
	Flags      int32
	CAS        int64
}

type Client struct {
	router *ServerList

	// ...
}
