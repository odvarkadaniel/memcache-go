package main

import (
	"fmt"
	"time"

	memcache "github.com/odvarkadaniel/memcache-go/src"
)

func main() {
	// cmd := exec.Command("memcached",
	// 	"--port=11211",
	// 	"--listen=127.0.0.1")

	// if err := cmd.Start(); err != nil {
	// 	panic("failed to start memcahced")
	// }

	// defer cmd.Wait()
	// defer cmd.Process.Kill()

	// time.Sleep(time.Second * 1)
	s := time.Now()
	servers := []string{
		"127.0.0.1:11211",
	}
	cl := memcache.New(servers)
	item := &memcache.Item{Key: "test", Value: []byte("test"), Flags: 0, Expiration: time.Second * 60}
	fmt.Println("Init:", time.Since(s))

	st := time.Now()
	for i := 0; i < 200000; i++ {
		cl.Set(item)
	}
	fmt.Println("Final time: ", time.Since(st))

	// err := cl.Set(&memcache.Item{
	// 	Key:        "hello",
	// 	Value:      []byte("world"),
	// 	Expiration: time.Second * 10,
	// 	Flags:      0,
	// })
	// if err != nil {
	// 	log.Println(err)
	// }

	// it, err := cl.Get("hello")
	// if err != nil {
	// 	log.Println(err)
	// } else {
	// 	log.Println(it.Key, "=>", string(it.Value))
	// }

	// err = cl.Append(&memcache.Item{
	// 	Key:        "hello",
	// 	Value:      []byte(" from memcached!"),
	// 	Expiration: time.Second * 60,
	// 	Flags:      0,
	// })
	// if err != nil {
	// 	log.Println(err)
	// }

	// it1, err := cl.Get("hello")
	// if err != nil {
	// 	log.Println(err)
	// } else {
	// 	log.Println(it1.Key, "=>", string(it1.Value))
	// }
}
