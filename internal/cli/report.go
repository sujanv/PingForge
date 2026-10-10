package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/sujanv/pingforge/internal/report"
)

var (
	reportTarget string
	reportSince  time.Duration
	reportFormat string
	reportOut    string
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate historical latency, SLA availability, and trend reports",
	Long: `Query historical probe records from persistent storage and generate
comprehensive latency SLA reports. Supports interactive HTML dashboards,
terminal ASCII tables with sparklines, CSV files, and structured JSON.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if globalStore == nil {
			return fmt.Errorf("historical storage is disabled or unavailable")
		}

		if reportSince <= 0 {
			reportSince = 24 * time.Hour
		}

		reporter := report.NewReporter(globalStore)
		ctx := context.Background()

		err := reporter.GenerateReport(ctx, reportTarget, reportSince, reportFormat, reportOut)
		if err != nil {
			return err
		}

		if reportOut != "" {
			fmt.Printf("Report successfully exported to: %s\n", reportOut)
		}
		return nil
	},
}

func init() {
	reportCmd.Flags().StringVar(&reportTarget, "target", "", "filter report for a specific host or target")
	reportCmd.Flags().DurationVar(&reportSince, "since", 24*time.Hour, "historical time window (e.g. 1h, 24h, 7d)")
	reportCmd.Flags().StringVarP(&reportFormat, "format", "f", "table", "output report format: table, html, csv, json")
	reportCmd.Flags().StringVarP(&reportOut, "out", "o", "", "file path to save report (e.g. report.html, report.csv)")
}
