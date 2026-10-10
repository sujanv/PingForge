package tests

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sujanv/pingforge/internal/core"
	"github.com/sujanv/pingforge/internal/netprobe"
)

func TestProbeHealthSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok","version":"1.0"}`)
	}))
	defer ts.Close()

	opts := netprobe.HealthCheckOptions{
		ExpectedStatus: []int{200},
		BodyRegex:      `"status":"ok"`,
		Timeout:        2 * time.Second,
	}

	ctx := context.Background()
	result := netprobe.ProbeHealth(ctx, ts.URL, opts)

	if result.Status != core.StatusUp {
		t.Fatalf("expected status UP, got %v (err: %s)", result.Status, result.Error)
	}
	if result.HTTP.StatusCode != 200 {
		t.Errorf("expected 200, got %d", result.HTTP.StatusCode)
	}
	if !result.HTTP.BodyMatched {
		t.Errorf("expected body to match regex")
	}
}

func TestProbeHealthFailureStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	opts := netprobe.HealthCheckOptions{
		ExpectedStatus: []int{200},
		Timeout:        2 * time.Second,
	}

	ctx := context.Background()
	result := netprobe.ProbeHealth(ctx, ts.URL, opts)

	if result.Status != core.StatusDown {
		t.Fatalf("expected status DOWN for 500, got %v", result.Status)
	}
}
