package cli

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var dnssecCmd = &cobra.Command{
	Use:   "dnssec [domain]",
	Short: "Audit DNSSEC validation, cryptographic signatures, and chain of trust",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		domain := args[0]
		timeout, _ := cmd.Flags().GetDuration("timeout")

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		status := netprobe.ValidateDNSSEC(ctx, domain, timeout)
		return formatOutput(status)
	},
}

func init() {
	dnssecCmd.Flags().Duration("timeout", 5*time.Second, "DNSSEC lookup timeout")
}
