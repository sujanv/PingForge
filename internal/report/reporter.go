package report

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sujanv/pingforge/internal/core"
	"github.com/sujanv/pingforge/internal/output"
	"github.com/sujanv/pingforge/internal/storage"
)

// Reporter aggregates historical metrics and generates reports across formats.
type Reporter struct {
	store storage.Storage
}

// NewReporter creates a Reporter instance with the given storage backend.
func NewReporter(store storage.Storage) *Reporter {
	return &Reporter{store: store}
}

// GenerateReport produces a formatted report for targets within since duration.
func (r *Reporter) GenerateReport(ctx context.Context, target string, since time.Duration, format string, outFile string) error {
	summaries, err := r.store.GetSummaries(ctx, since)
	if err != nil {
		return fmt.Errorf("retrieve summaries: %w", err)
	}

	if target != "" {
		filtered := make([]core.HistorySummary, 0)
		for _, s := range summaries {
			if s.Target == target {
				filtered = append(filtered, s)
			}
		}
		summaries = filtered
	}

	// Enrich summaries with ASCII sparklines
	for i := range summaries {
		history, _ := r.store.GetHistory(ctx, summaries[i].Target, since)
		var latencies []float64
		for _, h := range history {
			if h.LatencyMs > 0 {
				latencies = append(latencies, h.LatencyMs)
			}
		}
		summaries[i].Sparkline = GenerateSparkline(latencies)
	}

	switch format {
	case "html":
		htmlContent := GenerateHTMLReport(summaries, "PingForge - Latency & Availability Report")
		if outFile != "" {
			return os.WriteFile(outFile, []byte(htmlContent), 0644)
		}
		fmt.Println(htmlContent)
		return nil

	case "csv":
		if outFile != "" {
			f, err := os.Create(outFile)
			if err != nil {
				return err
			}
			defer f.Close()
			return ExportSummariesCSV(f, summaries)
		}
		return ExportSummariesCSV(os.Stdout, summaries)

	case "json":
		formatter := output.NewJSONFormatter(true)
		if outFile != "" {
			f, err := os.Create(outFile)
			if err != nil {
				return err
			}
			defer f.Close()
			return formatter.Format(f, summaries)
		}
		return formatter.Format(os.Stdout, summaries)

	default: // table
		formatter := output.NewTableFormatter(false)
		return formatter.Format(os.Stdout, summaries)
	}
}
