package netprobe

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"
)

// HTTP2Result contains HTTP/2 multiplexing probe results
type HTTP2Result struct {
	URL              string            `json:"url"`
	NegotiatedALPN   string            `json:"alpn"`
	ProtocolVersion  string            `json:"protocol_version"`
	StatusCode       int               `json:"status_code"`
	MultiplexCapable bool              `json:"multiplex_capable"`
	StreamHeaders    map[string]string `json:"stream_headers"`
	HandshakeMs      float64           `json:"handshake_ms"`
	TotalDurationMs  float64           `json:"total_duration_ms"`
	Error            string            `json:"error,omitempty"`
}

// ProbeHTTP2 tests RFC 7540 HTTP/2 connectivity and ALPN negotiation
func ProbeHTTP2(ctx context.Context, targetURL string, timeout time.Duration) *HTTP2Result {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	res := &HTTP2Result{
		URL:           targetURL,
		StreamHeaders: make(map[string]string),
	}

	tlsConfig := &tls.Config{
		NextProtos:         []string{"h2", "http/1.1"},
		InsecureSkipVerify: true,
	}

	transport := &http.Transport{
		TLSClientConfig:   tlsConfig,
		ForceAttemptHTTP2: true,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		res.Error = err.Error()
		return res
	}

	resp, err := client.Do(req)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	defer resp.Body.Close()

	res.TotalDurationMs = float64(time.Since(start).Microseconds()) / 1000.0
	res.StatusCode = resp.StatusCode
	res.ProtocolVersion = resp.Proto

	if resp.TLS != nil {
		res.NegotiatedALPN = resp.TLS.NegotiatedProtocol
		if resp.TLS.NegotiatedProtocol == "h2" || resp.ProtoMajor == 2 {
			res.MultiplexCapable = true
		}
	}

	for k, v := range resp.Header {
		if len(v) > 0 {
			res.StreamHeaders[k] = v[0]
		}
	}

	return res
}
