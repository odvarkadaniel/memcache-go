package memcache

import "testing"

func TestStorageCommands(t *testing.T) {
	testCases := []struct {
		name        string
		verb        string
		item        *Item
		expectedErr bool
	}{}

	mc := New(TCP, []string{"127.0.0.1:11211"})

	for _, tc := range testCases {
	}
}
