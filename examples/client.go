package main

import (
	"fmt"
	"time"

	memcache "github.com/odvarkadaniel/memcache-go/src"
)

func main() {
	servers := []string{
		"127.0.0.1:11211",
		// "127.0.0.1:1235",
		// "127.0.0.1:1236",
	}
	connType := memcache.TCP
	cl := memcache.New(connType, servers)

	cl.Set(&memcache.Item{
		Key:        "hello",
		Data:       []byte("world"),
		Expiration: time.Second * 60,
		Flags:      0,
	})

	it, _ := cl.Get("hello")
	if it == nil {
		fmt.Println("could not find key hello11")

		return
	}

	fmt.Println(it.Key, "=>", string(it.Data))
}
