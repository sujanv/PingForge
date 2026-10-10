package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var (
	dnsResolver    string
	dnsRecordTypes []string
	dnsPropagation bool
)

var dnsCmd = &cobra.Command{
	Use:   "dns <domain>",
	Short: "Query DNS resolution, verify records, and test propagation",
	Long: `Query DNS records across system and public resolvers (Cloudflare, Google, Quad9).
Inspects A, AAAA, CNAME, MX, TXT, NS, and SRV records with query latency diagnostics.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		domain := args[0]
		cfg := globalConfig.DNS

		if dnsResolver != "" {
			if !strings.Contains(dnsResolver, ":") {
				dnsResolver += ":53"
			}
			cfg.Resolvers = []string{dnsResolver}
		}
		if len(dnsRecordTypes) > 0 {
			cfg.RecordTypes = dnsRecordTypes
		}

		ctx, cancel := context.WithTimeout(context.Background(), globalConfig.Timeout)
		defer cancel()

		if dnsPropagation {
			propagation := netprobe.VerifyPropagation(ctx, domain, cfg.Timeout)
			return formatOutput(propagation)
		}

		result := netprobe.ProbeDNS(ctx, domain, cfg)

		if globalStore != nil {
			_ = globalStore.Save(context.Background(), result)
		}

		if err := formatOutput(result); err != nil {
			return err
		}

		if result.Status == "DOWN" {
			return fmt.Errorf("dns check failed: %s", result.Error)
		}
		return nil
	},
}

func init() {
	dnsCmd.Flags().StringVarP(&dnsResolver, "resolver", "r", "", "custom DNS resolver IP (e.g. 1.1.1.1:53)")
	dnsCmd.Flags().StringSliceVar(&dnsRecordTypes, "type", []string{"A", "AAAA", "CNAME", "MX", "TXT", "NS"}, "record types to query")
	dnsCmd.Flags().BoolVar(&dnsPropagation, "propagation", false, "verify DNS propagation across global public resolvers")
}
