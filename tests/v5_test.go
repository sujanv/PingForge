package tests

import (
	"bytes"
	"context"
	"encoding/binary"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sujanv/pingforge/internal/core"
	"github.com/sujanv/pingforge/internal/netprobe"
	"github.com/sujanv/pingforge/internal/report"
)

func TestProbeNTPMock(t *testing.T) {
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	go func() {
		buf := make([]byte, 48)
		_, addr, err := conn.ReadFrom(buf)
		if err != nil {
			return
		}
		resp := netprobe.NTPPacket{
			LiVnMode:    0x1C, // VN: 3, Mode: 4 (Server)
			Stratum:     2,
			ReferenceID: 0x7F000001,
		}
		var wbuf bytes.Buffer
		_ = binary.Write(&wbuf, binary.BigEndian, resp)
		_, _ = conn.WriteTo(wbuf.Bytes(), addr)
	}()

	addr := conn.LocalAddr().(*net.UDPAddr)
	ctx := context.Background()
	res := netprobe.ProbeNTP(ctx, "127.0.0.1", addr.Port, 500*time.Millisecond)

	if res.Server == "" {
		t.Errorf("expected server name")
	}
}

func TestBufferbloatCalculation(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("DATA"))
	}))
	defer ts.Close()

	ctx := context.Background()
	res := netprobe.ProbeBufferbloat(ctx, ts.URL, "127.0.0.1", 80, 500*time.Millisecond)

	if res.Grade == "" {
		t.Errorf("expected grade assigned")
	}
}

func TestWHOISMock(t *testing.T) {
	ctx := context.Background()
	res := netprobe.QueryWHOIS(ctx, "example.com", 2*time.Second)

	if res.Query != "example.com" {
		t.Errorf("expected query example.com, got %s", res.Query)
	}
}

func TestSNMPPacketBuilder(t *testing.T) {
	pkt := netprobe.BuildSNMPGetPacket("public", "1.3.6.1.2.1.1.1.0")
	if len(pkt) < 20 {
		t.Errorf("expected valid SNMP packet bytes, got %d", len(pkt))
	}
}

func TestDualStackProbe(t *testing.T) {
	ctx := context.Background()
	res := netprobe.ProbeDualStack(ctx, "localhost", 80, 1*time.Second)

	if res.Domain != "localhost" {
		t.Errorf("expected localhost")
	}
}

func TestCertExpiryAuditor(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	u := ts.Listener.Addr().(*net.TCPAddr)
	ctx := context.Background()
	summary := netprobe.AuditCertExpiry(ctx, "127.0.0.1", u.Port, 2*time.Second)

	if summary.UrgencyStatus == "UNKNOWN" {
		t.Errorf("expected calculated urgency status")
	}
}

func TestFormatInfluxLineProtocol(t *testing.T) {
	results := []core.ProbeResult{
		{
			Target:    core.Target{Host: "gateway.net"},
			Type:      core.ProbeTCP,
			Status:    core.StatusUp,
			LatencyMs: 14.52,
			Timestamp: time.Now(),
		},
	}
	line := report.FormatInfluxLineProtocol(results)
	if len(line) == 0 {
		t.Errorf("expected non-empty influx line")
	}
}

func TestFormatOTLPMetricsJSON(t *testing.T) {
	results := []core.ProbeResult{
		{
			Target:    core.Target{Host: "gateway.net"},
			Type:      core.ProbePing,
			Status:    core.StatusUp,
			LatencyMs: 8.42,
			Timestamp: time.Now(),
		},
	}
	data, err := report.FormatOTLPMetricsJSON(results)
	if err != nil || len(data) == 0 {
		t.Errorf("expected valid OTLP JSON data")
	}
}

func TestSNMPOIDConstants(t *testing.T) {
	if netprobe.OIDSysContact != "1.3.6.1.2.1.1.4.0" {
		t.Errorf("expected sysContact OID, got %s", netprobe.OIDSysContact)
	}
	if netprobe.OIDSysLocation != "1.3.6.1.2.1.1.6.0" {
		t.Errorf("expected sysLocation OID, got %s", netprobe.OIDSysLocation)
	}
}
