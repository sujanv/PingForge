package report

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/sujanv/pingforge/internal/core"
)

// ExportSummariesCSV writes historical summaries as comma-separated values.
func ExportSummariesCSV(w io.Writer, summaries []core.HistorySummary) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Header
	header := []string{"target", "total_probes", "successful_probes", "failed_probes", "uptime_percentage", "min_latency_ms", "avg_latency_ms", "max_latency_ms", "p95_latency_ms", "first_seen", "last_seen"}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, s := range summaries {
		row := []string{
			s.Target,
			fmt.Sprintf("%d", s.TotalProbes),
			fmt.Sprintf("%d", s.SuccessfulProbes),
			fmt.Sprintf("%d", s.FailedProbes),
			fmt.Sprintf("%.2f", s.UptimePercentage),
			fmt.Sprintf("%.2f", s.MinLatencyMs),
			fmt.Sprintf("%.2f", s.AvgLatencyMs),
			fmt.Sprintf("%.2f", s.MaxLatencyMs),
			fmt.Sprintf("%.2f", s.P95LatencyMs),
			s.FirstSeen.Format("2006-01-02 15:04:05"),
			s.LastSeen.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// ExportProbesCSV writes raw probe records as comma-separated values.
func ExportProbesCSV(w io.Writer, records []core.ProbeResult) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{"timestamp", "target", "probe_type", "status", "latency_ms", "message", "error"}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, r := range records {
		row := []string{
			r.Timestamp.Format("2006-01-02 15:04:05"),
			r.Target.String(),
			string(r.Type),
			string(r.Status),
			fmt.Sprintf("%.2f", r.LatencyMs),
			r.Message,
			r.Error,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}
