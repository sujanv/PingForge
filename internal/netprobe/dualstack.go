package netprobe

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// DualStackResult records IPv4 vs IPv6 Happy Eyeballs (RFC 8305) race metrics.
type DualStackResult struct {
	Domain             string   `json:"domain"`
	IPv4Addresses      []string `json:"ipv4_addresses"`
	IPv6Addresses      []string `json:"ipv6_addresses"`
	IPv4LatencyMs      float64  `json:"ipv4_latency_ms"`
	IPv6LatencyMs      float64  `json:"ipv6_latency_ms"`
	WinningFamily      string   `json:"winning_family"` // IPv4, IPv6, None
	HappyEyeballsWinMs float64  `json:"happy_eyeballs_win_ms"`
	Error              string   `json:"error,omitempty"`
}

// ProbeDualStack evaluates dual-stack resolution and connection race performance.
func ProbeDualStack(ctx context.Context, domain string, port int, timeout time.Duration) *DualStackResult {
	if port <= 0 {
		port = 80
	}
	if timeout <= 0 {
		timeout = 4 * time.Second
	}

	res := &DualStackResult{
		Domain:        domain,
		WinningFamily: "None",
	}

	ips, err := net.LookupIP(domain)
	if err != nil {
		res.Error = fmt.Sprintf("dns lookup failed: %v", err)
		return res
	}

	for _, ip := range ips {
		if ip.To4() != nil {
			res.IPv4Addresses = append(res.IPv4Addresses, ip.String())
		} else {
			res.IPv6Addresses = append(res.IPv6Addresses, ip.String())
		}
	}

	var wg sync.WaitGroup
	var v4Err, v6Err error

	// Probe IPv4
	if len(res.IPv4Addresses) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			addr := net.JoinHostPort(res.IPv4Addresses[0], fmt.Sprintf("%d", port))
			start := time.Now()
			conn, err := net.DialTimeout("tcp4", addr, timeout)
			if err == nil {
				conn.Close()
				res.IPv4LatencyMs = float64(time.Since(start).Microseconds()) / 1000.0
			} else {
				v4Err = err
			}
		}()
	}

	// Probe IPv6
	if len(res.IPv6Addresses) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			addr := net.JoinHostPort(res.IPv6Addresses[0], fmt.Sprintf("%d", port))
			start := time.Now()
			conn, err := net.DialTimeout("tcp6", addr, timeout)
			if err == nil {
				conn.Close()
				res.IPv6LatencyMs = float64(time.Since(start).Microseconds()) / 1000.0
			} else {
				v6Err = err
			}
		}()
	}

	wg.Wait()

	// Single-stack protocol winner fallback
	if res.IPv4LatencyMs == 0 && res.IPv6LatencyMs > 0 {
		res.WinningFamily = "IPv6 (Single-Stack)"
	} else if res.IPv6LatencyMs == 0 && res.IPv4LatencyMs > 0 {
		res.WinningFamily = "IPv4 (Single-Stack)"
	}

	if res.IPv6LatencyMs > 0 && (res.IPv4LatencyMs == 0 || res.IPv6LatencyMs <= res.IPv4LatencyMs+50.0) {
		res.WinningFamily = "IPv6"
		res.HappyEyeballsWinMs = res.IPv4LatencyMs - res.IPv6LatencyMs
	} else if res.IPv4LatencyMs > 0 {
		res.WinningFamily = "IPv4"
		res.HappyEyeballsWinMs = res.IPv6LatencyMs - res.IPv4LatencyMs
	} else {
		res.Error = fmt.Sprintf("dual-stack dial failed (v4: %v, v6: %v)", v4Err, v6Err)
	}

	return res
}
