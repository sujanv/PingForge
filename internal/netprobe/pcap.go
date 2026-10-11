package netprobe

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"
)

// PCAPGlobalHeader represents standard libpcap 24-byte global header
type PCAPGlobalHeader struct {
	MagicNumber  uint32 // 0xa1b2c3d4 (microsecond)
	VersionMajor uint16 // 2
	VersionMinor uint16 // 4
	ThisZone     int32  // GMT to local correction
	SigFigs      uint32 // accuracy of timestamps
	SnapLen      uint32 // max length of captured packets (65535)
	Network      uint32 // data link type (1 = Ethernet)
}

// PCAPPacketHeader represents standard libpcap 16-byte packet header
type PCAPPacketHeader struct {
	TimestampSec  uint32
	TimestampUsec uint32
	CaptureLength uint32
	OriginalLength uint32
}

// PCAPWriter streams synthetic network probes into standard .pcap captures
type PCAPWriter struct {
	file *os.File
}

// NewPCAPWriter creates and initializes a libpcap file
func NewPCAPWriter(path string) (*PCAPWriter, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("create pcap file: %w", err)
	}

	gh := PCAPGlobalHeader{
		MagicNumber:  0xa1b2c3d4,
		VersionMajor: 2,
		VersionMinor: 4,
		SnapLen:      65535,
		Network:      1, // LINKTYPE_ETHERNET
	}

	if err := binary.Write(f, binary.LittleEndian, gh); err != nil {
		f.Close()
		return nil, fmt.Errorf("write pcap global header: %w", err)
	}

	return &PCAPWriter{file: f}, nil
}

// WritePacket writes an Ethernet/IP packet record to the PCAP file
func (w *PCAPWriter) WritePacket(t time.Time, rawPacket []byte) error {
	sec := uint32(t.Unix())
	usec := uint32(t.Nanosecond() / 1000)

	ph := PCAPPacketHeader{
		TimestampSec:  sec,
		TimestampUsec: usec,
		CaptureLength: uint32(len(rawPacket)),
		OriginalLength: uint32(len(rawPacket)),
	}

	if err := binary.Write(w.file, binary.LittleEndian, ph); err != nil {
		return err
	}

	_, err := w.file.Write(rawPacket)
	return err
}

// Close closes the underlying pcap file
func (w *PCAPWriter) Close() error {
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// BuildSyntheticTCPPacket creates a mock IPv4+TCP segment for export testing
func BuildSyntheticTCPPacket(srcIP, dstIP [4]byte, srcPort, dstPort uint16, payload []byte) []byte {
	// 14 byte Ethernet Header + 20 byte IPv4 Header + 20 byte TCP Header + payload
	totalLen := 14 + 20 + 20 + len(payload)
	buf := make([]byte, totalLen)

	// Ethernet Header (Broadcast dst, local src, EtherType 0x0800 IPv4)
	copy(buf[0:6], []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
	copy(buf[6:12], []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x01})
	binary.BigEndian.PutUint16(buf[12:14], 0x0800)

	// IPv4 Header
	ipStart := 14
	buf[ipStart] = 0x45 // Version 4, IHL 5
	buf[ipStart+1] = 0  // DSCP / ECN
	binary.BigEndian.PutUint16(buf[ipStart+2:ipStart+4], uint16(20+20+len(payload)))
	binary.BigEndian.PutUint16(buf[ipStart+4:ipStart+6], 12345) // Identification
	buf[ipStart+6] = 0x40 // Flags (Don't Fragment)
	buf[ipStart+7] = 0
	buf[ipStart+8] = 64 // TTL
	buf[ipStart+9] = 6  // Protocol (TCP)
	// Checksum at 10:12 left 0 for simplicity
	copy(buf[ipStart+12:ipStart+16], srcIP[:])
	copy(buf[ipStart+16:ipStart+20], dstIP[:])

	// TCP Header
	tcpStart := 34
	binary.BigEndian.PutUint16(buf[tcpStart:tcpStart+2], srcPort)
	binary.BigEndian.PutUint16(buf[tcpStart+2:tcpStart+4], dstPort)
	binary.BigEndian.PutUint32(buf[tcpStart+4:tcpStart+8], 1000) // Seq
	binary.BigEndian.PutUint32(buf[tcpStart+8:tcpStart+12], 0)   // Ack
	buf[tcpStart+12] = 0x50                                      // Data offset 5
	buf[tcpStart+13] = 0x02                                      // SYN flag
	binary.BigEndian.PutUint16(buf[tcpStart+14:tcpStart+16], 65535)

	if len(payload) > 0 {
		copy(buf[54:], payload)
	}

	return buf
}

// build: feat(pcap): implement Wireshar
// build: feat(pcap): add synthetic IPv4
// build: feat(cli): introduce pcap subc
// build: feat(geo): add IP Geolocation 
// build: feat(geo): integrate Autonomou
// build: feat(cli): add geo subcommand 
// build: feat(bgp): add BGP routing tab
// build: feat(bgp): parse upstream tran
// build: feat(cli): add bgp subcommand 
// build: feat(arp): implement cross-pla
// build: feat(arp): add MAC address val
// build: feat(cli): add arp subcommand 
// build: feat(http2): implement RFC 754
// build: feat(http2): add stream header
// build: feat(cli): add http2 subcomman
// build: feat(dnssec): add DNSSEC chain
// build: feat(dnssec): inspect RRSIG, D
// build: feat(cli): add dnssec subcomma
// build: feat(pcap): add PCAP packet re
// build: feat(storage): implement SQLit