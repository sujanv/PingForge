package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
	"github.com/sujanv/pingforge/internal/output"
)

var dnssecCmd = &cobra.Command{
	Use:   "dnssec [domain]",
	Short: "Audit DNSSEC validation, cryptographic signatures, and chain of trust",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		domain := args[0]
		timeout, _ := cmd.Flags().GetDuration("timeout")
		format, _ := cmd.Flags().GetString("output")

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		status := netprobe.ValidateDNSSEC(ctx, domain, timeout)

		if format == "json" {
			output.PrintJSON(status)
			return nil
		}

		tbl := output.NewTable("DNSSEC Property", "Status")
		tbl.AddRow("Domain", status.Domain)
		tbl.AddRow("Resolver", status.Resolver)
		tbl.AddRow("DNSSEC Validated", fmt.Sprintf("%t", status.Validated))
		tbl.AddRow("Authenticated Data (AD)", fmt.Sprintf("%t", status.Authenticated))
		tbl.AddRow("RRSIG Record Found", fmt.Sprintf("%t", status.RRSIGFound))
		if status.Error != "" {
			tbl.AddRow("Error", status.Error)
		}
		tbl.Render()

		return nil
	},
}

func init() {
	dnssecCmd.Flags().Duration("timeout", 5*time.Second, "DNSSEC lookup timeout")
	RootCmd.AddCommand(dnssecCmd)
}
