package report

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sujanv/pingforge/internal/core"
)

// FormatInfluxLineProtocol encodes probe results to InfluxDB 2.x Line Protocol format.
func FormatInfluxLineProtocol(results []core.ProbeResult) string {
	var sb strings.Builder
	for _, r := range results {
		statusVal := 1
		if r.Status != core.StatusUp {
			statusVal = 0
		}
		sb.WriteString(fmt.Sprintf("pingforge_latency,target=%s,protocol=%s,status=%s latency_ms=%.3f,success=%di %d\n",
			escapeTag(r.Target.Host),
			escapeTag(string(r.Type)),
			escapeTag(string(r.Status)),
			r.LatencyMs,
			statusVal,
			r.Timestamp.UnixNano(),
		))
	}
	return sb.String()
}

// FormatOTLPMetricsJSON serializes telemetry into OpenTelemetry Metric JSON structure.
func FormatOTLPMetricsJSON(results []core.ProbeResult) ([]byte, error) {
	type DataPoint struct {
		TimeUnixNano int64             `json:"time_unix_nano"`
		AsDouble     float64           `json:"as_double"`
		Attributes   map[string]string `json:"attributes"`
	}

	type Metric struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Unit        string `json:"unit"`
		Gauge       struct {
			DataPoints []DataPoint `json:"data_points"`
		} `json:"gauge"`
	}

	metric := Metric{
		Name:        "pingforge.probe.latency",
		Description: "End-to-end network diagnostic probe latency",
		Unit:        "ms",
	}

	for _, r := range results {
		dp := DataPoint{
			TimeUnixNano: r.Timestamp.UnixNano(),
			AsDouble:     r.LatencyMs,
			Attributes: map[string]string{
				"target":   r.Target.Host,
				"protocol": string(r.Type),
				"status":   string(r.Status),
			},
		}
		metric.Gauge.DataPoints = append(metric.Gauge.DataPoints, dp)
	}

	wrapper := struct {
		Metrics []Metric `json:"metrics"`
	}{
		Metrics: []Metric{metric},
	}

	return json.MarshalIndent(wrapper, "", "  ")
}

func escapeTag(s string) string {
	s = strings.ReplaceAll(s, " ", "\\ ")
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, "=", "\\=")
	return s
}
