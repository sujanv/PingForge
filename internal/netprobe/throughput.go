package netprobe

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// ThroughputResult stores bandwidth measurement metrics.
type ThroughputResult struct {
	Target          string  `json:"target"`
	BytesTransferred int64  `json:"bytes_transferred"`
	DurationSeconds float64 `json:"duration_seconds"`
	ThroughputMbps  float64 `json:"throughput_mbps"`
	ThroughputMBps  float64 `json:"throughput_mb_s"`
	Error           string  `json:"error,omitempty"`
}

// ProbeThroughput measures download speed and bandwidth against target.
func ProbeThroughput(ctx context.Context, target string, testDuration time.Duration) *ThroughputResult {
	if testDuration <= 0 {
		testDuration = 3 * time.Second
	}

	res := &ThroughputResult{Target: target}

	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return probeHTTPThroughput(ctx, target, testDuration)
	}

	// Raw TCP streaming throughput
	if !strings.Contains(target, ":") {
		target += ":80"
	}

	d := net.Dialer{Timeout: 5 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", target)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	defer conn.Close()

	// Send basic HTTP probe
	_, _ = conn.Write([]byte("GET / HTTP/1.1\r\nHost: " + target + "\r\nConnection: close\r\n\r\n"))

	start := time.Now()
	buf := make([]byte, 32*1024)
	var totalBytes int64

	for time.Since(start) < testDuration {
		_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		n, err := conn.Read(buf)
		if n > 0 {
			totalBytes += int64(n)
		}
		if err != nil {
			break
		}
	}

	durSec := time.Since(start).Seconds()
	res.BytesTransferred = totalBytes
	res.DurationSeconds = durSec
	if durSec > 0 {
		res.ThroughputMBps = (float64(totalBytes) / (1024 * 1024)) / durSec
		res.ThroughputMbps = (float64(totalBytes*8) / (1000 * 1000)) / durSec
	}

	return res
}

func probeHTTPThroughput(ctx context.Context, urlStr string, testDuration time.Duration) *ThroughputResult {
	res := &ThroughputResult{Target: urlStr}
	reqCtx, cancel := context.WithTimeout(ctx, testDuration+2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, urlStr, nil)
	if err != nil {
		res.Error = err.Error()
		return res
	}

	client := &http.Client{}
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	defer resp.Body.Close()

	buf := make([]byte, 32*1024)
	var totalBytes int64

	for time.Since(start) < testDuration {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			totalBytes += int64(n)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			break
		}
	}

	durSec := time.Since(start).Seconds()
	res.BytesTransferred = totalBytes
	res.DurationSeconds = durSec
	if durSec > 0 {
		res.ThroughputMBps = (float64(totalBytes) / (1024 * 1024)) / durSec
		res.ThroughputMbps = (float64(totalBytes*8) / (1000 * 1000)) / durSec
	}

	return res
}
