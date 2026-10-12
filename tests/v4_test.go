package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sujanv/pingforge/internal/alerts"
	"github.com/sujanv/pingforge/internal/netprobe"
)

func TestProbeHTTP3AltSvc(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Alt-Svc", `h3=":443"; ma=2592000, h3-29=":443"; ma=2592000`)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ctx := context.Background()
	res := netprobe.ProbeHTTP3(ctx, ts.URL, 2*time.Second)

	if !res.QUICSupported {
		t.Errorf("expected QUIC supported via Alt-Svc")
	}
	if res.H3DraftVersion != `h3=":443"` {
		t.Errorf("expected h3 draft, got %s", res.H3DraftVersion)
	}
}

func TestMTRDiagnostics(t *testing.T) {
	ctx := context.Background()
	report := netprobe.RunMTRDiagnostics(ctx, "127.0.0.1", 80, 2, 500*time.Millisecond)

	if report.CycleCount != 2 {
		t.Errorf("expected 2 cycles, got %d", report.CycleCount)
	}
	if len(report.Hops) == 0 {
		t.Errorf("expected hops generated")
	}
}

func TestInspectOCSPStaplingMock(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ctx := context.Background()
	res := netprobe.InspectOCSPStapling(ctx, "127.0.0.1", 443, 2*time.Second)

	if res.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %s", res.Host)
	}
}

func TestDispatchWebhook(t *testing.T) {
	received := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = true
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	payload := alerts.WebhookPayload{
		Title:    "Test Alert",
		Severity: "CRITICAL",
		Message:  "PingForge Latency Spike",
	}

	ctx := context.Background()
	err := alerts.DispatchWebhook(ctx, ts.URL, payload, 2*time.Second)
	if err != nil {
		t.Fatalf("webhook dispatch failed: %v", err)
	}
	if !received {
		t.Errorf("expected webhook received")
	}
}
