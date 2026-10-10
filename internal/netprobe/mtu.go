package netprobe

import (
	"context"
	"net"
	"strconv"
	"time"
)

// MTUThreshold records result for a tested packet size.
type MTUThreshold struct {
	Size       int     `json:"size"`
	Success    bool    `json:"success"`
	LatencyMs  float64 `json:"latency_ms"`
}

// MTUResult stores Path MTU detection metrics.
type MTUResult struct {
	Host          string         `json:"host"`
	Port          int            `json:"port"`
	OptimalMTU    int            `json:"optimal_mtu"`
	TestedSizes   []MTUThreshold `json:"tested_sizes"`
	EstimatedMSS  int            `json:"estimated_mss"`
	TotalDuration float64        `json:"total_duration_ms"`
}

// ProbeMTU detects socket buffer threshold and path payload size limits.
func ProbeMTU(ctx context.Context, host string, port int, timeout time.Duration) *MTUResult {
	if port <= 0 {
		port = 80
	}
	if timeout <= 0 {
		timeout = 2 * time.Second
	}

	res := &MTUResult{
		Host:        host,
		Port:        port,
		TestedSizes: make([]MTUThreshold, 0),
	}

	sizes := []int{576, 1280, 1420, 1500, 2048, 9000}
	addr := net.JoinHostPort(host, strconv.Itoa(port))

	start := time.Now()
	maxSuccess := 576

	for _, sz := range sizes {
		d := net.Dialer{Timeout: timeout}
		pStart := time.Now()
		conn, err := d.DialContext(ctx, "tcp", addr)
		if err != nil {
			res.TestedSizes = append(res.TestedSizes, MTUThreshold{Size: sz, Success: false})
			continue
		}

		payload := make([]byte, sz)
		_ = conn.SetDeadline(time.Now().Add(timeout))
		_, wErr := conn.Write(payload)
		_ = conn.Close()

		lat := float64(time.Since(pStart).Microseconds()) / 1000.0
		if wErr == nil {
			res.TestedSizes = append(res.TestedSizes, MTUThreshold{Size: sz, Success: true, LatencyMs: lat})
			if sz > maxSuccess {
				maxSuccess = sz
			}
		} else {
			res.TestedSizes = append(res.TestedSizes, MTUThreshold{Size: sz, Success: false})
		}
	}

	res.OptimalMTU = maxSuccess
	if maxSuccess >= 40 {
		res.EstimatedMSS = maxSuccess - 40 // subtract standard IP+TCP header size
	}
	res.TotalDuration = float64(time.Since(start).Microseconds()) / 1000.0
	return res
}
