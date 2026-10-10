package netprobe

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// WebSocketResult records connection and ping-pong latency metrics.
type WebSocketResult struct {
	URL              string  `json:"url"`
	HandshakeMs      float64 `json:"handshake_ms"`
	PingPongRTTMs    float64 `json:"ping_pong_rtt_ms"`
	TotalDurationMs  float64 `json:"total_duration_ms"`
	SecWebSocketProt string  `json:"sec_websocket_protocol,omitempty"`
	Connected        bool    `json:"connected"`
	Error            string  `json:"error,omitempty"`
}

// ProbeWebSocket tests RFC 6455 WebSocket handshake and ping-pong frame RTT.
func ProbeWebSocket(ctx context.Context, rawURL string, subprotocols []string, timeout time.Duration) *WebSocketResult {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	res := &WebSocketResult{URL: rawURL}
	u, err := url.Parse(rawURL)
	if err != nil {
		res.Error = fmt.Sprintf("invalid websocket URL: %v", err)
		return res
	}

	port := u.Port()
	isTLS := strings.EqualFold(u.Scheme, "wss")
	if port == "" {
		if isTLS {
			port = "443"
		} else {
			port = "80"
		}
	}

	addr := net.JoinHostPort(u.Hostname(), port)
	dialer := &net.Dialer{Timeout: timeout}

	start := time.Now()
	var conn net.Conn

	if isTLS {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: u.Hostname()})
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}

	if err != nil {
		res.Error = fmt.Sprintf("dial %s failed: %v", addr, err)
		return res
	}
	defer conn.Close()

	// Generate nonce key
	keyBytes := make([]byte, 16)
	_, _ = rand.Read(keyBytes)
	secKey := base64.StdEncoding.EncodeToString(keyBytes)

	reqPath := u.RequestURI()
	if reqPath == "" {
		reqPath = "/"
	}

	// Send HTTP Upgrade Request
	handshakeStart := time.Now()
	var reqBuilder strings.Builder
	reqBuilder.WriteString(fmt.Sprintf("GET %s HTTP/1.1\r\n", reqPath))
	reqBuilder.WriteString(fmt.Sprintf("Host: %s\r\n", u.Host))
	reqBuilder.WriteString("Upgrade: websocket\r\n")
	reqBuilder.WriteString("Connection: Upgrade\r\n")
	reqBuilder.WriteString(fmt.Sprintf("Sec-WebSocket-Key: %s\r\n", secKey))
	reqBuilder.WriteString("Sec-WebSocket-Version: 13\r\n")
	if len(subprotocols) > 0 {
		reqBuilder.WriteString(fmt.Sprintf("Sec-WebSocket-Protocol: %s\r\n", strings.Join(subprotocols, ", ")))
	}
	reqBuilder.WriteString("\r\n")

	_ = conn.SetDeadline(time.Now().Add(timeout))
	if _, err := conn.Write([]byte(reqBuilder.String())); err != nil {
		res.Error = fmt.Sprintf("write handshake: %v", err)
		return res
	}

	reader := bufio.NewReader(conn)
	resp, err := http.ReadResponse(reader, &http.Request{Method: "GET"})
	if err != nil {
		res.Error = fmt.Sprintf("read handshake response: %v", err)
		return res
	}
	res.HandshakeMs = float64(time.Since(handshakeStart).Microseconds()) / 1000.0

	if resp.StatusCode != http.StatusSwitchingProtocols {
		res.Error = fmt.Sprintf("unexpected HTTP status: %d (expected 101)", resp.StatusCode)
		return res
	}

	// Verify Sec-WebSocket-Accept
	expectedAccept := computeWebSocketAccept(secKey)
	actualAccept := resp.Header.Get("Sec-WebSocket-Accept")
	if actualAccept != expectedAccept {
		res.Error = fmt.Sprintf("invalid Sec-WebSocket-Accept key (expected %s, got %s)", expectedAccept, actualAccept)
		return res
	}

	res.Connected = true
	res.SecWebSocketProt = resp.Header.Get("Sec-WebSocket-Protocol")

	// Send Ping frame (0x89) with 4-byte mask
	pingPayload := []byte("PINGFORGE")
	pingFrame := make([]byte, 6+len(pingPayload))
	pingFrame[0] = 0x89 // FIN + Opcode 9 (Ping)
	pingFrame[1] = 0x80 | byte(len(pingPayload)) // Masked + payload length
	maskKey := []byte{0x12, 0x34, 0x56, 0x78}
	copy(pingFrame[2:6], maskKey)
	for i, b := range pingPayload {
		pingFrame[6+i] = b ^ maskKey[i%4]
	}

	pingStart := time.Now()
	if _, err := conn.Write(pingFrame); err != nil {
		res.Error = fmt.Sprintf("write ping frame: %v", err)
		return res
	}

	// Read frame header for Pong (0x8A)
	header := make([]byte, 2)
	if _, err := reader.Read(header); err == nil {
		res.PingPongRTTMs = float64(time.Since(pingStart).Microseconds()) / 1000.0
	}
	res.TotalDurationMs = float64(time.Since(start).Microseconds()) / 1000.0

	return res
}

func computeWebSocketAccept(key string) string {
	const magic = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	h := sha1.New()
	h.Write([]byte(key + magic))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// Updated for feat(websocket): add ping-pong frame round-trip latency measurement
