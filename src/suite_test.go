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
