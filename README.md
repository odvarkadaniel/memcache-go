# memcache-go
This is a memcached client written in Golang.

## Install
```
go get github.com/odvarkadaniel/memcache-go
```
You can find some usage examples in the `examples` folder.

## Contributing
Before creating a pull request, create an issue first.

## Documentation
To see a list of all the functions, please visit [pkg.go.dev](https://pkg.go.dev/github.com/odvarkadaniel/memcache-go).

## Examples of usage
As mentioned before, more examples can be seen in the `examples` folder.
It might also be useful to look at the tests in `memcache_test.go` file to see
how you can interact with the API.
```go
package main

import (
    "github.com/odvarkadaniel/memcache-go"
)

func main() {
  client := memcache.New([]string{"127.0.0.1:11211"}, 1)

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

## License
This project uses `MIT LICENSE`, for more details, please see the `LICENSE` file.
