package cli

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var ocspCmd = &cobra.Command{
	Use:   "ocsp [host]",
	Short: "Audit SSL/TLS OCSP Stapling (RFC 6066) and certificate revocation status",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		host := args[0]
		port, _ := cmd.Flags().GetInt("port")
		timeout, _ := cmd.Flags().GetDuration("timeout")

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		res := netprobe.InspectOCSPStapling(ctx, host, port, timeout)
		return formatOutput(res)
	},
}

func init() {
	ocspCmd.Flags().IntP("port", "p", 443, "TLS server port")
	ocspCmd.Flags().Duration("timeout", 4*time.Second, "OCSP probe timeout")
}
