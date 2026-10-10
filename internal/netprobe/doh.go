package netprobe

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"time"
)

// DoHProvider defines known public DoH endpoints.
var DoHProviders = map[string]string{
	"cloudflare": "https://cloudflare-dns.com/dns-query",
	"google":     "https://dns.google/resolve",
	"quad9":      "https://dns.quad9.net/dns-query",
	"adguard":    "https://dns.adguard-dns.com/dns-query",
}

// DoHResponse represents a RFC 8427 / Google JSON DNS response.
type DoHResponse struct {
	Status   int  `json:"Status"`
	TC       bool `json:"TC"`
	RD       bool `json:"RD"`
	RA       bool `json:"RA"`
	AD       bool `json:"AD"`
	CD       bool `json:"CD"`
	Question []struct {
		Name string `json:"name"`
		Type int    `json:"type"`
	} `json:"Question"`
	Answer []struct {
		Name string `json:"name"`
		Type int    `json:"type"`
		TTL  int    `json:"TTL"`
		Data string `json:"data"`
	} `json:"Answer"`
}

// DoHResult encapsulates DNS-over-HTTPS diagnostics.
type DoHResult struct {
	Provider        string              `json:"provider"`
	EndpointURL     string              `json:"endpoint_url"`
	Domain          string              `json:"domain"`
	RecordType      string              `json:"record_type"`
	HTTPStatusCode  int                 `json:"http_status_code"`
	DNSStatusCode   int                 `json:"dns_status_code"`
	DNSLookupMs     float64             `json:"dns_lookup_ms"`
	TCPConnectMs    float64             `json:"tcp_connect_ms"`
	TLSHandshakeMs  float64             `json:"tls_handshake_ms"`
	TTFBMs          float64             `json:"ttfb_ms"`
	TotalDurationMs float64             `json:"total_duration_ms"`
	Records         map[string][]string `json:"records"`
	Error           string              `json:"error,omitempty"`
}

// ProbeDoH performs a DNS query over HTTPS to the specified provider endpoint.
func ProbeDoH(ctx context.Context, domain, providerName, recordType string, timeout time.Duration) *DoHResult {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	if recordType == "" {
		recordType = "A"
	}

	endpoint, ok := DoHProviders[strings.ToLower(providerName)]
	if !ok {
		if strings.HasPrefix(providerName, "http://") || strings.HasPrefix(providerName, "https://") {
			endpoint = providerName
		} else {
			endpoint = DoHProviders["cloudflare"]
		}
	}

	res := &DoHResult{
		Provider:    providerName,
		EndpointURL: endpoint,
		Domain:      domain,
		RecordType:  recordType,
		Records:     make(map[string][]string),
	}

	// Prepare request URL with query parameters
	u, err := url.Parse(endpoint)
	if err != nil {
		res.Error = fmt.Sprintf("invalid DoH endpoint: %v", err)
		return res
	}

	q := u.Query()
	q.Set("name", domain)
	q.Set("type", recordType)
	u.RawQuery = q.Encode()

	var dnsStart, dnsDone, connStart, connDone, tlsStart, tlsDone, ttfbStart time.Time

	trace := &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) { dnsStart = time.Now() },
		DNSDone:  func(_ httptrace.DNSDoneInfo) { dnsDone = time.Now() },
		ConnectStart: func(_, _ string) { connStart = time.Now() },
		ConnectDone: func(_, _ string, _ error) { connDone = time.Now() },
		TLSHandshakeStart: func() { tlsStart = time.Now() },
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) { tlsDone = time.Now() },
		GotFirstResponseByte: func() { ttfbStart = time.Now() },
	}

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(httptrace.WithClientTrace(reqCtx, trace), http.MethodGet, u.String(), nil)
	if err != nil {
		res.Error = err.Error()
		return res
	}

	req.Header.Set("Accept", "application/dns-json")
	req.Header.Set("User-Agent", "PingForge/1.0 (DoH Probe)")

	client := &http.Client{Timeout: timeout}
	start := time.Now()
	resp, err := client.Do(req)
	res.TotalDurationMs = float64(time.Since(start).Microseconds()) / 1000.0

	if err != nil {
		res.Error = err.Error()
		return res
	}
	defer resp.Body.Close()

	res.HTTPStatusCode = resp.StatusCode

	if !dnsStart.IsZero() && !dnsDone.IsZero() {
		res.DNSLookupMs = float64(dnsDone.Sub(dnsStart).Microseconds()) / 1000.0
	}
	if !connStart.IsZero() && !connDone.IsZero() {
		res.TCPConnectMs = float64(connDone.Sub(connStart).Microseconds()) / 1000.0
	}
	if !tlsStart.IsZero() && !tlsDone.IsZero() {
		res.TLSHandshakeMs = float64(tlsDone.Sub(tlsStart).Microseconds()) / 1000.0
	}
	if !ttfbStart.IsZero() {
		res.TTFBMs = float64(ttfbStart.Sub(start).Microseconds()) / 1000.0
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		res.Error = fmt.Sprintf("read DoH body: %v", err)
		return res
	}

	if resp.StatusCode != http.StatusOK {
		res.Error = fmt.Sprintf("DoH provider returned HTTP %d: %s", resp.StatusCode, string(body))
		return res
	}

	var dohResp DoHResponse
	if err := json.Unmarshal(body, &dohResp); err != nil {
		res.Error = fmt.Sprintf("parse DoH JSON: %v", err)
		return res
	}

	res.DNSStatusCode = dohResp.Status
	for _, ans := range dohResp.Answer {
		res.Records[recordType] = append(res.Records[recordType], ans.Data)
	}

	return res
}
