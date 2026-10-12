package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/sujanv/pingforge/internal/core"
	"github.com/sujanv/pingforge/internal/netprobe"
	"github.com/sujanv/pingforge/internal/report"
)

func TestPCAPWriter(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.pcap")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	w, err := netprobe.NewPCAPWriter(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to create pcap writer: %v", err)
	}

	pkt := netprobe.BuildSyntheticTCPPacket([4]byte{127, 0, 0, 1}, [4]byte{127, 0, 0, 1}, 12345, 80, []byte("HELLO-PCAP"))
	if err := w.WritePacket(time.Now(), pkt); err != nil {
		t.Fatalf("write packet failed: %v", err)
	}
	_ = w.Close()

	info, err := os.Stat(tmpFile.Name())
	if err != nil || info.Size() < 24 {
		t.Errorf("expected valid pcap file, got size %d", info.Size())
	}
}

func TestHTTP2Probe(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "pingforge-v3")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ctx := context.Background()
	res := netprobe.ProbeHTTP2(ctx, ts.URL, 2*time.Second)

	if res.StatusCode != 200 {
		t.Errorf("expected 200, got %d", res.StatusCode)
	}
	if res.StreamHeaders["X-Custom-Header"] != "pingforge-v3" {
		t.Errorf("expected custom header value")
	}
}

func TestDNSSECValidatorMock(t *testing.T) {
	ctx := context.Background()
	status := netprobe.ValidateDNSSEC(ctx, "cloudflare.com", 2*time.Second)

	if status.Domain != "cloudflare.com" {
		t.Errorf("expected domain cloudflare.com, got %s", status.Domain)
	}
}

func TestPostScriptReport(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-report-*.ps")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	results := []core.ProbeResult{
		{Target: core.Target{Host: "example.com"}, Type: core.ProbeTCP, Status: core.StatusUp, LatencyMs: 12.5},
	}

	err = report.GeneratePostScriptReport("PingForge Telemetry", results, tmpFile.Name())
	if err != nil {
		t.Fatalf("generate report failed: %v", err)
	}

	data, err := os.ReadFile(tmpFile.Name())
	if err != nil || len(data) == 0 {
		t.Errorf("expected non-empty report")
	}
}
