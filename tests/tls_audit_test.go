package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/sujanv/pingforge/internal/netprobe"
)

func TestAuditTLSMock(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	port, _ := strconv.Atoi(u.Port())

	ctx := context.Background()
	audit := netprobe.AuditTLS(ctx, u.Hostname(), port, 2*time.Second)

	if audit.SelectedVersion == "" {
		t.Errorf("expected selected version, got empty")
	}
	if len(audit.CertChain) == 0 {
		t.Errorf("expected cert chain, got empty")
	}
}
