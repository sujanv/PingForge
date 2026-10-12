package cli

import (
	"context"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var meshCmd = &cobra.Command{
	Use:   "mesh [host1,host2,host3...]",
	Short: "Run full-mesh latency probes across multi-node peer networks",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hosts := strings.Split(args[0], ",")
		port, _ := cmd.Flags().GetInt("port")
		timeout, _ := cmd.Flags().GetDuration("timeout")

		ctx, cancel := context.WithTimeout(context.Background(), timeout*2)
		defer cancel()

		res := netprobe.BuildSubnetMesh(ctx, hosts, port, timeout)
		return formatOutput(res)
	},
}

func init() {
	meshCmd.Flags().IntP("port", "p", 80, "Destination port for mesh probes")
	meshCmd.Flags().Duration("timeout", 2*time.Second, "Probe timeout per peer")
}
