package storage

import (
	"math"
	"time"

	"github.com/sujanv/pingforge/internal/core"
)

// AnomalyReport highlights latency spikes and statistical outliers.
type AnomalyReport struct {
	Target          string    `json:"target"`
	Timestamp       time.Time `json:"timestamp"`
	ObservedLatency float64   `json:"observed_latency_ms"`
	MeanLatency     float64   `json:"mean_latency_ms"`
	StdDev          float64   `json:"std_dev_ms"`
	ZScore          float64   `json:"z_score"`
	IsSpike         bool      `json:"is_spike"`
}

// DetectAnomalies evaluates probe records and identifies latency anomalies (Z-Score >= 2.5).
func DetectAnomalies(records []core.ProbeResult, thresholdZ float64) []AnomalyReport {
	if thresholdZ <= 0 {
		thresholdZ = 2.5
	}

	grouped := make(map[string][]core.ProbeResult)
	for _, r := range records {
		grouped[r.Target.String()] = append(grouped[r.Target.String()], r)
	}

	var anomalies []AnomalyReport

	for target, group := range grouped {
		if len(group) < 5 {
			continue
		}

		var sum float64
		var count int
		for _, r := range group {
			if r.LatencyMs > 0 {
				sum += r.LatencyMs
				count++
			}
		}

		if count < 5 {
			continue
		}

		mean := sum / float64(count)

		var varianceSum float64
		for _, r := range group {
			if r.LatencyMs > 0 {
				diff := r.LatencyMs - mean
				varianceSum += diff * diff
			}
		}
		stdDev := math.Sqrt(varianceSum / float64(count-1))
		if stdDev <= 0 {
			continue
		}

		for _, r := range group {
			if r.LatencyMs > 0 {
				z := (r.LatencyMs - mean) / stdDev
				if z >= thresholdZ {
					anomalies = append(anomalies, AnomalyReport{
						Target:          target,
						Timestamp:       r.Timestamp,
						ObservedLatency: r.LatencyMs,
						MeanLatency:     mean,
						StdDev:          stdDev,
						ZScore:          z,
						IsSpike:         true,
					})
				}
			}
		}
	}

	return anomalies
}
