package tests

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/sujanv/pingforge/internal/netprobe"
)

func TestProbeMTUMock(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			buf := make([]byte, 10000)
			_, _ = c.Read(buf)
			c.Close()
		}
	}()

	addr := l.Addr().(*net.TCPAddr)
	ctx := context.Background()
	res := netprobe.ProbeMTU(ctx, "127.0.0.1", addr.Port, 2*time.Second)

	if res.OptimalMTU < 576 {
		t.Errorf("expected optimal MTU >= 576, got %d", res.OptimalMTU)
	}
}
