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
	"net"
	"os/exec"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMemcache(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Memcache Suite")
}

var cmd *exec.Cmd
var defaultPort string = "11211"
var defaultIP string = "127.0.0.1"

var _ = BeforeSuite(func() {
	cmd = exec.Command("memcached",
		"--port="+defaultPort,
		"--listen="+defaultIP)

	err := cmd.Start()
	Expect(err).ToNot(HaveOccurred(), "failed to start memcached")

	for i := 0; i < 5; i++ {
		if nc, err := net.Dial("tcp", "127.0.0.1:11211"); err == nil {
			nc.Close()
			break
		}
		time.Sleep(time.Duration(50*i) * time.Millisecond)
	}
})

var _ = AfterSuite(func() {
	cmd.Process.Kill()
})
