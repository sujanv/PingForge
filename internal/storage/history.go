package storage

import (
	"math"
	"sort"

	"github.com/sujanv/pingforge/internal/core"
)

// AggregateSummaries computes SLA uptime and latency distribution for records.
func AggregateSummaries(records []core.ProbeResult) []core.HistorySummary {
	grouped := make(map[string][]core.ProbeResult)
	for _, r := range records {
		tgt := r.Target.String()
		grouped[tgt] = append(grouped[tgt], r)
	}

	summaries := make([]core.HistorySummary, 0, len(grouped))

	for target, group := range grouped {
		if len(group) == 0 {
			continue
		}

		total := len(group)
		success := 0
		var latencies []float64
		var sumLat float64
		minLat := math.MaxFloat64
		maxLat := -math.MaxFloat64
		firstSeen := group[0].Timestamp
		lastSeen := group[0].Timestamp

		for _, r := range group {
			if r.Timestamp.Before(firstSeen) {
				firstSeen = r.Timestamp
			}
			if r.Timestamp.After(lastSeen) {
				lastSeen = r.Timestamp
			}

			if r.Status == core.StatusUp {
				success++
			}

			if r.LatencyMs > 0 {
				lat := r.LatencyMs
				latencies = append(latencies, lat)
				sumLat += lat
				if lat < minLat {
					minLat = lat
				}
				if lat > maxLat {
					maxLat = lat
				}
			}
		}

		uptime := 0.0
		if total > 0 {
			uptime = (float64(success) / float64(total)) * 100.0
		}

		avgLat := 0.0
		p95 := 0.0
		if len(latencies) > 0 {
			avgLat = sumLat / float64(len(latencies))
			sort.Float64s(latencies)
			idx := int(float64(len(latencies)-1) * 0.95)
			p95 = latencies[idx]
		} else {
			minLat = 0
			maxLat = 0
		}

		summaries = append(summaries, core.HistorySummary{
			Target:           target,
			TotalProbes:      total,
			SuccessfulProbes: success,
			FailedProbes:     total - success,
			UptimePercentage: uptime,
			MinLatencyMs:     minLat,
			AvgLatencyMs:     avgLat,
			MaxLatencyMs:     maxLat,
			P95LatencyMs:     p95,
			FirstSeen:        firstSeen,
			LastSeen:         lastSeen,
		})
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Target < summaries[j].Target
	})

	return summaries
}
