package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var arpCmd = &cobra.Command{
	Use:   "arp",
	Short: "Scan and display local ARP / Neighbor cache entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		entries, err := netprobe.ScanARPTable(ctx)
		if err != nil {
			return fmt.Errorf("arp scan failed: %w", err)
		}

		return formatOutput(entries)
	},
}
