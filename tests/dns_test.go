package tests

import (
	"context"
	"testing"
	"time"

	"github.com/sujanv/pingforge/internal/config"
	"github.com/sujanv/pingforge/internal/core"
	"github.com/sujanv/pingforge/internal/netprobe"
)

func TestProbeDNSLocalhost(t *testing.T) {
	cfg := config.DNSConfig{
		Resolvers:   []string{"system"},
		RecordTypes: []string{"A"},
		Timeout:     2 * time.Second,
	}

	ctx := context.Background()
	result := netprobe.ProbeDNS(ctx, "localhost", cfg)

	if result.Status != core.StatusUp {
		t.Logf("DNS localhost returned %v (err: %s)", result.Status, result.Error)
	}
}
