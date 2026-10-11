package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
	"github.com/sujanv/pingforge/internal/output"
)

var arpCmd = &cobra.Command{
	Use:   "arp",
	Short: "Scan and display local ARP / Neighbor cache entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		format, _ := cmd.Flags().GetString("output")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		entries, err := netprobe.ScanARPTable(ctx)
		if err != nil {
			return fmt.Errorf("arp scan failed: %w", err)
		}

		if format == "json" {
			output.PrintJSON(entries)
			return nil
		}

		tbl := output.NewTable("IP Address", "MAC Address", "Interface", "Type")
		for _, e := range entries {
			tbl.AddRow(e.IP, e.MAC, e.Interface, e.Type)
		}
		tbl.Render()

		return nil
	},
}

func init() {
	RootCmd.AddCommand(arpCmd)
}
