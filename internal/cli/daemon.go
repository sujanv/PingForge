package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/core"
	"github.com/sujanv/pingforge/internal/daemon"
	"github.com/sujanv/pingforge/internal/engine"
)

var (
	daemonListen   string
	daemonInterval time.Duration
	daemonFile     string
)

var daemonCmd = &cobra.Command{
	Use:   "daemon [target1 target2 ...]",
	Short: "Run continuous network monitoring daemon with Prometheus /metrics",
	Long: `Run PingForge in background daemon mode. Continuously monitors targets,
persists metrics to historical storage, and serves Prometheus metrics on /metrics.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var targets []core.Target

		for _, arg := range args {
			parsed, err := engine.ParseTargetString(arg)
			if err != nil {
				return err
			}
			targets = append(targets, parsed...)
		}

		if daemonFile != "" {
			fileTargets, err := engine.ReadTargetsFile(daemonFile)
			if err != nil {
				return fmt.Errorf("read target file: %w", err)
			}
			targets = append(targets, fileTargets...)
		}

		if len(targets) == 0 {
			// Default targets if none provided
			targets = []core.Target{
				{Host: "1.1.1.1", Port: 53, Protocol: "tcp"},
				{Host: "8.8.8.8", Port: 53, Protocol: "tcp"},
				{Host: "google.com", Port: 443, Protocol: "tcp"},
			}
		}

		server := daemon.NewServer(*globalConfig, targets, globalStore)

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		return server.Start(ctx, daemonListen, daemonInterval)
	},
}

func init() {
	daemonCmd.Flags().StringVarP(&daemonListen, "listen", "l", ":9090", "HTTP listen address for Prometheus /metrics and healthz")
	daemonCmd.Flags().DurationVarP(&daemonInterval, "interval", "i", 15*time.Second, "polling interval across target probes")
	daemonCmd.Flags().StringVarP(&daemonFile, "file", "f", "", "path to target file list")
}
