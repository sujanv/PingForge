package tests

import (
	"math"
	"testing"

	"github.com/sujanv/pingforge/internal/core"
	"github.com/sujanv/pingforge/internal/netprobe"
)

func TestCalculateStats(t *testing.T) {
	stats := &core.LatencyStats{
		Samples: []float64{10.0, 20.0, 30.0, 40.0, 50.0},
	}

	netprobe.CalculateStats(stats, 2.5)

	if stats.MinMs != 10.0 {
		t.Errorf("expected MinMs 10.0, got %.2f", stats.MinMs)
	}
	if stats.MaxMs != 50.0 {
		t.Errorf("expected MaxMs 50.0, got %.2f", stats.MaxMs)
	}
	if stats.AvgMs != 30.0 {
		t.Errorf("expected AvgMs 30.0, got %.2f", stats.AvgMs)
	}
	if stats.P50Ms != 30.0 {
		t.Errorf("expected P50Ms 30.0, got %.2f", stats.P50Ms)
	}

	// Variance for [10,20,30,40,50]: mean 30, sum sq diff = 400+100+0+100+400 = 1000, /4 = 250, sqrt = 15.811
	expectedStdDev := math.Sqrt(250)
	if math.Abs(stats.StdDevMs-expectedStdDev) > 0.001 {
		t.Errorf("expected StdDev %.3f, got %.3f", expectedStdDev, stats.StdDevMs)
	}

	if stats.JitterMs != 2.5 {
		t.Errorf("expected JitterMs 2.5, got %.2f", stats.JitterMs)
	}
}

func TestPercentiles(t *testing.T) {
	stats := &core.LatencyStats{
		Samples: []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0},
	}
	netprobe.CalculateStats(stats, 0)

	if stats.P90Ms < 9.0 {
		t.Errorf("expected P90 >= 9.0, got %.2f", stats.P90Ms)
	}
	if stats.P99Ms < 9.5 {
		t.Errorf("expected P99 >= 9.5, got %.2f", stats.P99Ms)
	}
}
