package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var geoCmd = &cobra.Command{
	Use:   "geo [ip]",
	Short: "Lookup IP Geolocation, Autonomous System (ASN), and ISP details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ip := args[0]
		timeout, _ := cmd.Flags().GetDuration("timeout")

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		info, err := netprobe.LookupIPGeo(ctx, ip, timeout)
		if err != nil {
			return fmt.Errorf("geolocation lookup failed: %w", err)
		}

		return formatOutput(info)
	},
}

func init() {
	geoCmd.Flags().Duration("timeout", 4*time.Second, "Lookup timeout")
}
