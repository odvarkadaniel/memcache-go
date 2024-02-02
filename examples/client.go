package main

import (
	"log"
	"os/exec"
	"time"

	memcache "github.com/odvarkadaniel/memcache-go/src"
)

func main() {
	cmd := exec.Command("memcached",
		"--port=11211",
		"--listen=127.0.0.1")

	if err := cmd.Start(); err != nil {
		panic("failed to start memcahced")
	}

	defer cmd.Wait()
	defer cmd.Process.Kill()

	time.Sleep(time.Second * 1)

	servers := []string{
		"127.0.0.1:11211",
	}
	connType := memcache.TCP
	cl := memcache.New(connType, servers)

	err := cl.Set(&memcache.Item{
		Key:        "hello",
		Value:      []byte("world"),
		Expiration: time.Second * 10,
		Flags:      0,
	})
	if err != nil {
		log.Println(err)
	}

	it, err := cl.Get("hello")
	if err != nil {
		log.Println(err)
	} else {
		log.Println(it.Key, "=>", string(it.Value))
	}

	err = cl.Append(&memcache.Item{
		Key:        "hello",
		Value:      []byte(" from memcached!"),
		Expiration: time.Second * 60,
		Flags:      0,
	})
	if err != nil {
		log.Println(err)
	}

	it1, err := cl.Get("hello")
	if err != nil {
		log.Println(err)
	} else {
		log.Println(it1.Key, "=>", string(it1.Value))
	}
}
