package engine

import (
	"context"
	"time"

	"github.com/sujanv/pingforge/internal/config"
	"github.com/sujanv/pingforge/internal/core"
	"github.com/sujanv/pingforge/internal/netprobe"
)

// Runner manages concurrent probe executions across multiple targets.
type Runner struct {
	cfg config.Config
}

// NewRunner creates a new runner instance.
func NewRunner(cfg config.Config) *Runner {
	return &Runner{cfg: cfg}
}

// ExecuteProbes executes probes across all targets concurrently.
func (r *Runner) ExecuteProbes(ctx context.Context, targets []core.Target, probeType core.ProbeType) []core.ProbeResult {
	concurrency := r.cfg.Concurrency
	if concurrency <= 0 {
		concurrency = 10
	}

	task := func(taskCtx context.Context, t core.Target) core.ProbeResult {
		timeout := t.Timeout
		if timeout <= 0 {
			timeout = r.cfg.Timeout
		}

		switch probeType {
		case core.ProbeHealth:
			targetURL := t.URL
			if targetURL == "" {
				targetURL = "http://" + t.Host
			}
			opts := netprobe.HealthCheckOptions{
				ExpectedStatus:  r.cfg.Health.ExpectedStatus,
				FollowRedirects: r.cfg.Health.FollowRedirects,
				MaxRedirects:    r.cfg.Health.MaxRedirects,
				UserAgent:       r.cfg.Health.UserAgent,
				LatencyBudget:   r.cfg.Health.LatencyBudget,
				Timeout:         timeout,
			}
			res := netprobe.ProbeHealth(taskCtx, targetURL, opts)
			return *res

		case core.ProbeDNS:
			res := netprobe.ProbeDNS(taskCtx, t.Host, r.cfg.DNS)
			return *res

		case core.ProbePing:
			port := t.Port
			if port <= 0 {
				port = r.cfg.Ping.Port
				if port <= 0 {
					port = 80
				}
			}
			stats := netprobe.ProbeLatency(taskCtx, t.Host, port, r.cfg.Ping, timeout)
			status := core.StatusUp
			if stats.Received == 0 {
				status = core.StatusDown
			} else if stats.LossRate > 20.0 {
				status = core.StatusDegraded
			}
			return core.ProbeResult{
				ID:        t.Host,
				Target:    t,
				Type:      core.ProbePing,
				Status:    status,
				Latency:   time.Duration(stats.AvgMs * float64(time.Millisecond)),
				LatencyMs: stats.AvgMs,
				Timestamp: time.Now(),
				Message:   "Ping statistics collected",
				Details: map[string]interface{}{
					"min_ms":    stats.MinMs,
					"avg_ms":    stats.AvgMs,
					"max_ms":    stats.MaxMs,
					"jitter_ms": stats.JitterMs,
					"loss_rate": stats.LossRate,
					"p95_ms":    stats.P95Ms,
				},
			}

		case core.ProbeTCP:
			fallthrough
		default:
			port := t.Port
			if port <= 0 {
				port = 80
			}
			res := netprobe.ProbeTCP(taskCtx, t.Host, port, r.cfg.TCP, timeout)
			return *res
		}
	}

	pool := NewPool[core.Target, core.ProbeResult](concurrency, task)
	return pool.Run(ctx, targets)
}
