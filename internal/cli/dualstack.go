package cli

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var dualstackCmd = &cobra.Command{
	Use:   "dualstack [domain]",
	Short: "Audit dual-stack IPv4 (A) vs IPv6 (AAAA) resolution and Happy Eyeballs race",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		domain := args[0]
		port, _ := cmd.Flags().GetInt("port")
		timeout, _ := cmd.Flags().GetDuration("timeout")

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		res := netprobe.ProbeDualStack(ctx, domain, port, timeout)
		return formatOutput(res)
	},
}

func init() {
	dualstackCmd.Flags().IntP("port", "p", 80, "Destination connection port")
	dualstackCmd.Flags().Duration("timeout", 4*time.Second, "Connection race timeout")
}
