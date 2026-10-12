package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var bgpCmd = &cobra.Command{
	Use:   "bgp [target-ip-or-asn]",
	Short: "Audit BGP route visibility, origin ASN, and upstream transit providers",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]
		timeout, _ := cmd.Flags().GetDuration("timeout")

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		route, err := netprobe.QueryBGPRoute(ctx, target, timeout)
		if err != nil {
			return fmt.Errorf("bgp route query failed: %w", err)
		}

		return formatOutput(route)
	},
}

func init() {
	bgpCmd.Flags().Duration("timeout", 5*time.Second, "BGP lookup timeout")
}
