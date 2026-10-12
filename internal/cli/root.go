package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/config"
	"github.com/sujanv/pingforge/internal/output"
	"github.com/sujanv/pingforge/internal/storage"
)

var (
	cfgFile      string
	jsonOutput   bool
	yamlOutput   bool
	timeoutFlag  time.Duration
	concurrency  int
	verboseFlag  bool
	noColorFlag  bool

	globalConfig *config.Config
	globalStore  storage.Storage
)

// RootCmd is the primary CLI command for PingForge.
var RootCmd = &cobra.Command{
	Use:   "pingforge",
	Short: "PingForge - Advanced Network Diagnostics Toolkit",
	Long: `PingForge is an advanced, high-performance CLI network diagnostics toolkit.
Provides TCP connectivity checks, multi-resolver DNS lookups, latency and jitter measurement,
HTTP/HTTPS health probes, concurrent CIDR/batch scanning, and historical latency reports.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(cfgFile)
		if err != nil {
			return err
		}
		globalConfig = cfg

		if timeoutFlag > 0 {
			globalConfig.Timeout = timeoutFlag
		}
		if concurrency > 0 {
			globalConfig.Concurrency = concurrency
		}
		if jsonOutput {
			globalConfig.OutputFormat = "json"
		} else if yamlOutput {
			globalConfig.OutputFormat = "yaml"
		}
		if noColorFlag {
			globalConfig.NoColor = true
		}
		if verboseFlag {
			globalConfig.Verbose = true
		}

		if globalConfig.Storage.Enabled {
			store, err := storage.NewJSONStore(globalConfig.Storage.FilePath, globalConfig.Storage.MaxRecordsFile)
			if err == nil {
				globalStore = store
			}
		}

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if globalStore != nil {
			_ = globalStore.Close()
		}
	},
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "C", "", "config file path (default: ~/.pingforge.json)")
	RootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output results as indented JSON")
	RootCmd.PersistentFlags().BoolVar(&yamlOutput, "yaml", false, "output results as YAML")
	RootCmd.PersistentFlags().DurationVarP(&timeoutFlag, "timeout", "t", 0, "operation timeout (e.g. 5s, 500ms)")
	RootCmd.PersistentFlags().IntVar(&concurrency, "concurrency", 0, "concurrent worker count for probes (default: 10)")
	RootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "enable detailed verbose logs")
	RootCmd.PersistentFlags().BoolVar(&noColorFlag, "no-color", false, "disable ANSI color output")

	RootCmd.AddCommand(tcpCmd)
	RootCmd.AddCommand(dnsCmd)
	RootCmd.AddCommand(pingCmd)
	RootCmd.AddCommand(healthCmd)
	RootCmd.AddCommand(scanCmd)
	RootCmd.AddCommand(traceCmd)
	RootCmd.AddCommand(batchCmd)
	RootCmd.AddCommand(reportCmd)
	RootCmd.AddCommand(daemonCmd)
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(dohCmd)
	RootCmd.AddCommand(dotCmd)
	RootCmd.AddCommand(tlsAuditCmd)
	RootCmd.AddCommand(wsCmd)
	RootCmd.AddCommand(udpCmd)
	RootCmd.AddCommand(mtuCmd)
	RootCmd.AddCommand(throughputCmd)
	RootCmd.AddCommand(liveCmd)
	RootCmd.AddCommand(geoCmd)
	RootCmd.AddCommand(bgpCmd)
	RootCmd.AddCommand(arpCmd)
	RootCmd.AddCommand(http2Cmd)
	RootCmd.AddCommand(dnssecCmd)
	RootCmd.AddCommand(pcapCmd)
}

// Execute runs the root CLI command.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// formatOutput handles rendering based on global format flags.
func formatOutput(data interface{}) error {
	switch globalConfig.OutputFormat {
	case "json":
		return output.NewJSONFormatter(true).Format(os.Stdout, data)
	case "yaml":
		return output.NewYAMLFormatter().Format(os.Stdout, data)
	case "markdown":
		return output.NewMarkdownFormatter().Format(os.Stdout, data)
	case "ndjson":
		return output.NewNDJSONFormatter().Format(os.Stdout, data)
	default:
		return output.NewTableFormatter(globalConfig.NoColor).Format(os.Stdout, data)
	}
}
