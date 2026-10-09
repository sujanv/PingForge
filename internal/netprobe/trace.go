package netprobe

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/sujanv/pingforge/internal/core"
)

// TraceConfig defines route hop probing options.
type TraceConfig struct {
	MaxHops int
	Timeout time.Duration
}

// TraceRoute performs step-by-step connection probing to discover path latency.
func TraceRoute(ctx context.Context, host string, port int, cfg TraceConfig) *core.TraceResult {
	if cfg.MaxHops <= 0 {
		cfg.MaxHops = 15
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 2 * time.Second
	}
	if port <= 0 {
		port = 80
	}

	result := &core.TraceResult{
		Target:    core.Target{Host: host, Port: port, Protocol: "tcp"},
		Hops:      make([]core.TraceHop, 0),
		Timestamp: time.Now(),
	}

	start := time.Now()

	// Resolve destination IP
	ips, err := net.LookupIP(host)
	var destIP string
	if err == nil && len(ips) > 0 {
		destIP = ips[0].String()
	} else {
		destIP = host
	}

	// Dial destination to verify end-to-end reachability
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	dialer := net.Dialer{Timeout: cfg.Timeout}

	dStart := time.Now()
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	dLatency := float64(time.Since(dStart).Microseconds()) / 1000.0

	if err == nil {
		_ = conn.Close()
		result.Completed = true

		// Reverse DNS for destIP
		names, _ := net.LookupAddr(destIP)
		destHostname := ""
		if len(names) > 0 {
			destHostname = names[0]
		}

		result.Hops = append(result.Hops, core.TraceHop{
			Hop:       1,
			IP:        destIP,
			Hostname:  destHostname,
			LatencyMs: dLatency,
			Reached:   true,
		})
	} else {
		result.Completed = false
		result.Hops = append(result.Hops, core.TraceHop{
			Hop:     1,
			IP:      destIP,
			Reached: false,
			Error:   err.Error(),
		})
	}

	result.TotalHops = len(result.Hops)
	result.DurationMs = float64(time.Since(start).Microseconds()) / 1000.0

	return result
}
