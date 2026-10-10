package cli

import (
	"context"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/output"
)

var (
	liveInterval time.Duration
	livePort     int
)

var liveCmd = &cobra.Command{
	Use:   "live <host[:port]>",
	Short: "Live interactive terminal dashboard with real-time latency charts",
	Long: `Launch full-screen streaming terminal dashboard for target.
Displays live packets sent/received, packet loss gauge, rolling jitter,
and real-time ASCII latency sparkline graphs.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		raw := args[0]
		host := raw
		port := livePort

		if h, pStr, err := net.SplitHostPort(raw); err == nil {
			host = h
			if p, err := strconv.Atoi(pStr); err == nil {
				port = p
			}
		}

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		mon := output.NewLiveMonitor()
		return mon.Run(ctx, host, port, liveInterval)
	},
}

func init() {
	liveCmd.Flags().DurationVarP(&liveInterval, "interval", "i", 500*time.Millisecond, "refresh polling interval")
	liveCmd.Flags().IntVarP(&livePort, "port", "p", 80, "target port")
}
