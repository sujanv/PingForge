package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var (
	dohProvider   string
	dohRecordType string
)

var dohCmd = &cobra.Command{
	Use:   "doh <domain>",
	Short: "Query DNS-over-HTTPS (DoH) encrypted endpoints",
	Long: `Execute DNS-over-HTTPS queries against RFC 8484 and JSON DoH endpoints.
Supported public providers: cloudflare, google, quad9, adguard, or custom HTTPS URL.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		domain := args[0]
		timeout := globalConfig.Timeout

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		result := netprobe.ProbeDoH(ctx, domain, dohProvider, dohRecordType, timeout)
		if result.Error != "" {
			return fmt.Errorf("DoH probe error: %s", result.Error)
		}
		return formatOutput(result)
	},
}

func init() {
	dohCmd.Flags().StringVarP(&dohProvider, "provider", "p", "cloudflare", "DoH provider name (cloudflare, google, quad9, adguard) or custom URL")
	dohCmd.Flags().StringVarP(&dohRecordType, "type", "T", "A", "DNS record type (A, AAAA, MX, TXT)")
}
