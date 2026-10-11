package netprobe

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

// PCAPReplayer reads raw PCAP frames and re-injects packets for network emulation
type PCAPReplayer struct {
	Path       string
	PacketCount int
}

// ReplayPCAP streams saved packets from a PCAP file to a target TCP/UDP socket
func ReplayPCAP(ctx context.Context, pcapFile string, targetAddr string) (int, error) {
	f, err := os.Open(pcapFile)
	if err != nil {
		return 0, fmt.Errorf("open pcap: %w", err)
	}
	defer f.Close()

	var gh PCAPGlobalHeader
	if err := binary.Read(f, binary.LittleEndian, &gh); err != nil {
		return 0, fmt.Errorf("read global header: %w", err)
	}

	conn, err := net.DialTimeout("udp", targetAddr, 2*time.Second)
	if err != nil {
		return 0, fmt.Errorf("dial target: %w", err)
	}
	defer conn.Close()

	count := 0
	for {
		select {
		case <-ctx.Done():
			return count, ctx.Err()
		default:
		}

		var ph PCAPPacketHeader
		if err := binary.Read(f, binary.LittleEndian, &ph); err != nil {
			if err == io.EOF {
				break
			}
			return count, err
		}

		data := make([]byte, ph.CaptureLength)
		if _, err := io.ReadFull(f, data); err != nil {
			break
		}

		_, _ = conn.Write(data)
		count++
	}

	return count, nil
}
