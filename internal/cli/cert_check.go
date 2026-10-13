package cli

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var certCheckCmd = &cobra.Command{
	Use:   "cert-check [host]",
	Short: "Audit SSL/TLS certificate lifetime, expiration countdown, and urgency",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		host := args[0]
		port, _ := cmd.Flags().GetInt("port")
		timeout, _ := cmd.Flags().GetDuration("timeout")

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		summary := netprobe.AuditCertExpiry(ctx, host, port, timeout)
		return formatOutput(summary)
	},
}

func init() {
	certCheckCmd.Flags().IntP("port", "p", 443, "SSL/TLS service port")
	certCheckCmd.Flags().Duration("timeout", 4*time.Second, "Handshake timeout")
}
