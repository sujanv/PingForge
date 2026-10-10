package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var dotResolver string

var dotCmd = &cobra.Command{
	Use:   "dot <domain>",
	Short: "Query DNS-over-TLS (DoT) resolver on port 853",
	Long: `Query RFC 7858 DNS-over-TLS on port 853 with TLS handshake timing,
certificate verification, and wire protocol latency breakdown.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		domain := args[0]
		timeout := globalConfig.Timeout

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		result := netprobe.ProbeDoT(ctx, domain, dotResolver, timeout)
		if result.Error != "" {
			return fmt.Errorf("DoT probe error: %s", result.Error)
		}
		return formatOutput(result)
	},
}

func init() {
	dotCmd.Flags().StringVarP(&dotResolver, "resolver", "r", "1.1.1.1:853", "DoT resolver address (host:853)")
}
