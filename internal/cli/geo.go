package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
	"github.com/sujanv/pingforge/internal/output"
)

var geoCmd = &cobra.Command{
	Use:   "geo [ip]",
	Short: "Lookup IP Geolocation, Autonomous System (ASN), and ISP details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ip := args[0]
		timeout, _ := cmd.Flags().GetDuration("timeout")
		format, _ := cmd.Flags().GetString("output")

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		info, err := netprobe.LookupIPGeo(ctx, ip, timeout)
		if err != nil {
			return fmt.Errorf("geolocation lookup failed: %w", err)
		}

		if format == "json" {
			output.PrintJSON(info)
			return nil
		}

		tbl := output.NewTable("Property", "Value")
		tbl.AddRow("IP Address", info.IP)
		tbl.AddRow("Location", fmt.Sprintf("%s, %s, %s (%s)", info.City, info.Region, info.Country, info.CountryCode))
		tbl.AddRow("Coordinates", fmt.Sprintf("%.4f, %.4f", info.Latitude, info.Longitude))
		tbl.AddRow("ASN", info.ASN)
		tbl.AddRow("Organization / ISP", info.Org)
		tbl.AddRow("Timezone", info.Timezone)
		tbl.Render()

		return nil
	},
}

func init() {
	geoCmd.Flags().Duration("timeout", 4*time.Second, "Lookup timeout")
	RootCmd.AddCommand(geoCmd)
}
