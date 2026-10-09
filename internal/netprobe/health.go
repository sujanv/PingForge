package netprobe

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"regexp"
	"strings"
	"time"

	"github.com/sujanv/pingforge/internal/core"
)

// HealthCheckOptions provides parameters for HTTP/HTTPS health checks.
type HealthCheckOptions struct {
	ExpectedStatus  []int
	BodyRegex       string
	FollowRedirects bool
	MaxRedirects    int
	Headers         map[string]string
	UserAgent       string
	LatencyBudget   time.Duration
	Timeout         time.Duration
}

// ProbeHealth performs a deep HTTP/HTTPS health check with detailed timing phases (DNS, TCP, TLS, TTFB).
func ProbeHealth(ctx context.Context, targetURL string, opts HealthCheckOptions) *core.ProbeResult {
	if opts.Timeout <= 0 {
		opts.Timeout = 10 * time.Second
	}
	if len(opts.ExpectedStatus) == 0 {
		opts.ExpectedStatus = []int{200, 201, 204}
	}
	if opts.UserAgent == "" {
		opts.UserAgent = "PingForge/1.0"
	}

	target := core.Target{
		URL:      targetURL,
		Protocol: "http",
		Timeout:  opts.Timeout,
	}

	res := &core.ProbeResult{
		ID:        fmt.Sprintf("health-%d", time.Now().UnixNano()),
		Target:    target,
		Type:      core.ProbeHealth,
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	var (
		dnsStart, dnsDone       time.Time
		connStart, connDone     time.Time
		tlsStart, tlsDone       time.Time
		firstByteTime           time.Time
		redirectCount           int
	)

	trace := &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) { dnsStart = time.Now() },
		DNSDone:  func(_ httptrace.DNSDoneInfo) { dnsDone = time.Now() },
		ConnectStart: func(_, _ string) { connStart = time.Now() },
		ConnectDone: func(_, _ string, _ error) { connDone = time.Now() },
		TLSHandshakeStart: func() { tlsStart = time.Now() },
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) { tlsDone = time.Now() },
		GotFirstResponseByte: func() { firstByteTime = time.Now() },
	}

	reqCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(httptrace.WithClientTrace(reqCtx, trace), http.MethodGet, targetURL, nil)
	if err != nil {
		res.Status = core.StatusDown
		res.Error = err.Error()
		res.Message = fmt.Sprintf("invalid request: %v", err)
		return res
	}

	req.Header.Set("User-Agent", opts.UserAgent)
	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{
		Timeout: opts.Timeout,
	}

	if !opts.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	} else {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			redirectCount = len(via)
			if redirectCount > opts.MaxRedirects && opts.MaxRedirects > 0 {
				return fmt.Errorf("stopped after %d redirects", opts.MaxRedirects)
			}
			return nil
		}
	}

	overallStart := time.Now()
	resp, err := client.Do(req)
	totalDuration := time.Since(overallStart)

	res.Latency = totalDuration
	res.LatencyMs = float64(totalDuration.Microseconds()) / 1000.0

	if err != nil {
		res.Status = core.StatusDown
		res.Error = err.Error()
		res.Message = fmt.Sprintf("HTTP request error: %v", err)
		return res
	}
	defer resp.Body.Close()

	// Read body (up to 1MB)
	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))

	httpInfo := &core.HTTPInfo{
		StatusCode:      resp.StatusCode,
		Proto:           resp.Proto,
		ContentLength:   resp.ContentLength,
		RedirectCount:   redirectCount,
		FinalURL:        resp.Request.URL.String(),
		Headers:         make(map[string]string),
		TotalDurationMs: res.LatencyMs,
	}

	for k, v := range resp.Header {
		httpInfo.Headers[k] = strings.Join(v, ", ")
	}

	if !dnsStart.IsZero() && !dnsDone.IsZero() {
		httpInfo.DNSLookupMs = float64(dnsDone.Sub(dnsStart).Microseconds()) / 1000.0
	}
	if !connStart.IsZero() && !connDone.IsZero() {
		httpInfo.TCPConnectMs = float64(connDone.Sub(connStart).Microseconds()) / 1000.0
	}
	if !tlsStart.IsZero() && !tlsDone.IsZero() {
		httpInfo.TLSHandshakeMs = float64(tlsDone.Sub(tlsStart).Microseconds()) / 1000.0
	}
	if !firstByteTime.IsZero() {
		httpInfo.TTFBMs = float64(firstByteTime.Sub(overallStart).Microseconds()) / 1000.0
	}

	// TLS cert inspection if available
	if resp.TLS != nil {
		res.TLS = extractTLSDetails(*resp.TLS)
	}

	// Validate status code
	statusOk := false
	for _, expected := range opts.ExpectedStatus {
		if resp.StatusCode == expected {
			statusOk = true
			break
		}
	}

	// Validate body regex if specified
	bodyOk := true
	if opts.BodyRegex != "" {
		matched, _ := regexp.Match(opts.BodyRegex, bodyBytes)
		bodyOk = matched
		httpInfo.BodyMatched = matched
	} else {
		httpInfo.BodyMatched = true
	}

	res.HTTP = httpInfo

	if !statusOk {
		res.Status = core.StatusDown
		res.Message = fmt.Sprintf("Unexpected status code %d (expected %v)", resp.StatusCode, opts.ExpectedStatus)
	} else if !bodyOk {
		res.Status = core.StatusDegraded
		res.Message = fmt.Sprintf("Status %d OK, but body regex %q did not match", resp.StatusCode, opts.BodyRegex)
	} else if opts.LatencyBudget > 0 && totalDuration > opts.LatencyBudget {
		res.Status = core.StatusDegraded
		res.Message = fmt.Sprintf("Status %d OK, but latency %.2fms exceeded budget %.2fms", resp.StatusCode, res.LatencyMs, float64(opts.LatencyBudget.Milliseconds()))
	} else {
		res.Status = core.StatusUp
		res.Message = fmt.Sprintf("Healthy (Status %d, TTFB: %.2fms, Total: %.2fms)", resp.StatusCode, httpInfo.TTFBMs, res.LatencyMs)
	}

	return res
}
