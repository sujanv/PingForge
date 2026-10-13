package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var mdnsCmd = &cobra.Command{
	Use:   "mdns",
	Short: "Discover local LAN ZeroConf and Bonjour services via multicast DNS",
	RunE: func(cmd *cobra.Command, args []string) error {
		timeout, _ := cmd.Flags().GetDuration("timeout")

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		services, err := netprobe.ScanMDNS(ctx, timeout)
		if err != nil {
			return fmt.Errorf("mdns scan failed: %w", err)
		}

		return formatOutput(services)
	},
}

func init() {
	mdnsCmd.Flags().Duration("timeout", 2*time.Second, "Discovery listen window")
}
