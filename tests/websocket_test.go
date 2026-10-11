package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sujanv/pingforge/internal/netprobe"
)

func TestProbeWebSocketMock(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Upgrade") == "websocket" {
			w.Header().Set("Upgrade", "websocket")
			w.Header().Set("Connection", "Upgrade")
			w.Header().Set("Sec-WebSocket-Accept", "s3pPLMBiTxaQ9kYGzzhZRbK+xOo=")
			w.WriteHeader(http.StatusSwitchingProtocols)
		}
	}))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")
	ctx := context.Background()
	result := netprobe.ProbeWebSocket(ctx, wsURL, nil, 2*time.Second)

	if result.Error == "" && !result.Connected {
		t.Errorf("expected connected")
	}
}
