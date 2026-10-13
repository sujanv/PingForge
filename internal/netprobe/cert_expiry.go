package netprobe

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

// CertExpirySummary details SSL/TLS certificate lifetime and renewal status.
type CertExpirySummary struct {
	Host               string    `json:"host"`
	Port               int       `json:"port"`
	SubjectCN          string    `json:"subject_cn"`
	Issuer             string    `json:"issuer"`
	NotBefore          time.Time `json:"not_before"`
	NotAfter           time.Time `json:"not_after"`
	DaysRemaining      int       `json:"days_remaining"`
	HoursRemaining     int       `json:"hours_remaining"`
	UrgencyStatus      string    `json:"urgency_status"` // OK, WARNING, CRITICAL, EXPIRED
	SANs               []string  `json:"sans"`
	SignatureAlgorithm string    `json:"signature_algorithm"`
	Error              string    `json:"error,omitempty"`
}

// AuditCertExpiry connects to a target host and computes certificate validity window.
func AuditCertExpiry(ctx context.Context, host string, port int, timeout time.Duration) *CertExpirySummary {
	if port <= 0 {
		port = 443
	}
	if timeout <= 0 {
		timeout = 4 * time.Second
	}

	summary := &CertExpirySummary{
		Host:          host,
		Port:          port,
		UrgencyStatus: "UNKNOWN",
		SANs:          make([]string, 0),
	}

	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	dialer := &net.Dialer{Timeout: timeout}

	tlsConfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: true,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	if err != nil {
		summary.Error = fmt.Sprintf("tls dial failed: %v", err)
		return summary
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		summary.Error = "no peer certificates found"
		return summary
	}

	leaf := certs[0]
	summary.SubjectCN = leaf.Subject.CommonName
	summary.Issuer = leaf.Issuer.CommonName
	summary.NotBefore = leaf.NotBefore
	summary.NotAfter = leaf.NotAfter
	summary.SANs = leaf.DNSNames
	summary.SignatureAlgorithm = leaf.SignatureAlgorithm.String()

	now := time.Now()
	remDuration := leaf.NotAfter.Sub(now)
	summary.DaysRemaining = int(remDuration.Hours() / 24)
	summary.HoursRemaining = int(remDuration.Hours())

	switch {
	case summary.DaysRemaining <= 0:
		summary.UrgencyStatus = "EXPIRED"
	case summary.DaysRemaining <= 7:
		summary.UrgencyStatus = "CRITICAL"
	case summary.DaysRemaining <= 30:
		summary.UrgencyStatus = "WARNING"
	default:
		summary.UrgencyStatus = "OK"
	}

	return summary
}
