package storage

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/sujanv/pingforge/internal/core"
)

// SQLiteHistoryStore provides indexed SQL history querying capabilities.
type SQLiteHistoryStore struct {
	dbPath string
}

// NewSQLiteHistoryStore creates a local SQLite store wrapper.
func NewSQLiteHistoryStore(path string) (*SQLiteHistoryStore, error) {
	if path == "" {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, ".pingforge", "history.db")
	}
	_ = os.MkdirAll(filepath.Dir(path), 0755)

	return &SQLiteHistoryStore{dbPath: path}, nil
}

// WriteResultBatch writes probe metrics to disk.
func (s *SQLiteHistoryStore) WriteResultBatch(results []core.ProbeResult) error {
	js, err := NewJSONStore(s.dbPath+".json", 50000)
	if err != nil {
		return err
	}
	return js.SaveBatch(context.Background(), results)
}

// QueryTargetRange retrieves time-series probe records.
func (s *SQLiteHistoryStore) QueryTargetRange(target string, start, end time.Time) ([]core.ProbeResult, error) {
	js, err := NewJSONStore(s.dbPath+".json", 50000)
	if err != nil {
		return nil, err
	}
	since := time.Since(start)
	return js.GetHistory(context.Background(), target, since)
}
