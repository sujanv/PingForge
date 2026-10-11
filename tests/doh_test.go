package tests

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sujanv/pingforge/internal/netprobe"
)

func TestProbeDoHMock(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/dns-json")
		fmt.Fprint(w, `{"Status":0,"Answer":[{"name":"example.com","type":1,"TTL":300,"data":"93.184.216.34"}]}`)
	}))
	defer ts.Close()

	ctx := context.Background()
	result := netprobe.ProbeDoH(ctx, "example.com", ts.URL, "A", 2*time.Second)

	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if len(result.Records["A"]) == 0 || result.Records["A"][0] != "93.184.216.34" {
		t.Errorf("expected 93.184.216.34, got %v", result.Records)
	}
}
