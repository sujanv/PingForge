package cli

import (
	"context"
	"net"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var (
	traceMaxHops int
	tracePort    int
)

var traceCmd = &cobra.Command{
	Use:   "trace <host[:port]>",
	Short: "Inspect route latency and network hop reachability",
	Long: `Trace network route to host, identifying intermediate router addresses
and latency measurements.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		raw := args[0]
		host := raw
		port := tracePort

		if h, pStr, err := net.SplitHostPort(raw); err == nil {
			host = h
			if p, err := strconv.Atoi(pStr); err == nil {
				port = p
			}
		}

		cfg := netprobe.TraceConfig{
			MaxHops: traceMaxHops,
			Timeout: globalConfig.Timeout,
		}

		ctx, cancel := context.WithTimeout(context.Background(), globalConfig.Timeout)
		defer cancel()

		result := netprobe.TraceRoute(ctx, host, port, cfg)
		return formatOutput(result)
	},
}

func init() {
	traceCmd.Flags().IntVar(&traceMaxHops, "max-hops", 15, "maximum number of route hops to query")
	traceCmd.Flags().IntVarP(&tracePort, "port", "p", 80, "target destination TCP port")
}
