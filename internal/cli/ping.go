package cli

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
	"github.com/sujanv/pingforge/internal/output"
)

var (
	pingCount    int
	pingInterval time.Duration
	pingPort     int
)

var pingCmd = &cobra.Command{
	Use:   "ping <host[:port]>",
	Short: "Measure round-trip latency, RFC 3550 jitter, and percentiles",
	Long: `Execute rapid consecutive connection probes to evaluate round-trip latency (RTT).
Computes sample variance, standard deviation, RFC 3550 interarrival jitter,
packet loss rates, and latency percentiles (P50, P90, P95, P99).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		raw := args[0]
		host := raw
		port := pingPort

		if h, pStr, err := net.SplitHostPort(raw); err == nil {
			host = h
			if p, err := strconv.Atoi(pStr); err == nil {
				port = p
			}
		}

		cfg := globalConfig.Ping
		if pingCount > 0 {
			cfg.Count = pingCount
		}
		if pingInterval > 0 {
			cfg.Interval = pingInterval
		}
		if port > 0 {
			cfg.Port = port
		}

		sp := output.NewSpinner(fmt.Sprintf("Pinging %s:%d (%d probes)...", host, cfg.Port, cfg.Count))
		if !jsonOutput && !yamlOutput {
			sp.Start()
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Count)*time.Second+5*time.Second)
		defer cancel()

		stats := netprobe.ProbeLatency(ctx, host, cfg.Port, cfg, globalConfig.Timeout)
		sp.Stop()

		return formatOutput(stats)
	},
}

func init() {
	pingCmd.Flags().IntVarP(&pingCount, "count", "c", 5, "number of probe packets to send")
	pingCmd.Flags().DurationVarP(&pingInterval, "interval", "i", 200*time.Millisecond, "interval between probe packets")
	pingCmd.Flags().IntVarP(&pingPort, "port", "p", 80, "target TCP port for latency probing")
}
