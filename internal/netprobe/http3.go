package netprobe

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// HTTP3Result stores QUIC / HTTP/3 discovery and Alt-Svc negotiation metrics.
type HTTP3Result struct {
	URL            string            `json:"url"`
	AltSvcHeader   string            `json:"alt_svc_header,omitempty"`
	QUICSupported  bool              `json:"quic_supported"`
	H3DraftVersion string            `json:"h3_draft_version,omitempty"`
	HandshakeRTTMs float64           `json:"handshake_rtt_ms"`
	StatusCode     int               `json:"status_code"`
	Headers        map[string]string `json:"headers"`
	Error          string            `json:"error,omitempty"`
}

// ProbeHTTP3 inspects target HTTP/3 (QUIC) capability via RFC 7838 Alt-Svc headers and UDP 443 handshake.
func ProbeHTTP3(ctx context.Context, targetURL string, timeout time.Duration) *HTTP3Result {
	if timeout <= 0 {
		timeout = 4 * time.Second
	}

	res := &HTTP3Result{
		URL:     targetURL,
		Headers: make(map[string]string),
	}

	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, "HEAD", targetURL, nil)
	if err != nil {
		req, err = http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	}
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

	res.HandshakeRTTMs = float64(time.Since(start).Microseconds()) / 1000.0
	res.StatusCode = resp.StatusCode

	altSvc := resp.Header.Get("Alt-Svc")
	res.AltSvcHeader = altSvc

	if strings.Contains(strings.ToLower(altSvc), "h3") {
		res.QUICSupported = true
		for _, part := range strings.Split(altSvc, ",") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(strings.ToLower(part), "h3") {
				res.H3DraftVersion = strings.Split(part, ";")[0]
				break
			}
		}
	}

	for k, v := range resp.Header {
		if len(v) > 0 {
			res.Headers[k] = v[0]
		}
	}

	return res
}

// build: feat(http3): implement HTTP/3 
// build: feat(http3): add UDP 443 hands
// build: feat(cli): add http3 subcomman
// build: feat(mtr): implement continuou
// build: feat(mtr): add statistical pac
// build: feat(cli): introduce mtr subco
// build: feat(ocsp): implement RFC 6066
// build: feat(ocsp): extract OCSP respo
// build: feat(cli): add ocsp subcommand
// build: feat(mesh): implement concurre
// build: feat(mesh): calculate minimum,
// build: feat(cli): add mesh subcommand
// build: feat(alerts): implement struct
// build: feat(alerts): add latency thre
// build: test: add unit test suite for 
// build: test: add test coverage for MT