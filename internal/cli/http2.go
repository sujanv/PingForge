package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
	"github.com/sujanv/pingforge/internal/output"
)

var http2Cmd = &cobra.Command{
	Use:   "http2 [url]",
	Short: "Probe HTTP/2 multiplexing, ALPN negotiation, and stream headers",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetURL := args[0]
		timeout, _ := cmd.Flags().GetDuration("timeout")
		format, _ := cmd.Flags().GetString("output")

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		res := netprobe.ProbeHTTP2(ctx, targetURL, timeout)

		if format == "json" {
			output.PrintJSON(res)
			return nil
		}

		tbl := output.NewTable("HTTP/2 Metric", "Result")
		tbl.AddRow("Target URL", res.URL)
		tbl.AddRow("Protocol Version", res.ProtocolVersion)
		tbl.AddRow("Negotiated ALPN", res.NegotiatedALPN)
		tbl.AddRow("Multiplex Capable", fmt.Sprintf("%t", res.MultiplexCapable))
		tbl.AddRow("Status Code", fmt.Sprintf("%d", res.StatusCode))
		tbl.AddRow("Round Trip Time", fmt.Sprintf("%.2f ms", res.TotalDurationMs))
		if res.Error != "" {
			tbl.AddRow("Error", res.Error)
		}
		tbl.Render()

		return nil
	},
}

func init() {
	http2Cmd.Flags().Duration("timeout", 5*time.Second, "HTTP/2 connection timeout")
	RootCmd.AddCommand(http2Cmd)
}
