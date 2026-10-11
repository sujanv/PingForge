package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
	"github.com/sujanv/pingforge/internal/output"
)

var bgpCmd = &cobra.Command{
	Use:   "bgp [target-ip-or-asn]",
	Short: "Audit BGP route visibility, origin ASN, and upstream transit providers",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]
		timeout, _ := cmd.Flags().GetDuration("timeout")
		format, _ := cmd.Flags().GetString("output")

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		route, err := netprobe.QueryBGPRoute(ctx, target, timeout)
		if err != nil {
			return fmt.Errorf("bgp route query failed: %w", err)
		}

		if format == "json" {
			output.PrintJSON(route)
			return nil
		}

		tbl := output.NewTable("BGP Attribute", "Details")
		tbl.AddRow("Target Resource", route.Resource)
		tbl.AddRow("Prefix", route.Prefix)
		tbl.AddRow("Origin ASN", route.ASN)
		tbl.AddRow("Route Visibility", fmt.Sprintf("%.2f%%", route.Visibility))
		tbl.AddRow("Sample Upstreams", fmt.Sprintf("%v", route.Upstreams))
		tbl.Render()

		return nil
	},
}

func init() {
	bgpCmd.Flags().Duration("timeout", 5*time.Second, "BGP lookup timeout")
	RootCmd.AddCommand(bgpCmd)
}
