package tests

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/sujanv/pingforge/internal/netprobe"
)

func TestProbeUDPMock(t *testing.T) {
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	go func() {
		buf := make([]byte, 1024)
		n, addr, err := conn.ReadFrom(buf)
		if err == nil {
			_, _ = conn.WriteTo(append([]byte("ECHO:"), buf[:n]...), addr)
		}
	}()

	addr := conn.LocalAddr().(*net.UDPAddr)
	ctx := context.Background()
	res := netprobe.ProbeUDP(ctx, "127.0.0.1", addr.Port, []byte("HELLO"), 2*time.Second)

	if !res.Reachable {
		t.Errorf("expected reachable UDP")
	}
}
