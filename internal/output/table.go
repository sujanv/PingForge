package output

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/sujanv/pingforge/internal/core"
)

// TableFormatter renders diagnostic results as attractive ANSI tables.
type TableFormatter struct {
	NoColor bool
}

// NewTableFormatter creates a new TableFormatter.
func NewTableFormatter(noColor bool) *TableFormatter {
	if noColor {
		color.NoColor = true
	}
	return &TableFormatter{NoColor: noColor}
}

// Format formats probe results or latency stats.
func (t *TableFormatter) Format(w io.Writer, data interface{}) error {
	switch v := data.(type) {
	case *core.ProbeResult:
		t.renderSingleProbe(w, v)
	case []core.ProbeResult:
		t.renderBatchProbes(w, v)
	case *core.LatencyStats:
		t.renderLatencyStats(w, v)
	case *core.ScanResult:
		t.renderScanResult(w, v)
	case *core.TraceResult:
		t.renderTraceResult(w, v)
	case []core.HistorySummary:
		t.renderHistorySummaries(w, v)
	default:
		fmt.Fprintf(w, "%+v\n", data)
	}
	return nil
}

func (t *TableFormatter) renderSingleProbe(w io.Writer, p *core.ProbeResult) {
	statusStr := colorizeStatus(p.Status)
	latencyStr := colorizeLatency(p.LatencyMs)

	fmt.Fprintf(w, "\n── %s %s [%s] ──\n\n", color.CyanString("PINGFORGE"), p.Type, p.Target.String())
	fmt.Fprintf(w, "  Status     : %s\n", statusStr)
	fmt.Fprintf(w, "  Latency    : %s\n", latencyStr)
	fmt.Fprintf(w, "  Message    : %s\n", p.Message)
	if p.Error != "" {
		fmt.Fprintf(w, "  Error      : %s\n", color.RedString(p.Error))
	}

	if p.TLS != nil {
		fmt.Fprintf(w, "\n  TLS Handshake & Certificate:\n")
		fmt.Fprintf(w, "    Version     : %s\n", p.TLS.Version)
		fmt.Fprintf(w, "    Cipher      : %s\n", p.TLS.CipherSuite)
		fmt.Fprintf(w, "    Subject     : %s\n", p.TLS.Subject)
		fmt.Fprintf(w, "    Issuer      : %s\n", p.TLS.Issuer)
		fmt.Fprintf(w, "    Expires In  : %d days (%s)\n", p.TLS.DaysUntilExpiration, p.TLS.NotAfter.Format("2006-01-02"))
		fmt.Fprintf(w, "    Verified    : %v\n", p.TLS.Verified)
	}

	if p.HTTP != nil {
		fmt.Fprintf(w, "\n  HTTP Connection Timing Breakdown:\n")
		fmt.Fprintf(w, "    DNS Lookup  : %.2f ms\n", p.HTTP.DNSLookupMs)
		fmt.Fprintf(w, "    TCP Connect : %.2f ms\n", p.HTTP.TCPConnectMs)
		fmt.Fprintf(w, "    TLS Handshake: %.2f ms\n", p.HTTP.TLSHandshakeMs)
		fmt.Fprintf(w, "    TTFB        : %.2f ms\n", p.HTTP.TTFBMs)
		fmt.Fprintf(w, "    Total       : %.2f ms\n", p.HTTP.TotalDurationMs)
		fmt.Fprintf(w, "    Status Code : %d\n", p.HTTP.StatusCode)
		if p.HTTP.RedirectCount > 0 {
			fmt.Fprintf(w, "    Redirects   : %d (final: %s)\n", p.HTTP.RedirectCount, p.HTTP.FinalURL)
		}
	}

	if p.DNS != nil {
		fmt.Fprintf(w, "\n  DNS Records (Resolver: %s, Query Time: %.2f ms):\n", p.DNS.Resolver, p.DNS.QueryTimeMs)
		for recType, vals := range p.DNS.Records {
			fmt.Fprintf(w, "    %-6s : %s\n", recType, strings.Join(vals, ", "))
		}
	}

	if len(p.Details) > 0 {
		fmt.Fprintf(w, "\n  Additional Metrics:\n")
		for k, v := range p.Details {
			fmt.Fprintf(w, "    %-18s : %v\n", k, v)
		}
	}
	fmt.Fprintln(w)
}

func (t *TableFormatter) renderBatchProbes(w io.Writer, results []core.ProbeResult) {
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "TARGET\tTYPE\tSTATUS\tLATENCY\tMESSAGE")
	fmt.Fprintln(tw, "------\t----\t------\t-------\t-------")

	for _, p := range results {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			p.Target.String(),
			string(p.Type),
			colorizeStatus(p.Status),
			colorizeLatency(p.LatencyMs),
			truncate(p.Message, 45),
		)
	}
	_ = tw.Flush()
	fmt.Fprintln(w)
}

func (t *TableFormatter) renderLatencyStats(w io.Writer, s *core.LatencyStats) {
	fmt.Fprintf(w, "\n── %s Latency Probe [%s] ──\n\n", color.CyanString("PINGFORGE"), s.Target.String())
	fmt.Fprintf(w, "  Probes Sent      : %d\n", s.Sent)
	fmt.Fprintf(w, "  Probes Received  : %d\n", s.Received)
	fmt.Fprintf(w, "  Packet Loss Rate : %.1f%%\n\n", s.LossRate)

	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "METRIC\tVALUE")
	fmt.Fprintln(tw, "------\t-----")
	fmt.Fprintf(tw, "Minimum Latency\t%.2f ms\n", s.MinMs)
	fmt.Fprintf(tw, "Average Latency\t%.2f ms\n", s.AvgMs)
	fmt.Fprintf(tw, "Maximum Latency\t%.2f ms\n", s.MaxMs)
	fmt.Fprintf(tw, "Standard Deviation (σ)\t%.2f ms\n", s.StdDevMs)
	fmt.Fprintf(tw, "RFC 3550 Jitter\t%.2f ms\n", s.JitterMs)
	fmt.Fprintf(tw, "P50 (Median)\t%.2f ms\n", s.P50Ms)
	fmt.Fprintf(tw, "P90 Percentile\t%.2f ms\n", s.P90Ms)
	fmt.Fprintf(tw, "P95 Percentile\t%.2f ms\n", s.P95Ms)
	fmt.Fprintf(tw, "P99 Percentile\t%.2f ms\n", s.P99Ms)
	_ = tw.Flush()
	fmt.Fprintln(w)
}

func (t *TableFormatter) renderScanResult(w io.Writer, s *core.ScanResult) {
	fmt.Fprintf(w, "\n── %s Port Scan [%s] (Scanned: %d, Duration: %.2f ms) ──\n\n",
		color.CyanString("PINGFORGE"), s.Target.Host, s.TotalScanned, s.DurationMs)

	if len(s.OpenPorts) == 0 {
		fmt.Fprintf(w, "  %s\n\n", color.YellowString("No open ports found in specified range."))
		return
	}

	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "PORT\tSTATE\tSERVICE\tLATENCY\tBANNER")
	fmt.Fprintln(tw, "----\t-----\t-------\t-------\t------")

	sort.Slice(s.OpenPorts, func(i, j int) bool {
		return s.OpenPorts[i].Port < s.OpenPorts[j].Port
	})

	for _, p := range s.OpenPorts {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%.2f ms\t%s\n",
			p.Port,
			color.GreenString("OPEN"),
			p.Service,
			p.LatencyMs,
			truncate(p.Banner, 30),
		)
	}
	_ = tw.Flush()
	fmt.Fprintln(w)
}

func (t *TableFormatter) renderTraceResult(w io.Writer, tr *core.TraceResult) {
	fmt.Fprintf(w, "\n── %s Route Trace [%s] ──\n\n", color.CyanString("PINGFORGE"), tr.Target.Host)

	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "HOP\tIP\tHOSTNAME\tLATENCY\tSTATUS")
	fmt.Fprintln(tw, "---\t--\t--------\t-------\t------")

	for _, hop := range tr.Hops {
		st := color.GreenString("REACHED")
		if !hop.Reached {
			st = color.RedString("TIMED OUT")
		}
		fmt.Fprintf(tw, "%d\t%s\t%s\t%.2f ms\t%s\n",
			hop.Hop,
			hop.IP,
			hop.Hostname,
			hop.LatencyMs,
			st,
		)
	}
	_ = tw.Flush()
	fmt.Fprintln(w)
}

func (t *TableFormatter) renderHistorySummaries(w io.Writer, summaries []core.HistorySummary) {
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "TARGET\tTOTAL\tUPTIME\tMIN (ms)\tAVG (ms)\tMAX (ms)\tP95 (ms)\tTREND")
	fmt.Fprintln(tw, "------\t-----\t------\t--------\t--------\t--------\t--------\t-----")

	for _, s := range summaries {
		uptimeColor := color.GreenString
		if s.UptimePercentage < 99.0 {
			uptimeColor = color.YellowString
		}
		if s.UptimePercentage < 95.0 {
			uptimeColor = color.RedString
		}

		fmt.Fprintf(tw, "%s\t%d\t%s\t%.2f\t%.2f\t%.2f\t%.2f\t%s\n",
			s.Target,
			s.TotalProbes,
			uptimeColor(fmt.Sprintf("%.1f%%", s.UptimePercentage)),
			s.MinLatencyMs,
			s.AvgLatencyMs,
			s.MaxLatencyMs,
			s.P95LatencyMs,
			s.Sparkline,
		)
	}
	_ = tw.Flush()
	fmt.Fprintln(w)
}

func colorizeStatus(s core.Status) string {
	switch s {
	case core.StatusUp:
		return color.GreenString("✔ UP")
	case core.StatusDown:
		return color.RedString("✘ DOWN")
	case core.StatusDegraded:
		return color.YellowString("⚠ DEGRADED")
	default:
		return color.HiBlackString("? UNKNOWN")
	}
}

func colorizeLatency(ms float64) string {
	str := fmt.Sprintf("%.2f ms", ms)
	if ms <= 0 {
		return color.HiBlackString(str)
	}
	if ms < 50.0 {
		return color.GreenString(str)
	}
	if ms < 150.0 {
		return color.CyanString(str)
	}
	if ms < 300.0 {
		return color.YellowString(str)
	}
	return color.RedString(str)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
