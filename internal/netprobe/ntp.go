package netprobe

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

// NTPResult stores precision network time synchronization metrics.
type NTPResult struct {
	Server        string    `json:"server"`
	Stratum       int       `json:"stratum"`
	Precision     int8      `json:"precision"`
	LeapIndicator int       `json:"leap_indicator"`
	ReferenceID   string    `json:"reference_id"`
	RootDelayMs   float64   `json:"root_delay_ms"`
	DispersionMs  float64   `json:"dispersion_ms"`
	OffsetMs      float64   `json:"offset_ms"`
	DelayMs       float64   `json:"delay_ms"`
	QueryTimeMs   float64   `json:"query_time_ms"`
	ServerTime    time.Time `json:"server_time"`
	Error         string    `json:"error,omitempty"`
}

// NTPPacket represents the RFC 5905 48-byte NTP client/server datagram.
type NTPPacket struct {
	LiVnMode       uint8
	Stratum        uint8
	Poll           int8
	Precision      int8
	RootDelay      uint32
	RootDispersion uint32
	ReferenceID    uint32
	RefTimestamp   uint64
	OrigTimestamp  uint64
	RxTimestamp    uint64
	TxTimestamp    uint64
}

const ntpEpochDelta = 2208988800

// ProbeNTP queries an RFC 5905 Network Time Protocol server on UDP port 123.
func ProbeNTP(ctx context.Context, host string, port int, timeout time.Duration) *NTPResult {
	if port <= 0 {
		port = 123
	}
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	res := &NTPResult{Server: addr}

	conn, err := net.DialTimeout("udp", addr, timeout)
	if err != nil {
		res.Error = fmt.Sprintf("dial ntp server: %v", err)
		return res
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(timeout))

	req := &NTPPacket{
		LiVnMode: 0x1B, // LI: 0, VN: 3, Mode: 3 (Client)
	}

	t0 := time.Now()
	req.TxTimestamp = toNTPTime(t0)

	if err := binary.Write(conn, binary.BigEndian, req); err != nil {
		res.Error = fmt.Sprintf("send ntp request: %v", err)
		return res
	}

	resp := &NTPPacket{}
	if err := binary.Read(conn, binary.BigEndian, resp); err != nil {
		res.Error = fmt.Sprintf("read ntp response: %v", err)
		return res
	}
	t3 := time.Now()

	res.Stratum = int(resp.Stratum)
	res.Precision = resp.Precision
	res.LeapIndicator = int((resp.LiVnMode >> 6) & 0x03)
	res.ReferenceID = decodeNTPRefID(resp.ReferenceID, resp.Stratum)
	res.RootDelayMs = float64(resp.RootDelay) / 65.536
	res.DispersionMs = float64(resp.RootDispersion) / 65.536

	t1 := fromNTPTime(resp.RxTimestamp)
	t2 := fromNTPTime(resp.TxTimestamp)
	res.ServerTime = t2

	// RFC 5905 Offset: ((t1 - t0) + (t2 - t3)) / 2
	offset := ((t1.Sub(t0)) + (t2.Sub(t3))) / 2
	res.OffsetMs = float64(offset.Microseconds()) / 1000.0

	// Round-trip delay: (t3 - t0) - (t2 - t1)
	delay := (t3.Sub(t0)) - (t2.Sub(t1))
	if delay < 0 {
		delay = t3.Sub(t0)
	}
	res.DelayMs = float64(delay.Microseconds()) / 1000.0
	res.QueryTimeMs = float64(time.Since(t0).Microseconds()) / 1000.0

	return res
}

func toNTPTime(t time.Time) uint64 {
	secs := uint64(t.Unix() + ntpEpochDelta)
	frac := (uint64(t.Nanosecond()) << 32) / 1e9
	return (secs << 32) | frac
}

func fromNTPTime(ntp uint64) time.Time {
	secs := int64(ntp>>32) - ntpEpochDelta
	nanos := int64((ntp & 0xFFFFFFFF) * 1e9 >> 32)
	return time.Unix(secs, nanos)
}

func decodeNTPRefID(refID uint32, stratum uint8) string {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, refID)
	if stratum <= 1 {
		return string(b)
	}
	return fmt.Sprintf("%d.%d.%d.%d", b[0], b[1], b[2], b[3])
}
