package cli

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var throughputDuration time.Duration

var throughputCmd = &cobra.Command{
	Use:   "throughput <url|host:port>",
	Short: "Measure data transfer throughput and download bandwidth",
	Long: `Measure sustained TCP streaming or HTTP download throughput.
Calculates total transferred bytes and bandwidth speed in MB/s and Mbps.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]
		dur := throughputDuration
		if dur <= 0 {
			dur = 3 * time.Second
		}

		ctx := context.Background()
		result := netprobe.ProbeThroughput(ctx, target, dur)
		return formatOutput(result)
	},
}

func init() {
	throughputCmd.Flags().DurationVarP(&throughputDuration, "duration", "d", 3*time.Second, "benchmark test duration")
}
