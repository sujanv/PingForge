package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/sujanv/pingforge/internal/config"
	"github.com/sujanv/pingforge/internal/core"
	"github.com/sujanv/pingforge/internal/engine"
	"github.com/sujanv/pingforge/internal/storage"
)

// Server coordinates background daemon polling and metrics serving.
type Server struct {
	cfg     config.Config
	targets []core.Target
	store   storage.Storage
	runner  *engine.Runner
	mu      sync.RWMutex
	latest  []core.ProbeResult
}

// NewServer initializes a daemon server.
func NewServer(cfg config.Config, targets []core.Target, store storage.Storage) *Server {
	return &Server{
		cfg:     cfg,
		targets: targets,
		store:   store,
		runner:  engine.NewRunner(cfg),
		latest:  make([]core.ProbeResult, 0),
	}
}

// Start begins continuous polling and launches HTTP listener.
func (s *Server) Start(ctx context.Context, listenAddr string, interval time.Duration) error {
	if interval <= 0 {
		interval = 30 * time.Second
	}

	// Trigger initial probe
	s.pollOnce(ctx)

	// Polling loop in background
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.pollOnce(ctx)
			}
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", s.handleMetrics)
	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/api/summary", s.handleSummary)
	mux.HandleFunc("/api/latest", s.handleLatest)

	server := &http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	log.Printf("[PingForge Daemon] Listening on %s (polling %d targets every %s)", listenAddr, len(s.targets), interval)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) pollOnce(ctx context.Context) {
	if len(s.targets) == 0 {
		return
	}

	results := s.runner.ExecuteProbes(ctx, s.targets, core.ProbeTCP)

	s.mu.Lock()
	s.latest = results
	s.mu.Unlock()

	if s.store != nil {
		_ = s.store.SaveBatch(ctx, results)
	}
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	latest := s.latest
	s.mu.RUnlock()

	var summaries []core.HistorySummary
	if s.store != nil {
		summaries, _ = s.store.GetSummaries(r.Context(), 24*time.Hour)
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	WritePrometheusMetrics(w, latest, summaries)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

func (s *Server) handleSummary(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		http.Error(w, "Storage not configured", http.StatusServiceUnavailable)
		return
	}

	summaries, err := s.store.GetSummaries(r.Context(), 24*time.Hour)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(summaries)
}

func (s *Server) handleLatest(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.latest)
}
