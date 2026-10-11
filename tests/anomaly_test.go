package tests

import (
	"testing"
	"time"

	"github.com/sujanv/pingforge/internal/core"
	"github.com/sujanv/pingforge/internal/storage"
)

func TestDetectAnomalies(t *testing.T) {
	tgt := core.Target{Host: "api.example.com"}
	now := time.Now()

	records := []core.ProbeResult{
		{Target: tgt, LatencyMs: 10.0, Timestamp: now.Add(-10 * time.Minute)},
		{Target: tgt, LatencyMs: 11.0, Timestamp: now.Add(-9 * time.Minute)},
		{Target: tgt, LatencyMs: 10.5, Timestamp: now.Add(-8 * time.Minute)},
		{Target: tgt, LatencyMs: 9.8, Timestamp: now.Add(-7 * time.Minute)},
		{Target: tgt, LatencyMs: 10.2, Timestamp: now.Add(-6 * time.Minute)},
		{Target: tgt, LatencyMs: 100.0, Timestamp: now.Add(-5 * time.Minute)}, // Anomaly spike
	}

	anomalies := storage.DetectAnomalies(records, 2.0)
	if len(anomalies) != 1 {
		t.Fatalf("expected 1 anomaly, got %d", len(anomalies))
	}
	if anomalies[0].ObservedLatency != 100.0 {
		t.Errorf("expected 100.0ms spike, got %.2f", anomalies[0].ObservedLatency)
	}
}
