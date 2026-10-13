package cli

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var bufferbloatCmd = &cobra.Command{
	Use:   "bufferbloat [url-or-host]",
	Short: "Profile network queue delay and bufferbloat grade under traffic saturation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]
		duration, _ := cmd.Flags().GetDuration("duration")

		rawURL := target
		host := target
		port := 80

		if u, err := url.Parse(target); err == nil && u.Host != "" {
			host = u.Hostname()
			if u.Scheme == "https" {
				port = 443
			}
		} else {
			rawURL = fmt.Sprintf("http://%s", target)
		}

		ctx, cancel := context.WithTimeout(context.Background(), duration+2*time.Second)
		defer cancel()

		res := netprobe.ProbeBufferbloat(ctx, rawURL, host, port, duration)
		return formatOutput(res)
	},
}

func init() {
	bufferbloatCmd.Flags().DurationP("duration", "d", 3*time.Second, "Test duration under active traffic load")
}
