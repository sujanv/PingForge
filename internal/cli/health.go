package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
)

var (
	healthExpectedStatus []int
	healthBodyRegex      string
	healthBudget         time.Duration
	healthNoRedirect     bool
	healthHeaders        []string
)

var healthCmd = &cobra.Command{
	Use:   "health <url>",
	Short: "Deep HTTP/HTTPS endpoint health check and timing breakdown",
	Long: `Perform an in-depth HTTP/HTTPS health probe with granular timing phases.
Breaks down DNS resolution, TCP connection, TLS handshake, and Time to First Byte (TTFB).
Validates expected status codes, regular expressions on response bodies, and latency budgets.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rawURL := args[0]
		if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
			rawURL = "https://" + rawURL
		}

		headersMap := make(map[string]string)
		for _, h := range healthHeaders {
			parts := strings.SplitN(h, ":", 2)
			if len(parts) == 2 {
				headersMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}

		opts := netprobe.HealthCheckOptions{
			ExpectedStatus:  globalConfig.Health.ExpectedStatus,
			BodyRegex:       healthBodyRegex,
			FollowRedirects: !healthNoRedirect,
			MaxRedirects:    globalConfig.Health.MaxRedirects,
			Headers:         headersMap,
			UserAgent:       globalConfig.Health.UserAgent,
			LatencyBudget:   healthBudget,
			Timeout:         globalConfig.Timeout,
		}

		if len(healthExpectedStatus) > 0 {
			opts.ExpectedStatus = healthExpectedStatus
		}
		if healthBudget <= 0 {
			opts.LatencyBudget = globalConfig.Health.LatencyBudget
		}

		ctx, cancel := context.WithTimeout(context.Background(), globalConfig.Timeout)
		defer cancel()

		result := netprobe.ProbeHealth(ctx, rawURL, opts)

		if globalStore != nil {
			_ = globalStore.Save(context.Background(), result)
		}

		if err := formatOutput(result); err != nil {
			return err
		}

		if result.Status == "DOWN" {
			return fmt.Errorf("health check failed: %s", result.Message)
		}
		return nil
	},
}

func init() {
	healthCmd.Flags().IntSliceVar(&healthExpectedStatus, "status", nil, "acceptable HTTP status codes (e.g. 200,204)")
	healthCmd.Flags().StringVar(&healthBodyRegex, "match", "", "regex pattern that response body must contain")
	healthCmd.Flags().DurationVar(&healthBudget, "budget", 0, "maximum acceptable response latency budget (e.g. 500ms)")
	healthCmd.Flags().BoolVar(&healthNoRedirect, "no-redirect", false, "do not follow HTTP redirects")
	healthCmd.Flags().StringSliceVarP(&healthHeaders, "header", "H", nil, "custom request headers in Key:Value format")
}
