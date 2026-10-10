package tests

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/sujanv/pingforge/internal/config"
	"github.com/sujanv/pingforge/internal/core"
	"github.com/sujanv/pingforge/internal/netprobe"
)

func TestProbeTCPSuccess(t *testing.T) {
	// Start local test TCP listener
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	cfg := config.TCPConfig{
		KeepAlive: 5 * time.Second,
		NoDelay:   true,
	}

	ctx := context.Background()
	result := netprobe.ProbeTCP(ctx, addr.IP.String(), addr.Port, cfg, 2*time.Second)

	if result.Status != core.StatusUp {
		t.Fatalf("expected status UP, got %v (err: %s)", result.Status, result.Error)
	}
	if result.LatencyMs <= 0 {
		t.Errorf("expected positive latency, got %.2f ms", result.LatencyMs)
	}
}

func TestProbeTCPTimeout(t *testing.T) {
	cfg := config.TCPConfig{}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Use non-routable address to force timeout
	result := netprobe.ProbeTCP(ctx, "192.0.2.1", 81, cfg, 100*time.Millisecond)

	if result.Status != core.StatusDown {
		t.Fatalf("expected status DOWN on timeout, got %v", result.Status)
	}
}
