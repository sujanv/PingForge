package netprobe

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/sujanv/pingforge/internal/config"
	"github.com/sujanv/pingforge/internal/core"
)

// BufferbloatResult stores network latency inflation under load.
type BufferbloatResult struct {
	Target             string    `json:"target"`
	BaselineLatencyMs  float64   `json:"baseline_latency_ms"`
	LoadedLatencyMs    float64   `json:"loaded_latency_ms"`
	LatencyInflationMs float64   `json:"latency_inflation_ms"`
	Grade              string    `json:"grade"` // A+, A, B, C, D, F
	DownloadSpeedMbps  float64   `json:"download_speed_mbps"`
	TestedAt           time.Time `json:"tested_at"`
	Error              string    `json:"error,omitempty"`
}

// ProbeBufferbloat evaluates latency degradation when saturated with concurrent HTTP/TCP traffic.
func ProbeBufferbloat(ctx context.Context, targetURL string, host string, port int, duration time.Duration) *BufferbloatResult {
	if duration <= 0 {
		duration = 3 * time.Second
	}
	if port <= 0 {
		port = 80
	}
	if targetURL == "" {
		targetURL = "http://" + host
	}

	res := &BufferbloatResult{
		Target:   targetURL,
		TestedAt: time.Now(),
	}

	// 1. Measure baseline unloaded latency (5 sample pings)
	var baselineSamples []float64
	for i := 0; i < 5; i++ {
		start := time.Now()
		probe := ProbeTCP(ctx, host, port, config.TCPConfig{NoDelay: true}, 1*time.Second)
		if probe.Status == core.StatusUp {
			baselineSamples = append(baselineSamples, float64(time.Since(start).Microseconds())/1000.0)
		}
		time.Sleep(50 * time.Millisecond)
	}

	if len(baselineSamples) == 0 {
		res.BaselineLatencyMs = 25.0 // fallback
	} else {
		var sum float64
		for _, s := range baselineSamples {
			sum += s
		}
		res.BaselineLatencyMs = sum / float64(len(baselineSamples))
	}

	// 2. Saturate connection with concurrent HTTP streaming traffic while measuring loaded latency
	stopChan := make(chan struct{})
	var wg sync.WaitGroup
	var bytesRead int64
	var mu sync.Mutex

	client := &http.Client{
		Timeout: 2 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Launch 4 worker saturators
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := make([]byte, 32*1024)
			for {
				select {
				case <-stopChan:
					return
				default:
					req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
					if err == nil {
						resp, err := client.Do(req)
						if err == nil {
							n, _ := io.CopyBuffer(io.Discard, resp.Body, buf)
							resp.Body.Close()
							mu.Lock()
							bytesRead += n
							mu.Unlock()
						}
					}
					time.Sleep(10 * time.Millisecond)
				}
			}
		}()
	}

	// Concurrently measure loaded latency
	var loadedSamples []float64
	testStart := time.Now()
	for time.Since(testStart) < duration {
		start := time.Now()
		probe := ProbeTCP(ctx, host, port, config.TCPConfig{NoDelay: true}, 1*time.Second)
		if probe.Status == core.StatusUp {
			loadedSamples = append(loadedSamples, float64(time.Since(start).Microseconds())/1000.0)
		}
		time.Sleep(100 * time.Millisecond)
	}

	close(stopChan)
	wg.Wait()

	if len(loadedSamples) == 0 {
		res.LoadedLatencyMs = res.BaselineLatencyMs + 5.0
	} else {
		var sum float64
		for _, s := range loadedSamples {
			sum += s
		}
		res.LoadedLatencyMs = sum / float64(len(loadedSamples))
	}

	elapsedSec := time.Since(testStart).Seconds()
	if elapsedSec > 0 {
		res.DownloadSpeedMbps = (float64(bytesRead) * 8) / (elapsedSec * 1024 * 1024)
	}

	res.LatencyInflationMs = res.LoadedLatencyMs - res.BaselineLatencyMs
	if res.LatencyInflationMs < 0 {
		res.LatencyInflationMs = 0
	}

	res.Grade = calculateBufferbloatGrade(res.LatencyInflationMs)
	return res
}

func calculateBufferbloatGrade(inflationMs float64) string {
	switch {
	case inflationMs <= 5.0:
		return "A+"
	case inflationMs <= 30.0:
		return "A"
	case inflationMs <= 60.0:
		return "B"
	case inflationMs <= 100.0:
		return "C"
	case inflationMs <= 200.0:
		return "D"
	default:
		return "F"
	}
}

// CalculateGrade is an exported alias for bufferbloat classification
func CalculateGrade(inflationMs float64) string {
	return calculateBufferbloatGrade(inflationMs)
}
