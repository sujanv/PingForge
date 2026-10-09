package netprobe

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/sujanv/pingforge/internal/config"
	"github.com/sujanv/pingforge/internal/core"
)

// ProbeTCP performs a TCP connectivity check with latency measurement,
// optional TLS inspection, socket keep-alive, and banner grabbing.
func ProbeTCP(ctx context.Context, host string, port int, cfg config.TCPConfig, timeout time.Duration) *core.ProbeResult {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	target := core.Target{
		Host:     host,
		Port:     port,
		Protocol: "tcp",
		Timeout:  timeout,
	}

	res := &core.ProbeResult{
		ID:        fmt.Sprintf("tcp-%s-%d-%d", host, port, time.Now().UnixNano()),
		Target:    target,
		Type:      core.ProbeTCP,
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	address := net.JoinHostPort(host, strconv.Itoa(port))

	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: cfg.KeepAlive,
	}

	start := time.Now()
	conn, err := dialer.DialContext(ctx, "tcp", address)
	latency := time.Since(start)

	res.Latency = latency
	res.LatencyMs = float64(latency.Microseconds()) / 1000.0

	if err != nil {
		res.Status = core.StatusDown
		res.Error = err.Error()
		res.Message = fmt.Sprintf("TCP dial failed: %v", err)
		return res
	}
	defer conn.Close()

	res.Status = core.StatusUp
	res.Message = fmt.Sprintf("Connected to %s (%s) in %.2fms", address, conn.RemoteAddr().String(), res.LatencyMs)
	res.Details["remote_addr"] = conn.RemoteAddr().String()
	res.Details["local_addr"] = conn.LocalAddr().String()

	// Apply socket options if TCPConn
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		if cfg.KeepAlive > 0 {
			_ = tcpConn.SetKeepAlive(true)
			_ = tcpConn.SetKeepAlivePeriod(cfg.KeepAlive)
		}
		_ = tcpConn.SetNoDelay(cfg.NoDelay)
	}

	// Banner Grab if requested
	if cfg.ReadBanner {
		banner := grabBanner(conn, 1500*time.Millisecond)
		if banner != "" {
			res.Details["banner"] = banner
		}
	}

	// TLS handshake inspection if requested or port is typical TLS port
	if cfg.TLS || port == 443 || port == 8443 {
		tlsInfo := inspectTLS(ctx, host, port, timeout)
		res.TLS = tlsInfo
		if tlsInfo != nil && tlsInfo.Error != "" && res.Status == core.StatusUp {
			res.Status = core.StatusDegraded
		}
	}

	return res
}

func grabBanner(conn net.Conn, timeout time.Duration) string {
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	reader := bufio.NewReader(conn)

	// Try reading immediately without sending payload
	lineCh := make(chan string, 1)
	go func() {
		line, err := reader.ReadString('\n')
		if err == nil {
			lineCh <- strings.TrimSpace(line)
			return
		}
		// If no initial greeting, try probe string
		_, _ = conn.Write([]byte("HEAD / HTTP/1.0\r\n\r\n"))
		line, err = reader.ReadString('\n')
		if err == nil {
			lineCh <- strings.TrimSpace(line)
			return
		}
		lineCh <- ""
	}()

	select {
	case line := <-lineCh:
		if len(line) > 256 {
			return line[:256] + "..."
		}
		return line
	case <-time.After(timeout):
		return ""
	}
}

func inspectTLS(ctx context.Context, host string, port int, timeout time.Duration) *core.TLSInfo {
	address := net.JoinHostPort(host, strconv.Itoa(port))
	dialer := &net.Dialer{Timeout: timeout}

	tlsConfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: false,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", address, tlsConfig)
	if err != nil {
		// Try again with insecure to inspect cert anyway
		insecureConfig := &tls.Config{
			ServerName:         host,
			InsecureSkipVerify: true,
		}
		insecureConn, inErr := tls.DialWithDialer(dialer, "tcp", address, insecureConfig)
		if inErr != nil {
			return &core.TLSInfo{
				Verified: false,
				Error:    err.Error(),
			}
		}
		defer insecureConn.Close()
		state := insecureConn.ConnectionState()
		info := extractTLSDetails(state)
		info.Verified = false
		info.Error = fmt.Sprintf("untrusted cert: %v", err)
		return info
	}
	defer conn.Close()

	state := conn.ConnectionState()
	info := extractTLSDetails(state)
	info.Verified = true
	return info
}

func extractTLSDetails(state tls.ConnectionState) *core.TLSInfo {
	info := &core.TLSInfo{
		Version:     tlsVersionString(state.Version),
		CipherSuite: tls.CipherSuiteName(state.CipherSuite),
	}

	if len(state.PeerCertificates) > 0 {
		cert := state.PeerCertificates[0]
		info.Subject = cert.Subject.CommonName
		info.Issuer = cert.Issuer.CommonName
		info.NotBefore = cert.NotBefore
		info.NotAfter = cert.NotAfter
		info.DaysUntilExpiration = int(time.Until(cert.NotAfter).Hours() / 24)
	}

	return info
}

func tlsVersionString(ver uint16) string {
	switch ver {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("0x%04x", ver)
	}
}
