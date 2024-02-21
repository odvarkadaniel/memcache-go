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
	"errors"
	"net"
	"sync"
	"time"
)

type ConnType int8

const (
	TCP ConnType = iota
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

// Item represent a memcache item object
type Item struct {
	Key        string
	Value      []byte
	Expiration time.Duration
	Flags      int32
	CAS        int64
}

// Client is the object that is exposed to the user.
// It allows the user to interact with the API.
type Client struct {
	mu            sync.Mutex
	router        *ServerList
	idleConnCount int
	connPool      map[string][]*Connection
}

// Connection represents a single connection to a server.
// We want to hold the connection itself and also a ReadWriter
// due to optimizations.
type Connection struct {
	owner string
	conn  net.Conn
	rw    *bufio.ReadWriter
}
