package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/netprobe"
	"github.com/sujanv/pingforge/internal/output"
)

var (
	scanPortsFlag   string
	scanConcurrency int
	scanBanner      bool
)

var scanCmd = &cobra.Command{
	Use:   "scan <host>",
	Short: "Fast concurrent TCP port scan with service banner detection",
	Long: `Scan target host ports concurrently to detect open services.
Supports port ranges (1-1024), comma-separated lists (80,443,8080),
and grabs service greeting banners (SSH, HTTP, FTP, Redis).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		host := args[0]

		ports, err := parsePorts(scanPortsFlag)
		if err != nil {
			return err
		}

		concurrencyVal := scanConcurrency
		if concurrencyVal <= 0 {
			concurrencyVal = globalConfig.Concurrency
			if concurrencyVal <= 0 {
				concurrencyVal = 20
			}
		}

		cfg := netprobe.ScanConfig{
			Ports:       ports,
			Concurrency: concurrencyVal,
			Timeout:     globalConfig.Timeout,
			BannerGrab:  scanBanner,
		}

		sp := output.NewSpinner(fmt.Sprintf("Scanning %d ports on %s...", len(cfg.Ports), host))
		if !jsonOutput && !yamlOutput {
			sp.Start()
		}

		ctx := context.Background()
		result := netprobe.ScanPorts(ctx, host, cfg)
		sp.Stop()

		return formatOutput(result)
	},
}

func init() {
	scanCmd.Flags().StringVarP(&scanPortsFlag, "ports", "p", "", "ports to scan (e.g. '80,443', '20-100', or default common ports)")
	scanCmd.Flags().IntVar(&scanConcurrency, "concurrency", 20, "concurrent worker count")
	scanCmd.Flags().BoolVar(&scanBanner, "banner", true, "grab service banners for open ports")
}

func parsePorts(s string) ([]int, error) {
	if s == "" {
		return netprobe.DefaultScanPorts(), nil
	}

	var ports []int
	parts := strings.Split(s, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid port range %q", part)
			}
			low, err1 := strconv.Atoi(rangeParts[0])
			high, err2 := strconv.Atoi(rangeParts[1])
			if err1 != nil || err2 != nil || low > high || low < 1 || high > 65535 {
				return nil, fmt.Errorf("invalid port range bounds: %s", part)
			}
			for p := low; p <= high; p++ {
				ports = append(ports, p)
			}
		} else {
			p, err := strconv.Atoi(part)
			if err != nil || p < 1 || p > 65535 {
				return nil, fmt.Errorf("invalid port %q", part)
			}
			ports = append(ports, p)
		}
	}
	return ports, nil
}
