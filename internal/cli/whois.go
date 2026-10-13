package cli

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var whoisCmd = &cobra.Command{
	Use:   "whois [domain-or-ip]",
	Short: "Perform RFC 3912 WHOIS and RDAP domain / IP registration queries",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]
		timeout, _ := cmd.Flags().GetDuration("timeout")

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		res := netprobe.QueryWHOIS(ctx, target, timeout)
		return formatOutput(res)
	},
}

func init() {
	whoisCmd.Flags().Duration("timeout", 5*time.Second, "Query timeout")
}
