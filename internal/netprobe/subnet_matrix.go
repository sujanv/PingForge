package netprobe

import (
	"context"
	"sync"
	"time"

	"github.com/sujanv/pingforge/internal/config"
	"github.com/sujanv/pingforge/internal/core"
)

// SubnetNodeLatency records end-to-end peer ping matrix between hosts.
type SubnetNodeLatency struct {
	SourceHost  string  `json:"source"`
	TargetHost  string  `json:"target"`
	LatencyMs   float64 `json:"latency_ms"`
	IsReachable bool    `json:"reachable"`
}

// SubnetLatencyMesh aggregates all pairwise latency measurements.
type SubnetLatencyMesh struct {
	Nodes    []string            `json:"nodes"`
	Matrix   []SubnetNodeLatency `json:"matrix"`
	MinRTTMs float64             `json:"min_rtt_ms"`
	MaxRTTMs float64             `json:"max_rtt_ms"`
	AvgRTTMs float64             `json:"avg_rtt_ms"`
}

// BuildSubnetMesh probes a list of hosts concurrently to build an N-by-N latency matrix.
func BuildSubnetMesh(ctx context.Context, hosts []string, port int, timeout time.Duration) *SubnetLatencyMesh {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	if port <= 0 {
		port = 80
	}

	mesh := &SubnetLatencyMesh{
		Nodes:  hosts,
		Matrix: make([]SubnetNodeLatency, 0),
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	tcpCfg := config.TCPConfig{
		NoDelay: true,
	}

	for _, src := range hosts {
		for _, dst := range hosts {
			wg.Add(1)
			go func(s, d string) {
				defer wg.Done()
				start := time.Now()
				res := ProbeTCP(ctx, d, port, tcpCfg, timeout)
				rtt := float64(time.Since(start).Microseconds()) / 1000.0

				mu.Lock()
				mesh.Matrix = append(mesh.Matrix, SubnetNodeLatency{
					SourceHost:  s,
					TargetHost:  d,
					LatencyMs:   rtt,
					IsReachable: res.Status == core.StatusUp,
				})
				mu.Unlock()
			}(src, dst)
		}
	}

	wg.Wait()

	var sum float64
	var count int
	for _, m := range mesh.Matrix {
		if m.IsReachable && m.LatencyMs > 0 {
			if count == 0 || m.LatencyMs < mesh.MinRTTMs {
				mesh.MinRTTMs = m.LatencyMs
			}
			if m.LatencyMs > mesh.MaxRTTMs {
				mesh.MaxRTTMs = m.LatencyMs
			}
			sum += m.LatencyMs
			count++
		}
	}
	if count > 0 {
		mesh.AvgRTTMs = sum / float64(count)
	}

	return mesh
}
