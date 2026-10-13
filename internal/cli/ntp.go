package cli

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var ntpCmd = &cobra.Command{
	Use:   "ntp [server]",
	Short: "Query RFC 5905 Network Time Protocol server, clock stratum, and offset",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		server := "pool.ntp.org"
		if len(args) > 0 {
			server = args[0]
		}

		port, _ := cmd.Flags().GetInt("port")
		timeout, _ := cmd.Flags().GetDuration("timeout")

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		res := netprobe.ProbeNTP(ctx, server, port, timeout)
		return formatOutput(res)
	},
}

func init() {
	ntpCmd.Flags().IntP("port", "p", 123, "NTP UDP port")
	ntpCmd.Flags().Duration("timeout", 4*time.Second, "NTP query timeout")
}
