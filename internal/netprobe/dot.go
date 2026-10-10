package netprobe

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

// DoTResult contains DNS-over-TLS query metrics.
type DoTResult struct {
	Resolver        string    `json:"resolver"`
	Domain          string    `json:"domain"`
	TCPConnectMs    float64   `json:"tcp_connect_ms"`
	TLSHandshakeMs  float64   `json:"tls_handshake_ms"`
	QueryLatencyMs  float64   `json:"query_latency_ms"`
	TotalDurationMs float64   `json:"total_duration_ms"`
	TLSVersion      string    `json:"tls_version"`
	CipherSuite     string    `json:"cipher_suite"`
	AnswersCount    int       `json:"answers_count"`
	ServerName      string    `json:"server_name"`
	Error           string    `json:"error,omitempty"`
}

// ProbeDoT connects to a DNS-over-TLS resolver (port 853) and measures handshake and wire query latency.
func ProbeDoT(ctx context.Context, domain, resolverAddr string, timeout time.Duration) *DoTResult {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	if resolverAddr == "" {
		resolverAddr = "1.1.1.1:853"
	}
	if !strings.Contains(resolverAddr, ":") {
		resolverAddr += ":853"
	}

	host, _, err := net.SplitHostPort(resolverAddr)
	if err != nil {
		host = resolverAddr
	}

	res := &DoTResult{
		Resolver:   resolverAddr,
		Domain:     domain,
		ServerName: host,
	}

	start := time.Now()
	dialer := &net.Dialer{Timeout: timeout}

	connStart := time.Now()
	rawConn, err := dialer.DialContext(ctx, "tcp", resolverAddr)
	if err != nil {
		res.Error = fmt.Sprintf("TCP dial to %s failed: %v", resolverAddr, err)
		return res
	}
	defer rawConn.Close()
	res.TCPConnectMs = float64(time.Since(connStart).Microseconds()) / 1000.0

	tlsConfig := &tls.Config{
		ServerName: host,
	}
	tlsConn := tls.Client(rawConn, tlsConfig)

	handshakeStart := time.Now()
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		res.Error = fmt.Sprintf("TLS handshake failed: %v", err)
		return res
	}
	res.TLSHandshakeMs = float64(time.Since(handshakeStart).Microseconds()) / 1000.0

	state := tlsConn.ConnectionState()
	res.TLSVersion = tlsVersionString(state.Version)
	res.CipherSuite = tls.CipherSuiteName(state.CipherSuite)

	// Build standard DNS wire format query for A record
	wireQuery := buildDNSQuery(domain)
	lengthPrefix := make([]byte, 2)
	binary.BigEndian.PutUint16(lengthPrefix, uint16(len(wireQuery)))

	queryStart := time.Now()
	if _, err := tlsConn.Write(append(lengthPrefix, wireQuery...)); err != nil {
		res.Error = fmt.Sprintf("send wire query: %v", err)
		return res
	}

	// Read 2-byte response length
	_ = tlsConn.SetReadDeadline(time.Now().Add(timeout))
	respLenBuf := make([]byte, 2)
	if _, err := io.ReadFull(tlsConn, respLenBuf); err != nil {
		res.Error = fmt.Sprintf("read response length: %v", err)
		return res
	}
	respLen := binary.BigEndian.Uint16(respLenBuf)

	respBuf := make([]byte, respLen)
	if _, err := io.ReadFull(tlsConn, respBuf); err != nil {
		res.Error = fmt.Sprintf("read response payload: %v", err)
		return res
	}
	res.QueryLatencyMs = float64(time.Since(queryStart).Microseconds()) / 1000.0
	res.TotalDurationMs = float64(time.Since(start).Microseconds()) / 1000.0

	if len(respBuf) >= 8 {
		res.AnswersCount = int(binary.BigEndian.Uint16(respBuf[6:8]))
	}

	return res
}

func buildDNSQuery(domain string) []byte {
	var buf []byte
	// Transaction ID (2 bytes)
	buf = append(buf, 0x12, 0x34)
	// Flags: Standard query (2 bytes)
	buf = append(buf, 0x01, 0x00)
	// QDCOUNT: 1 question (2 bytes)
	buf = append(buf, 0x00, 0x01)
	// ANCOUNT, NSCOUNT, ARCOUNT: 0 (6 bytes)
	buf = append(buf, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00)

	// QNAME
	parts := strings.Split(domain, ".")
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}
		buf = append(buf, byte(len(part)))
		buf = append(buf, []byte(part)...)
	}
	buf = append(buf, 0x00) // End of domain

	// QTYPE: A (1) (2 bytes)
	buf = append(buf, 0x00, 0x01)
	// QCLASS: IN (1) (2 bytes)
	buf = append(buf, 0x00, 0x01)

	return buf
}
