package netprobe

import (
	"context"
	"math"
	"net"
	"sort"
	"strconv"
	"time"

	"github.com/sujanv/pingforge/internal/config"
	"github.com/sujanv/pingforge/internal/core"
)

// ProbeLatency measures round-trip latency across consecutive probes,
// calculating sample variance, standard deviation, RFC 3550 jitter, and percentiles.
func ProbeLatency(ctx context.Context, host string, port int, cfg config.PingConfig, timeout time.Duration) *core.LatencyStats {
	if cfg.Count <= 0 {
		cfg.Count = 5
	}
	if cfg.Interval <= 0 {
		cfg.Interval = 200 * time.Millisecond
	}
	if timeout <= 0 {
		timeout = 2 * time.Second
	}

	target := core.Target{
		Host:     host,
		Port:     port,
		Protocol: "tcp",
		Timeout:  timeout,
	}

	stats := &core.LatencyStats{
		Target:    target,
		Count:     cfg.Count,
		Sent:      0,
		Received:  0,
		Samples:   make([]float64, 0, cfg.Count),
		StartTime: time.Now(),
	}

	address := net.JoinHostPort(host, strconv.Itoa(port))
	dialer := &net.Dialer{Timeout: timeout}

	var previousTransit float64
	var currentJitter float64

	for i := 0; i < cfg.Count; i++ {
		select {
		case <-ctx.Done():
			break
		default:
		}

		stats.Sent++
		start := time.Now()
		conn, err := dialer.DialContext(ctx, "tcp", address)
		elapsed := time.Since(start)

		if err == nil {
			_ = conn.Close()
			stats.Received++
			ms := float64(elapsed.Microseconds()) / 1000.0
			stats.Samples = append(stats.Samples, ms)

			// RFC 3550 Interarrival Jitter computation:
			// J = J + (|D(i-1, i)| - J) / 16
			if stats.Received > 1 {
				d := math.Abs(ms - previousTransit)
				currentJitter += (d - currentJitter) / 16.0
			}
			previousTransit = ms
		}

		if i < cfg.Count-1 {
			time.Sleep(cfg.Interval)
		}
	}

	stats.DurationMs = float64(time.Since(stats.StartTime).Microseconds()) / 1000.0

	// Calculate loss rate
	if stats.Sent > 0 {
		stats.LossRate = (float64(stats.Sent-stats.Received) / float64(stats.Sent)) * 100.0
	}

	// Calculate statistics if we have samples
	if len(stats.Samples) > 0 {
		CalculateStats(stats, currentJitter)
	}

	return stats
}

// CalculateStats computes min, max, avg, standard deviation, and percentiles for samples.
func CalculateStats(stats *core.LatencyStats, jitter float64) {
	n := len(stats.Samples)
	if n == 0 {
		return
	}

	var sum float64
	minVal := math.MaxFloat64
	maxVal := -math.MaxFloat64

	for _, s := range stats.Samples {
		sum += s
		if s < minVal {
			minVal = s
		}
		if s > maxVal {
			maxVal = s
		}
	}

	stats.MinMs = minVal
	stats.MaxMs = maxVal
	stats.AvgMs = sum / float64(n)
	stats.JitterMs = jitter

	// Standard Deviation
	if n > 1 {
		var varianceSum float64
		for _, s := range stats.Samples {
			diff := s - stats.AvgMs
			varianceSum += diff * diff
		}
		stats.StdDevMs = math.Sqrt(varianceSum / float64(n-1))
	} else {
		stats.StdDevMs = 0
	}

	// Percentiles (sort copy of samples)
	sorted := make([]float64, n)
	copy(sorted, stats.Samples)
	sort.Float64s(sorted)

	stats.P50Ms = percentile(sorted, 50)
	stats.P90Ms = percentile(sorted, 90)
	stats.P95Ms = percentile(sorted, 95)
	stats.P99Ms = percentile(sorted, 99)
}

func percentile(sorted []float64, p float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if n == 1 {
		return sorted[0]
	}

	rank := (p / 100.0) * float64(n-1)
	lower := int(math.Floor(rank))
	upper := int(math.Ceil(rank))

	if lower == upper {
		return sorted[lower]
	}
	weight := rank - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}
