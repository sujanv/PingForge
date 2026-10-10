package netprobe

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// CertificateSummary details a single certificate in the chain.
type CertificateSummary struct {
	Subject       string    `json:"subject"`
	Issuer        string    `json:"issuer"`
	NotBefore     time.Time `json:"not_before"`
	NotAfter      time.Time `json:"not_after"`
	DaysRemaining int       `json:"days_remaining"`
	SignatureAlgo string    `json:"signature_algorithm"`
	PublicKeyType string    `json:"public_key_type"`
	SANs          []string  `json:"sans,omitempty"`
	IsCA          bool      `json:"is_ca"`
}

// TLSSecurityAudit results for a host.
type TLSSecurityAudit struct {
	Host             string               `json:"host"`
	Port             int                  `json:"port"`
	Grade            string               `json:"grade"` // A+, A, B, C, F
	SupportedProtocols map[string]bool    `json:"supported_protocols"`
	SelectedVersion  string               `json:"selected_version"`
	SelectedCipher   string               `json:"selected_cipher"`
	ALPNNegotiated   string               `json:"alpn_negotiated,omitempty"`
	CertChain        []CertificateSummary `json:"cert_chain"`
	Warnings         []string             `json:"warnings"`
	Vulnerabilities  []string             `json:"vulnerabilities"`
	LatencyMs        float64              `json:"latency_ms"`
}

// AuditTLS evaluates TLS configuration, cipher suites, protocol versions, and certificates.
func AuditTLS(ctx context.Context, host string, port int, timeout time.Duration) *TLSSecurityAudit {
	if port <= 0 {
		port = 443
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	audit := &TLSSecurityAudit{
		Host:               host,
		Port:               port,
		SupportedProtocols: make(map[string]bool),
		CertChain:          make([]CertificateSummary, 0),
		Warnings:           make([]string, 0),
		Vulnerabilities:    make([]string, 0),
	}

	addr := net.JoinHostPort(host, strconv.Itoa(port))

	// Test protocol versions
	versions := map[string]uint16{
		"TLS 1.0": tls.VersionTLS10,
		"TLS 1.1": tls.VersionTLS11,
		"TLS 1.2": tls.VersionTLS12,
		"TLS 1.3": tls.VersionTLS13,
	}

	for vName, vCode := range versions {
		supported := testTLSVersion(ctx, addr, host, vCode, timeout)
		audit.SupportedProtocols[vName] = supported
		if (vName == "TLS 1.0" || vName == "TLS 1.1") && supported {
			audit.Vulnerabilities = append(audit.Vulnerabilities, fmt.Sprintf("Supports deprecated and insecure %s", vName))
		}
	}

	// Main handshake with ALPN inspection
	dialer := &net.Dialer{Timeout: timeout}
	start := time.Now()
	tlsConfig := &tls.Config{
		ServerName:         host,
		NextProtos:         []string{"h2", "http/1.1"},
		InsecureSkipVerify: true, // Audit even self-signed certs
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	if err != nil {
		audit.Grade = "F"
		audit.Vulnerabilities = append(audit.Vulnerabilities, fmt.Sprintf("TLS Handshake failed: %v", err))
		return audit
	}
	defer conn.Close()

	audit.LatencyMs = float64(time.Since(start).Microseconds()) / 1000.0
	state := conn.ConnectionState()
	audit.SelectedVersion = tlsVersionString(state.Version)
	audit.SelectedCipher = tls.CipherSuiteName(state.CipherSuite)
	audit.ALPNNegotiated = state.NegotiatedProtocol

	// Inspect full certificate chain
	for _, cert := range state.PeerCertificates {
		days := int(time.Until(cert.NotAfter).Hours() / 24)
		summary := CertificateSummary{
			Subject:       cert.Subject.CommonName,
			Issuer:        cert.Issuer.CommonName,
			NotBefore:     cert.NotBefore,
			NotAfter:      cert.NotAfter,
			DaysRemaining: days,
			SignatureAlgo: cert.SignatureAlgorithm.String(),
			PublicKeyType: fmt.Sprintf("%T", cert.PublicKey),
			SANs:          cert.DNSNames,
			IsCA:          cert.IsCA,
		}
		audit.CertChain = append(audit.CertChain, summary)

		if days <= 0 {
			audit.Vulnerabilities = append(audit.Vulnerabilities, fmt.Sprintf("Certificate %s is EXPIRED", cert.Subject.CommonName))
		} else if days < 14 {
			audit.Warnings = append(audit.Warnings, fmt.Sprintf("Certificate %s expires in %d days", cert.Subject.CommonName, days))
		}
		if strings.Contains(strings.ToLower(cert.SignatureAlgorithm.String()), "sha1") {
			audit.Vulnerabilities = append(audit.Vulnerabilities, "Uses insecure SHA-1 signature algorithm")
		}
	}

	// Calculate grade
	audit.Grade = computeGrade(audit)
	return audit
}

func testTLSVersion(ctx context.Context, addr, host string, version uint16, timeout time.Duration) bool {
	dialer := &net.Dialer{Timeout: timeout}
	cfg := &tls.Config{
		ServerName:         host,
		MinVersion:         version,
		MaxVersion:         version,
		InsecureSkipVerify: true,
	}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, cfg)
	if err == nil {
		_ = conn.Close()
		return true
	}
	return false
}

func computeGrade(a *TLSSecurityAudit) string {
	if len(a.Vulnerabilities) > 0 {
		for _, v := range a.Vulnerabilities {
			if strings.Contains(v, "EXPIRED") || strings.Contains(v, "failed") {
				return "F"
			}
		}
		return "C"
	}
	if len(a.Warnings) > 0 {
		return "B"
	}
	if a.SupportedProtocols["TLS 1.3"] && !a.SupportedProtocols["TLS 1.0"] && !a.SupportedProtocols["TLS 1.1"] {
		return "A+"
	}
	return "A"
}
