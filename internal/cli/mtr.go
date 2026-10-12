package cli

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var mtrCmd = &cobra.Command{
	Use:   "mtr [host]",
	Short: "Continuous traceroute diagnostics with per-hop packet loss and jitter telemetry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		host := args[0]
		cycles, _ := cmd.Flags().GetInt("cycles")
		port, _ := cmd.Flags().GetInt("port")
		timeout, _ := cmd.Flags().GetDuration("timeout")

		ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Duration(cycles))
		defer cancel()

		report := netprobe.RunMTRDiagnostics(ctx, host, port, cycles, timeout)
		return formatOutput(report)
	},
}

func init() {
	mtrCmd.Flags().IntP("cycles", "c", 3, "Number of traceroute cycles to execute")
	mtrCmd.Flags().IntP("port", "p", 80, "Destination port for trace probes")
	mtrCmd.Flags().Duration("timeout", 2*time.Second, "Per-probe timeout")
}
