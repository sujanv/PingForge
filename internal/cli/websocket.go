package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var wsProtocols []string

var wsCmd = &cobra.Command{
	Use:   "ws <url>",
	Short: "Probe WebSocket connection, upgrade handshake, and ping-pong RTT",
	Long: `Validate RFC 6455 WebSocket connectivity over ws:// or wss://.
Measures HTTP 101 Switching Protocols upgrade latency, validates Sec-WebSocket-Accept,
and measures round-trip time (RTT) for WebSocket Ping/Pong control frames.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rawURL := args[0]
		if !strings.HasPrefix(rawURL, "ws://") && !strings.HasPrefix(rawURL, "wss://") {
			rawURL = "wss://" + rawURL
		}

		ctx, cancel := context.WithTimeout(context.Background(), globalConfig.Timeout)
		defer cancel()

		result := netprobe.ProbeWebSocket(ctx, rawURL, wsProtocols, globalConfig.Timeout)
		if result.Error != "" {
			return fmt.Errorf("WebSocket error: %s", result.Error)
		}
		return formatOutput(result)
	},
}

func init() {
	wsCmd.Flags().StringSliceVar(&wsProtocols, "subprotocol", nil, "requested WebSocket subprotocols")
}
