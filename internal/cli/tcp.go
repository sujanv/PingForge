package cli

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var (
	tcpTLS        bool
	tcpBanner     bool
	tcpKeepAlive  time.Duration
	tcpNoDelay    bool
)

var tcpCmd = &cobra.Command{
	Use:   "tcp <host[:port]>",
	Short: "Check TCP connectivity, latency, and socket diagnostics",
	Long: `Probe TCP connectivity to a target host and port.
Measures handshake latency with microsecond precision, inspects TLS certificates,
tests socket keep-alive, and grabs server service banners.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		raw := args[0]
		host := raw
		port := 80

		if h, pStr, err := net.SplitHostPort(raw); err == nil {
			host = h
			if p, err := strconv.Atoi(pStr); err == nil {
				port = p
			}
		}

		cfg := globalConfig.TCP
		if tcpTLS {
			cfg.TLS = true
		}
		if tcpBanner {
			cfg.ReadBanner = true
		}
		if tcpKeepAlive > 0 {
			cfg.KeepAlive = tcpKeepAlive
		}
		if cmd.Flags().Changed("nodelay") {
			cfg.NoDelay = tcpNoDelay
		}

		timeout := globalConfig.Timeout
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		result := netprobe.ProbeTCP(ctx, host, port, cfg, timeout)

		if globalStore != nil {
			_ = globalStore.Save(context.Background(), result)
		}

		if err := formatOutput(result); err != nil {
			return err
		}

		if result.Status == "DOWN" {
			return fmt.Errorf("tcp check failed: %s", result.Error)
		}
		return nil
	},
}

func init() {
	tcpCmd.Flags().BoolVar(&tcpTLS, "tls", false, "perform TLS handshake and inspect certificates")
	tcpCmd.Flags().BoolVar(&tcpBanner, "banner", false, "grab initial service banner from connection")
	tcpCmd.Flags().DurationVar(&tcpKeepAlive, "keepalive", 30*time.Second, "TCP keep-alive probe interval")
	tcpCmd.Flags().BoolVar(&tcpNoDelay, "nodelay", true, "enable TCP_NODELAY socket option")
}
