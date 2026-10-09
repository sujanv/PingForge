package netprobe

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/sujanv/pingforge/internal/core"
)

// WellKnownServices provides common port to service mapping
var WellKnownServices = map[int]string{
	21:    "FTP",
	22:    "SSH",
	23:    "Telnet",
	25:    "SMTP",
	53:    "DNS",
	80:    "HTTP",
	110:   "POP3",
	143:   "IMAP",
	443:   "HTTPS",
	465:   "SMTPS",
	587:   "Submission",
	993:   "IMAPS",
	995:   "POP3S",
	1433:  "MSSQL",
	1521:  "Oracle",
	3306:  "MySQL",
	3389:  "RDP",
	5432:  "PostgreSQL",
	6379:  "Redis",
	8080:  "HTTP-Proxy",
	8443:  "HTTPS-Alt",
	9090:  "Prometheus",
	27017: "MongoDB",
}

// DefaultScanPorts returns standard critical ports
func DefaultScanPorts() []int {
	return []int{21, 22, 25, 53, 80, 110, 143, 443, 465, 587, 993, 995, 1433, 3306, 3389, 5432, 6379, 8080, 8443, 9090}
}

// ScanConfig configures port scanning
type ScanConfig struct {
	Ports       []int
	Concurrency int
	Timeout     time.Duration
	BannerGrab  bool
}

// ScanPorts scans ports on the host concurrently with worker pool
func ScanPorts(ctx context.Context, host string, cfg ScanConfig) *core.ScanResult {
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 20
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 1500 * time.Millisecond
	}
	if len(cfg.Ports) == 0 {
		cfg.Ports = DefaultScanPorts()
	}

	result := &core.ScanResult{
		Target:       core.Target{Host: host, Protocol: "tcp"},
		TotalScanned: len(cfg.Ports),
		OpenPorts:    make([]core.PortInfo, 0),
		Timestamp:    time.Now(),
	}

	start := time.Now()

	portCh := make(chan int, len(cfg.Ports))
	for _, p := range cfg.Ports {
		portCh <- p
	}
	close(portCh)

	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for port := range portCh {
				select {
				case <-ctx.Done():
					return
				default:
				}

				addr := net.JoinHostPort(host, strconv.Itoa(port))
				d := net.Dialer{Timeout: cfg.Timeout}

				pStart := time.Now()
				conn, err := d.DialContext(ctx, "tcp", addr)
				pDuration := time.Since(pStart)
				pDurationMs := float64(pDuration.Microseconds()) / 1000.0

				mu.Lock()
				if err == nil {
					serviceName := WellKnownServices[port]
					if serviceName == "" {
						serviceName = fmt.Sprintf("service/%d", port)
					}

					var banner string
					if cfg.BannerGrab {
						banner = grabBanner(conn, 1*time.Second)
					}
					_ = conn.Close()

					result.OpenPorts = append(result.OpenPorts, core.PortInfo{
						Port:      port,
						State:     "open",
						Service:   serviceName,
						Banner:    banner,
						LatencyMs: pDurationMs,
					})
				} else {
					result.ClosedCount++
				}
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	result.DurationMs = float64(time.Since(start).Microseconds()) / 1000.0
	return result
}
