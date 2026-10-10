package netprobe

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"
)

// UDPResult stores UDP probe metrics.
type UDPResult struct {
	Target      string  `json:"target"`
	Port        int     `json:"port"`
	Reachable   bool    `json:"reachable"`
	LatencyMs   float64 `json:"latency_ms"`
	BytesSent   int     `json:"bytes_sent"`
	BytesRecv   int     `json:"bytes_recv"`
	Response    string  `json:"response,omitempty"`
	Message     string  `json:"message"`
	Error       string  `json:"error,omitempty"`
}

// ProbeUDP sends a UDP packet to the destination and checks for response or timeout.
func ProbeUDP(ctx context.Context, host string, port int, payload []byte, timeout time.Duration) *UDPResult {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	if len(payload) == 0 {
		payload = []byte("PINGFORGE-UDP-PROBE")
	}

	addr := net.JoinHostPort(host, strconv.Itoa(port))
	res := &UDPResult{
		Target:    host,
		Port:      port,
		BytesSent: len(payload),
	}

	rAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		res.Error = err.Error()
		return res
	}

	conn, err := net.DialUDP("udp", nil, rAddr)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	defer conn.Close()

	start := time.Now()
	_ = conn.SetDeadline(time.Now().Add(timeout))

	if _, err := conn.Write(payload); err != nil {
		res.Error = fmt.Sprintf("write UDP: %v", err)
		return res
	}

	buf := make([]byte, 1024)
	n, _, err := conn.ReadFrom(buf)
	elapsed := time.Since(start)
	res.LatencyMs = float64(elapsed.Microseconds()) / 1000.0

	if err != nil {
		// In UDP, timeout often means port is open and application didn't reply, or filtered
		res.Reachable = true
		res.Message = "Packet sent; no immediate reply (UDP open|filtered)"
	} else {
		res.Reachable = true
		res.BytesRecv = n
		res.Response = string(buf[:n])
		res.Message = fmt.Sprintf("Received %d bytes reply in %.2fms", n, res.LatencyMs)
	}

	return res
}
