package cli

import (
	"context"
	"net"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var tlsAuditCmd = &cobra.Command{
	Use:   "tls-audit <host[:port]>",
	Short: "Audit SSL/TLS configuration, protocol versions, ciphers, and security grade",
	Long: `Audit SSL/TLS server configurations. Evaluates support for TLS 1.0, 1.1, 1.2, 1.3,
inspects the certificate chain, extracts Subject Alternative Names (SANs),
audits ALPN protocol negotiation, and computes an automated security scorecard (A+ to F).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		raw := args[0]
		host := raw
		port := 443

		if h, pStr, err := net.SplitHostPort(raw); err == nil {
			host = h
			if p, err := strconv.Atoi(pStr); err == nil {
				port = p
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), globalConfig.Timeout)
		defer cancel()

		audit := netprobe.AuditTLS(ctx, host, port, globalConfig.Timeout)
		return formatOutput(audit)
	},
}
