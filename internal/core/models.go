package core

import (
	"fmt"
	"time"
)

// ProbeType defines the diagnostic probe category.
type ProbeType string

const (
	ProbeTCP     ProbeType = "TCP"
	ProbeDNS     ProbeType = "DNS"
	ProbePing    ProbeType = "PING"
	ProbeHealth  ProbeType = "HEALTH"
	ProbeScan    ProbeType = "SCAN"
	ProbeTrace   ProbeType = "TRACE"
)

// Status represents the operational health status.
type Status string

const (
	StatusUp       Status = "UP"
	StatusDown     Status = "DOWN"
	StatusDegraded Status = "DEGRADED"
	StatusUnknown  Status = "UNKNOWN"
)

// Target defines the endpoint and configuration for a probe.
type Target struct {
	Host     string            `json:"host" yaml:"host"`
	Port     int               `json:"port,omitempty" yaml:"port,omitempty"`
	URL      string            `json:"url,omitempty" yaml:"url,omitempty"`
	Protocol string            `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	Timeout  time.Duration     `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Tags     map[string]string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// String returns a readable representation of the target.
func (t Target) String() string {
	if t.URL != "" {
		return t.URL
	}
	if t.Port > 0 {
		return fmt.Sprintf("%s:%d", t.Host, t.Port)
	}
	return t.Host
}

// TLSInfo captures SSL/TLS negotiation details.
type TLSInfo struct {
	Version             string    `json:"version,omitempty"`
	CipherSuite         string    `json:"cipher_suite,omitempty"`
	Issuer              string    `json:"issuer,omitempty"`
	Subject             string    `json:"subject,omitempty"`
	NotBefore           time.Time `json:"not_before,omitempty"`
	NotAfter            time.Time `json:"not_after,omitempty"`
	DaysUntilExpiration int       `json:"days_until_expiration"`
	Verified            bool      `json:"verified"`
	Error               string    `json:"error,omitempty"`
}

// DNSInfo captures DNS resolution details.
type DNSInfo struct {
	Resolver      string              `json:"resolver"`
	QueryTimeMs   float64             `json:"query_time_ms"`
	Records       map[string][]string `json:"records,omitempty"`
	Authoritative bool                `json:"authoritative,omitempty"`
	Error         string              `json:"error,omitempty"`
}

// HTTPInfo captures HTTP response metrics and assertions.
type HTTPInfo struct {
	StatusCode       int               `json:"status_code"`
	Proto            string            `json:"proto"`
	ContentLength    int64             `json:"content_length"`
	Headers          map[string]string `json:"headers,omitempty"`
	DNSLookupMs      float64           `json:"dns_lookup_ms"`
	TCPConnectMs     float64           `json:"tcp_connect_ms"`
	TLSHandshakeMs   float64           `json:"tls_handshake_ms"`
	TTFBMs           float64           `json:"ttfb_ms"`
	TotalDurationMs  float64           `json:"total_duration_ms"`
	RedirectCount    int               `json:"redirect_count"`
	FinalURL         string            `json:"final_url,omitempty"`
	BodyMatched      bool              `json:"body_matched"`
	TLSInfo          *TLSInfo          `json:"tls_info,omitempty"`
}

// ProbeResult encapsulates diagnostic metrics for a single probe execution.
type ProbeResult struct {
	ID        string                 `json:"id"`
	Target    Target                 `json:"target"`
	Type      ProbeType              `json:"type"`
	Status    Status                 `json:"status"`
	Latency   time.Duration          `json:"latency"`
	LatencyMs float64                `json:"latency_ms"`
	Timestamp time.Time              `json:"timestamp"`
	Message   string                 `json:"message,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	TLS       *TLSInfo               `json:"tls,omitempty"`
	DNS       *DNSInfo               `json:"dns,omitempty"`
	HTTP      *HTTPInfo              `json:"http,omitempty"`
}

// LatencyStats contains statistical analysis of round-trip times.
type LatencyStats struct {
	Target       Target    `json:"target"`
	Count        int       `json:"count"`
	Sent         int       `json:"sent"`
	Received     int       `json:"received"`
	LossRate     float64   `json:"loss_rate_percent"`
	MinMs        float64   `json:"min_ms"`
	AvgMs        float64   `json:"avg_ms"`
	MaxMs        float64   `json:"max_ms"`
	StdDevMs     float64   `json:"std_dev_ms"`
	JitterMs     float64   `json:"jitter_ms"` // RFC 3550 Interarrival Jitter
	P50Ms        float64   `json:"p50_ms"`
	P90Ms        float64   `json:"p90_ms"`
	P95Ms        float64   `json:"p95_ms"`
	P99Ms        float64   `json:"p99_ms"`
	Samples      []float64 `json:"samples,omitempty"`
	StartTime    time.Time `json:"start_time"`
	DurationMs   float64   `json:"duration_ms"`
}

// PortInfo describes the status of a specific network port.
type PortInfo struct {
	Port      int     `json:"port"`
	State     string  `json:"state"` // open, closed, filtered
	Service   string  `json:"service,omitempty"`
	Banner    string  `json:"banner,omitempty"`
	LatencyMs float64 `json:"latency_ms"`
}

// ScanResult contains results of a port scanning probe.
type ScanResult struct {
	Target        Target     `json:"target"`
	OpenPorts     []PortInfo `json:"open_ports"`
	ClosedCount   int        `json:"closed_count"`
	FilteredCount int        `json:"filtered_count"`
	TotalScanned  int        `json:"total_scanned"`
	DurationMs    float64    `json:"duration_ms"`
	Timestamp     time.Time  `json:"timestamp"`
}

// TraceHop records metrics for an intermediate network hop.
type TraceHop struct {
	Hop       int     `json:"hop"`
	IP        string  `json:"ip"`
	Hostname  string  `json:"hostname,omitempty"`
	LatencyMs float64 `json:"latency_ms"`
	Reached   bool    `json:"reached"`
	Error     string  `json:"error,omitempty"`
}

// TraceResult records the full route path.
type TraceResult struct {
	Target     Target     `json:"target"`
	Hops       []TraceHop `json:"hops"`
	TotalHops  int        `json:"total_hops"`
	Completed  bool       `json:"completed"`
	DurationMs float64    `json:"duration_ms"`
	Timestamp  time.Time  `json:"timestamp"`
}

// HistorySummary aggregates historical metrics for reporting.
type HistorySummary struct {
	Target           string    `json:"target"`
	TotalProbes      int       `json:"total_probes"`
	SuccessfulProbes int       `json:"successful_probes"`
	FailedProbes     int       `json:"failed_probes"`
	UptimePercentage float64   `json:"uptime_percentage"`
	MinLatencyMs     float64   `json:"min_latency_ms"`
	AvgLatencyMs     float64   `json:"avg_latency_ms"`
	MaxLatencyMs     float64   `json:"max_latency_ms"`
	P95LatencyMs     float64   `json:"p95_latency_ms"`
	FirstSeen        time.Time `json:"first_seen"`
	LastSeen         time.Time `json:"last_seen"`
	Sparkline        string    `json:"sparkline,omitempty"`
}
