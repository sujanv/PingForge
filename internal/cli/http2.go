package cli

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var http2Cmd = &cobra.Command{
	Use:   "http2 [url]",
	Short: "Probe HTTP/2 multiplexing, ALPN negotiation, and stream headers",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetURL := args[0]
		timeout, _ := cmd.Flags().GetDuration("timeout")

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		res := netprobe.ProbeHTTP2(ctx, targetURL, timeout)
		return formatOutput(res)
	},
}

func init() {
	http2Cmd.Flags().Duration("timeout", 5*time.Second, "HTTP/2 connection timeout")
}
