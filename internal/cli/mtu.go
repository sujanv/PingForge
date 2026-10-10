package cli

import (
	"context"
	"net"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var mtuPort int

var mtuCmd = &cobra.Command{
	Use:   "mtu <host>",
	Short: "Detect Path MTU threshold and socket payload limits",
	Long: `Probe host with incrementally sized packet payloads (576 to 9000 bytes)
to identify maximum unfragmented transmission unit and estimated TCP MSS.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		raw := args[0]
		host := raw
		port := mtuPort

		if h, pStr, err := net.SplitHostPort(raw); err == nil {
			host = h
			if p, err := strconv.Atoi(pStr); err == nil {
				port = p
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), globalConfig.Timeout)
		defer cancel()

		result := netprobe.ProbeMTU(ctx, host, port, globalConfig.Timeout)
		return formatOutput(result)
	},
}

func init() {
	mtuCmd.Flags().IntVarP(&mtuPort, "port", "p", 80, "target destination TCP port for MTU probing")
}
