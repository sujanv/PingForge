package cli

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var snmpCmd = &cobra.Command{
	Use:   "snmp [host]",
	Short: "Probe network switch and router MIB OIDs via SNMP v2c",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		host := args[0]
		port, _ := cmd.Flags().GetInt("port")
		community, _ := cmd.Flags().GetString("community")
		oid, _ := cmd.Flags().GetString("oid")
		timeout, _ := cmd.Flags().GetDuration("timeout")

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		res := netprobe.ProbeSNMP(ctx, host, port, community, oid, timeout)
		return formatOutput(res)
	},
}

func init() {
	snmpCmd.Flags().IntP("port", "p", 161, "SNMP UDP port")
	snmpCmd.Flags().StringP("community", "c", "public", "SNMP community string")
	snmpCmd.Flags().StringP("oid", "o", "1.3.6.1.2.1.1.1.0", "Target MIB-II Object Identifier (OID)")
	snmpCmd.Flags().Bool("contact", false, "Query sysContact (1.3.6.1.2.1.1.4.0)")
	snmpCmd.Flags().Bool("location", false, "Query sysLocation (1.3.6.1.2.1.1.6.0)")
	snmpCmd.Flags().Duration("timeout", 3*time.Second, "SNMP response timeout")
}
