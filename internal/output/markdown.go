package output

import (
	"fmt"
	"io"

	"github.com/sujanv/pingforge/internal/core"
)

// MarkdownFormatter outputs GitHub Flavored Markdown tables.
type MarkdownFormatter struct{}

// NewMarkdownFormatter creates a new MarkdownFormatter.
func NewMarkdownFormatter() *MarkdownFormatter {
	return &MarkdownFormatter{}
}

// Format renders data as GFM table.
func (m *MarkdownFormatter) Format(w io.Writer, data interface{}) error {
	switch v := data.(type) {
	case []core.ProbeResult:
		fmt.Fprintln(w, "| Target | Type | Status | Latency (ms) | Message |")
		fmt.Fprintln(w, "| --- | --- | --- | --- | --- |")
		for _, p := range v {
			fmt.Fprintf(w, "| `%s` | %s | **%s** | %.2f | %s |\n",
				p.Target.String(), p.Type, p.Status, p.LatencyMs, p.Message)
		}
	case []core.HistorySummary:
		fmt.Fprintln(w, "| Target | Probes | Uptime SLA | Min (ms) | Avg (ms) | Max (ms) | P95 (ms) | Trend |")
		fmt.Fprintln(w, "| --- | --- | --- | --- | --- | --- | --- | --- |")
		for _, s := range v {
			fmt.Fprintf(w, "| `%s` | %d | %.2f%% | %.2f | %.2f | %.2f | %.2f | `%s` |\n",
				s.Target, s.TotalProbes, s.UptimePercentage, s.MinLatencyMs, s.AvgLatencyMs, s.MaxLatencyMs, s.P95LatencyMs, s.Sparkline)
		}
	default:
		fmt.Fprintf(w, "```json\n%+v\n```\n", data)
	}
	return nil
}
