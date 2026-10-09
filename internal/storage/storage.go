package storage

import (
	"context"
	"time"

	"github.com/sujanv/pingforge/internal/core"
)

// Storage defines the interface for persisting and retrieving probe history.
type Storage interface {
	Save(ctx context.Context, result *core.ProbeResult) error
	SaveBatch(ctx context.Context, results []core.ProbeResult) error
	GetHistory(ctx context.Context, target string, since time.Duration) ([]core.ProbeResult, error)
	GetSummaries(ctx context.Context, since time.Duration) ([]core.HistorySummary, error)
	Close() error
}
