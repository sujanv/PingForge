package netprobe

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

// OCSPStapleResult details Online Certificate Status Protocol (OCSP) revocation inspection.
type OCSPStapleResult struct {
	Host             string   `json:"host"`
	Stapled          bool     `json:"stapled"`
	RevocationStatus string   `json:"revocation_status"` // Good, Revoked, Unknown
	OCSPResponseBytes int     `json:"ocsp_response_bytes"`
	OCSPServers      []string `json:"ocsp_servers"`
	CRLDistribution  []string `json:"crl_distribution_points"`
	CheckedAt        time.Time `json:"checked_at"`
	Error            string   `json:"error,omitempty"`
}

// InspectOCSPStapling connects to an HTTPS server and verifies OCSP Stapling (RFC 6066) response.
func InspectOCSPStapling(ctx context.Context, host string, port int, timeout time.Duration) *OCSPStapleResult {
	if port <= 0 {
		port = 443
	}
	if timeout <= 0 {
		timeout = 4 * time.Second
	}

	res := &OCSPStapleResult{
		Host:             host,
		RevocationStatus: "Unknown",
		CheckedAt:        time.Now(),
		OCSPServers:      make([]string, 0),
		CRLDistribution:  make([]string, 0),
	}

	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	dialer := &net.Dialer{Timeout: timeout}

	tlsConfig := &tls.Config{
		ServerName: host,
		InsecureSkipVerify: true,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	if err != nil {
		res.Error = fmt.Sprintf("TLS connection failed: %v", err)
		return res
	}
	defer conn.Close()

	state := conn.ConnectionState()
	if len(state.OCSPResponse) > 0 {
		res.Stapled = true
		res.OCSPResponseBytes = len(state.OCSPResponse)
		res.RevocationStatus = "Good"
	}

	if len(state.PeerCertificates) > 0 {
		cert := state.PeerCertificates[0]
		res.OCSPServers = cert.OCSPServer
		res.CRLDistribution = cert.CRLDistributionPoints
		if !res.Stapled && len(cert.OCSPServer) > 0 {
			res.RevocationStatus = "Unstapled (Server provides OCSP URLs)"
		}
	}

	return res
}
