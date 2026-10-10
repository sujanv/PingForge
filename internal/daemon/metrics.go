package daemon

import (
	"fmt"
	"io"
	"strings"

	"github.com/sujanv/pingforge/internal/core"
)

// WritePrometheusMetrics writes Prometheus formatted metrics to w.
func WritePrometheusMetrics(w io.Writer, results []core.ProbeResult, summaries []core.HistorySummary) {
	fmt.Fprintln(w, "# HELP pingforge_probe_duration_seconds Latency of probe in seconds")
	fmt.Fprintln(w, "# TYPE pingforge_probe_duration_seconds gauge")
	for _, r := range results {
		sec := r.LatencyMs / 1000.0
		fmt.Fprintf(w, "pingforge_probe_duration_seconds{target=%q,type=%q} %.6f\n",
			r.Target.String(), string(r.Type), sec)
	}

	fmt.Fprintln(w, "\n# HELP pingforge_probe_success Probe status (1=UP, 0=DOWN)")
	fmt.Fprintln(w, "# TYPE pingforge_probe_success gauge")
	for _, r := range results {
		val := 0
		if r.Status == core.StatusUp {
			val = 1
		}
		fmt.Fprintf(w, "pingforge_probe_success{target=%q,type=%q} %d\n",
			r.Target.String(), string(r.Type), val)
	}

	fmt.Fprintln(w, "\n# HELP pingforge_uptime_ratio Historical availability ratio (0.0 to 1.0)")
	fmt.Fprintln(w, "# TYPE pingforge_uptime_ratio gauge")
	for _, s := range summaries {
		ratio := s.UptimePercentage / 100.0
		fmt.Fprintf(w, "pingforge_uptime_ratio{target=%q} %.4f\n", s.Target, ratio)
	}

	fmt.Fprintln(w, "\n# HELP pingforge_latency_p95_seconds Historical 95th percentile latency")
	fmt.Fprintln(w, "# TYPE pingforge_latency_p95_seconds gauge")
	for _, s := range summaries {
		fmt.Fprintf(w, "pingforge_latency_p95_seconds{target=%q} %.6f\n", s.Target, s.P95LatencyMs/1000.0)
	}
}

// FormatSummaryJSON formats JSON summary output for API
func FormatSummaryJSON(summaries []core.HistorySummary) string {
	var sb strings.Builder
	sb.WriteString("[\n")
	for i, s := range summaries {
		sb.WriteString(fmt.Sprintf(`  {"target": %q, "uptime": %.2f, "avg_latency_ms": %.2f, "total": %d}`,
			s.Target, s.UptimePercentage, s.AvgLatencyMs, s.TotalProbes))
		if i < len(summaries)-1 {
			sb.WriteString(",\n")
		} else {
			sb.WriteString("\n")
		}
	}
	sb.WriteString("]")
	return sb.String()
}
