package tests

import (
	"context"
	"testing"
	"time"

	"github.com/sujanv/pingforge/internal/engine"
)

func TestExpandCIDR(t *testing.T) {
	ips, err := engine.ExpandCIDR("192.168.1.0/29")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// /29 has 8 addresses, minus network and broadcast = 6 host addresses
	if len(ips) != 6 {
		t.Fatalf("expected 6 IPs, got %d (%v)", len(ips), ips)
	}
	if ips[0] != "192.168.1.1" {
		t.Errorf("expected first IP 192.168.1.1, got %s", ips[0])
	}
	if ips[len(ips)-1] != "192.168.1.6" {
		t.Errorf("expected last IP 192.168.1.6, got %s", ips[len(ips)-1])
	}
}

func TestWorkerPool(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	pool := engine.NewPool[int, int](3, func(ctx context.Context, item int) int {
		time.Sleep(10 * time.Millisecond)
		return item * 2
	})

	ctx := context.Background()
	results := pool.Run(ctx, items)

	if len(results) != len(items) {
		t.Fatalf("expected %d results, got %d", len(items), len(results))
	}
	for i, v := range results {
		if v != items[i]*2 {
			t.Errorf("index %d: expected %d, got %d", i, items[i]*2, v)
		}
	}
	if pool.ProcessedCount() != int64(len(items)) {
		t.Errorf("expected processed count %d, got %d", len(items), pool.ProcessedCount())
	}
}
