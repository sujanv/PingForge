package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/sujanv/pingforge/internal/core"
)

// GenerateHTMLReport creates a rich, self-contained HTML dashboard with responsive charts.
func GenerateHTMLReport(summaries []core.HistorySummary, title string) string {
	if title == "" {
		title = "PingForge - Network Diagnostics & Latency SLA Report"
	}

	totalTargets := len(summaries)
	totalProbes := 0
	uptimeSum := 0.0
	latencySum := 0.0

	for _, s := range summaries {
		totalProbes += s.TotalProbes
		uptimeSum += s.UptimePercentage
		latencySum += s.AvgLatencyMs
	}

	avgUptime := 100.0
	avgLatency := 0.0
	if totalTargets > 0 {
		avgUptime = uptimeSum / float64(totalTargets)
		avgLatency = latencySum / float64(totalTargets)
	}

	var rows strings.Builder
	for _, s := range summaries {
		statusBadge := "badge-success"
		if s.UptimePercentage < 99.0 {
			statusBadge = "badge-warning"
		}
		if s.UptimePercentage < 95.0 {
			statusBadge = "badge-danger"
		}

		rows.WriteString(fmt.Sprintf(`
			<tr>
				<td><strong>%s</strong></td>
				<td>%d</td>
				<td><span class="badge %s">%.2f%%</span></td>
				<td>%.2f ms</td>
				<td><strong>%.2f ms</strong></td>
				<td>%.2f ms</td>
				<td>%.2f ms</td>
				<td>%s</td>
			</tr>`,
			s.Target, s.TotalProbes, statusBadge, s.UptimePercentage,
			s.MinLatencyMs, s.AvgLatencyMs, s.MaxLatencyMs, s.P95LatencyMs,
			renderSVGSparkline(s.AvgLatencyMs, s.MaxLatencyMs),
		))
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>%s</title>
	<style>
		:root {
			--bg: #0f172a;
			--card: #1e293b;
			--border: #334155;
			--text: #f8fafc;
			--text-muted: #94a3b8;
			--primary: #38bdf8;
			--success: #22c55e;
			--warning: #f59e0b;
			--danger: #ef4444;
		}
		body {
			font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
			background-color: var(--bg);
			color: var(--text);
			margin: 0;
			padding: 2rem;
		}
		.container {
			max-width: 1200px;
			margin: 0 auto;
		}
		header {
			display: flex;
			justify-content: space-between;
			align-items: center;
			margin-bottom: 2rem;
			border-bottom: 1px solid var(--border);
			padding-bottom: 1rem;
		}
		h1 {
			font-size: 1.8rem;
			margin: 0;
			color: var(--primary);
			display: flex;
			align-items: center;
			gap: 0.5rem;
		}
		.meta {
			color: var(--text-muted);
			font-size: 0.9rem;
		}
		.grid {
			display: grid;
			grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
			gap: 1rem;
			margin-bottom: 2rem;
		}
		.card {
			background: var(--card);
			border: 1px solid var(--border);
			border-radius: 8px;
			padding: 1.2rem;
		}
		.card .label {
			color: var(--text-muted);
			font-size: 0.85rem;
			text-transform: uppercase;
			letter-spacing: 0.05em;
		}
		.card .value {
			font-size: 1.8rem;
			font-weight: 700;
			margin-top: 0.5rem;
			color: var(--text);
		}
		table {
			width: 100%%;
			border-collapse: collapse;
			background: var(--card);
			border: 1px solid var(--border);
			border-radius: 8px;
			overflow: hidden;
		}
		th, td {
			padding: 0.85rem 1rem;
			text-align: left;
			border-bottom: 1px solid var(--border);
		}
		th {
			background: #111827;
			color: var(--text-muted);
			font-size: 0.8rem;
			text-transform: uppercase;
			letter-spacing: 0.05em;
		}
		tr:hover {
			background: rgba(255, 255, 255, 0.02);
		}
		.badge {
			padding: 0.25rem 0.5rem;
			border-radius: 4px;
			font-size: 0.8rem;
			font-weight: 600;
		}
		.badge-success { background: rgba(34, 197, 94, 0.2); color: var(--success); }
		.badge-warning { background: rgba(245, 158, 11, 0.2); color: var(--warning); }
		.badge-danger { background: rgba(239, 68, 68, 0.2); color: var(--danger); }
		footer {
			margin-top: 3rem;
			text-align: center;
			color: var(--text-muted);
			font-size: 0.85rem;
		}
	</style>
</head>
<body>
	<div class="container">
		<header>
			<div>
				<h1>⚡ PingForge Historical Diagnostics</h1>
				<div class="meta">Generated at: %s</div>
			</div>
		</header>

		<div class="grid">
			<div class="card">
				<div class="label">Monitored Targets</div>
				<div class="value">%d</div>
			</div>
			<div class="card">
				<div class="label">Total Probes Executed</div>
				<div class="value">%d</div>
			</div>
			<div class="card">
				<div class="label">Average SLA Uptime</div>
				<div class="value" style="color: var(--success);">%.2f%%</div>
			</div>
			<div class="card">
				<div class="label">Fleet Average Latency</div>
				<div class="value" style="color: var(--primary);">%.2f ms</div>
			</div>
		</div>

		<div class="card" style="padding: 0; overflow-x: auto;">
			<table>
				<thead>
					<tr>
						<th>Target</th>
						<th>Probes</th>
						<th>Uptime SLA</th>
						<th>Min</th>
						<th>Avg</th>
						<th>Max</th>
						<th>P95</th>
						<th>Indicator</th>
					</tr>
				</thead>
				<tbody>
					%s
				</tbody>
			</table>
		</div>

		<footer>
			PingForge Advanced Network Diagnostics Toolkit • Built with Go • <a href="https://github.com/sujanv/PingForge" style="color: var(--primary); text-decoration: none;">GitHub Repository</a>
		</footer>
	</div>
</body>
</html>`,
		title,
		time.Now().Format("2006-01-02 15:04:05 MST"),
		totalTargets,
		totalProbes,
		avgUptime,
		avgLatency,
		rows.String(),
	)
}

func renderSVGSparkline(avg, max float64) string {
	ratio := 0.5
	if max > 0 {
		ratio = avg / max
	}
	width := int(ratio * 80)
	if width < 5 {
		width = 5
	}
	return fmt.Sprintf(`<svg width="80" height="12"><rect width="80" height="8" rx="4" fill="#334155"/><rect width="%d" height="8" rx="4" fill="#38bdf8"/></svg>`, width)
}
