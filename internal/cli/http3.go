package cli

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var http3Cmd = &cobra.Command{
	Use:   "http3 [url]",
	Short: "Audit HTTP/3 (QUIC) capability, Alt-Svc headers, and handshake RTT",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetURL := args[0]
		timeout, _ := cmd.Flags().GetDuration("timeout")

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		res := netprobe.ProbeHTTP3(ctx, targetURL, timeout)
		return formatOutput(res)
	},
}

func init() {
	http3Cmd.Flags().Duration("timeout", 4*time.Second, "HTTP/3 discovery timeout")
}
