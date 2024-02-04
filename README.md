# memcache-go
Is a memcached client written in Golang.

## Installation
```
go get github.com/odvarkadaniel/memcache-go
```

## Examples of usage
```go
package main

import (
    "github.com/odvarkadaniel/memcache-go"
)

func main() {
  client := memcache.New([]string{"127.0.0.1:11211"})

  item := &memcache.Item{
    Key: "Hello",
    Value: []byte("World"),
    Flags: 0,
    Expiration: time.Second * 60
  }

  if err := client.Set(item); err != nil {
    // Handle the error.
  }

  it, err := client.Get("Hello")
  if err != nil {
    // Handle the error.
  }

  // Do something with the item.
}
```
