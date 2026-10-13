package netprobe

import (
	"context"
	"fmt"
	"net"
	"time"
)

// SNMPResult details responses from network switches, routers, and SNMP agents.
type SNMPResult struct {
	Host        string  `json:"host"`
	Port        int     `json:"port"`
	Community   string  `json:"community"`
	OID         string  `json:"oid"`
	Value       string  `json:"value"`
	LatencyMs   float64 `json:"latency_ms"`
	IsAvailable bool    `json:"available"`
	Error       string  `json:"error,omitempty"`
}

// ProbeSNMP executes a lightweight SNMP v2c GetRequest on UDP port 161.
func ProbeSNMP(ctx context.Context, host string, port int, community string, oid string, timeout time.Duration) *SNMPResult {
	if port <= 0 {
		port = 161
	}
	if community == "" {
		community = "public"
	}
	if oid == "" {
		oid = "1.3.6.1.2.1.1.1.0" // sysDescr.0
	}
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	res := &SNMPResult{
		Host:      host,
		Port:      port,
		Community: community,
		OID:       oid,
	}

	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("udp", addr, timeout)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(timeout))

	packet := BuildSNMPGetPacket(community, oid)
	start := time.Now()
	if _, err := conn.Write(packet); err != nil {
		res.Error = err.Error()
		return res
	}

	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	res.LatencyMs = float64(time.Since(start).Microseconds()) / 1000.0

	if err != nil {
		res.Error = fmt.Sprintf("snmp read timeout or unreachable: %v", err)
		return res
	}

	res.IsAvailable = true
	res.Value = fmt.Sprintf("SNMP OK (%d bytes payload received)", n)
	return res
}

// BuildSNMPGetPacket generates a basic SNMP v2c GetRequest PDU.
func BuildSNMPGetPacket(community, oid string) []byte {
	commBytes := []byte(community)
	pdu := []byte{
		0x30, byte(28 + len(commBytes)), // Sequence
		0x02, 0x01, 0x01,                 // Version: 1 (SNMP v2c)
		0x04, byte(len(commBytes)),       // Community OctetString
	}
	pdu = append(pdu, commBytes...)
	pdu = append(pdu, []byte{
		0xa0, 0x17,                       // GetRequest PDU
		0x02, 0x04, 0x12, 0x34, 0x56, 0x78, // Request ID
		0x02, 0x01, 0x00,                 // Error Status: 0
		0x02, 0x01, 0x00,                 // Error Index: 0
		0x30, 0x0b,                       // VarBindList Sequence
		0x30, 0x09,                       // VarBind Sequence
		0x06, 0x05, 0x2b, 0x06, 0x01, 0x02, 0x01, // 1.3.6.1.2.1
		0x05, 0x00,                       // Null Value
	}...)
	return pdu
}
