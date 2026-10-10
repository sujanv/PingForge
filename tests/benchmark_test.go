package tests

import (
	"context"
	"testing"

	"github.com/sujanv/pingforge/internal/core"
	"github.com/sujanv/pingforge/internal/engine"
	"github.com/sujanv/pingforge/internal/netprobe"
	"github.com/sujanv/pingforge/internal/report"
)

func BenchmarkCalculateStats(b *testing.B) {
	samples := make([]float64, 100)
	for i := 0; i < 100; i++ {
		samples[i] = float64(10 + (i % 25))
	}
	stats := &core.LatencyStats{
		Samples: samples,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		netprobe.CalculateStats(stats, 1.25)
	}
}

func BenchmarkWorkerPool(b *testing.B) {
	items := make([]int, 50)
	for i := 0; i < 50; i++ {
		items[i] = i
	}
	pool := engine.NewPool[int, int](10, func(ctx context.Context, item int) int {
		return item * item
	})

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pool.Run(ctx, items)
	}
}

func BenchmarkGenerateSparkline(b *testing.B) {
	samples := []float64{12.4, 15.1, 14.8, 22.0, 18.3, 16.5, 30.1, 14.2, 12.9, 13.5}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = report.GenerateSparkline(samples)
	}
}
