package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sujanv/pingforge/internal/core"
)

// JSONStore is a file-backed JSON-lines persistent time-series storage.
type JSONStore struct {
	mu       sync.RWMutex
	filePath string
	maxRecs  int
}

// NewJSONStore creates or opens the JSON store at filePath.
func NewJSONStore(filePath string, maxRecs int) (*JSONStore, error) {
	if filePath == "" {
		home, _ := os.UserHomeDir()
		filePath = filepath.Join(home, ".pingforge", "history.json")
	}
	if maxRecs <= 0 {
		maxRecs = 50000
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		// Fallback to local directory if home directory is restricted
		filePath = "history.json"
	}

	return &JSONStore{
		filePath: filePath,
		maxRecs:  maxRecs,
	}, nil
}

// Save appends a single probe result to disk.
func (s *JSONStore) Save(ctx context.Context, result *core.ProbeResult) error {
	return s.SaveBatch(ctx, []core.ProbeResult{*result})
}

// SaveBatch appends multiple probe results to disk in a single transaction.
func (s *JSONStore) SaveBatch(ctx context.Context, results []core.ProbeResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.OpenFile(s.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open storage file: %w", err)
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	for _, res := range results {
		data, err := json.Marshal(res)
		if err != nil {
			continue
		}
		if _, err := writer.Write(append(data, '\n')); err != nil {
			return err
		}
	}
	return writer.Flush()
}

// GetHistory reads historical probe records matching target within since duration.
func (s *JSONStore) GetHistory(ctx context.Context, target string, since time.Duration) ([]core.ProbeResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	f, err := os.Open(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var results []core.ProbeResult
	cutoff := time.Time{}
	if since > 0 {
		cutoff = time.Now().Add(-since)
	}

	scanner := bufio.NewScanner(f)
	// Allow larger buffers for records with DNS details
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var res core.ProbeResult
		if err := json.Unmarshal(line, &res); err != nil {
			continue
		}

		if !cutoff.IsZero() && res.Timestamp.Before(cutoff) {
			continue
		}

		if target != "" && res.Target.String() != target && res.Target.Host != target {
			continue
		}

		results = append(results, res)
	}

	return results, scanner.Err()
}

// GetSummaries compiles SLA, uptime, and latency summaries per target.
func (s *JSONStore) GetSummaries(ctx context.Context, since time.Duration) ([]core.HistorySummary, error) {
	all, err := s.GetHistory(ctx, "", since)
	if err != nil {
		return nil, err
	}

	return AggregateSummaries(all), nil
}

// Close flushes and closes resources.
func (s *JSONStore) Close() error {
	return nil
}
