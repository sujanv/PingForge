package cli

import (
	"context"
	"net"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var udpPayload string

var udpCmd = &cobra.Command{
	Use:   "udp <host:port>",
	Short: "Send UDP diagnostic probe packet and test reachability",
	Long: `Send custom or default UDP probe packet to target host and port.
Measures round-trip response latency and checks port reachability.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		raw := args[0]
		host, portStr, err := net.SplitHostPort(raw)
		if err != nil {
			host = raw
			portStr = "53"
		}
		port, _ := strconv.Atoi(portStr)

		ctx, cancel := context.WithTimeout(context.Background(), globalConfig.Timeout)
		defer cancel()

		var payloadBytes []byte
		if udpPayload != "" {
			payloadBytes = []byte(udpPayload)
		}

		result := netprobe.ProbeUDP(ctx, host, port, payloadBytes, globalConfig.Timeout)
		return formatOutput(result)
	},
}

func init() {
	udpCmd.Flags().StringVarP(&udpPayload, "payload", "P", "", "custom payload string to send via UDP")
}
