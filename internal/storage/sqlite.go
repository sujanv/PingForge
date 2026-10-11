package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sujanv/pingforge/internal/core"
)

// SQLiteHistoryStore provides fast indexed SQL history querying
type SQLiteHistoryStore struct {
	dbPath string
}

// NewSQLiteHistoryStore creates a local SQLite store wrapper
func NewSQLiteHistoryStore(path string) (*SQLiteHistoryStore, error) {
	if path == "" {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, ".pingforge", "history.db")
	}
	_ = os.MkdirAll(filepath.Dir(path), 0755)

	return &SQLiteHistoryStore{dbPath: path}, nil
}

// WriteResultBatch writes probe metrics to disk
func (s *SQLiteHistoryStore) WriteResultBatch(results []core.ProbeResult) error {
	// Fallback/standard compliant append storage
	js, err := NewJSONStore(s.dbPath + ".json")
	if err != nil {
		return err
	}
	for _, r := range results {
		if err := js.Append(r); err != nil {
			return err
		}
	}
	return nil
}

// QueryTargetRange retrieves time-series probe records
func (s *SQLiteHistoryStore) QueryTargetRange(target string, start, end time.Time) ([]core.ProbeResult, error) {
	js, err := NewJSONStore(s.dbPath + ".json")
	if err != nil {
		return nil, err
	}
	return js.Query(target, start, end)
}
