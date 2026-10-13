package netprobe

import (
	"context"
	"net"
	"time"
)

// MDNSService represents a discovered local ZeroConf / Bonjour service.
type MDNSService struct {
	Service   string `json:"service"`
	Instance  string `json:"instance"`
	Host      string `json:"host"`
	IPAddress string `json:"ip"`
	Port      int    `json:"port"`
}

// ScanMDNS broadcasts a DNS-SD query on 224.0.0.251:5353 to locate local devices.
func ScanMDNS(ctx context.Context, timeout time.Duration) ([]MDNSService, error) {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}

	addr, err := net.ResolveUDPAddr("udp4", "224.0.0.251:5353")
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp4", nil)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Simple DNS-SD query for _services._dns-sd._udp.local
	query := []byte{
		0x00, 0x00, // ID
		0x00, 0x00, // Flags
		0x00, 0x01, // Questions: 1
		0x00, 0x00, // Answers
		0x00, 0x00, // Authority
		0x00, 0x00, // Additional
		0x09, '_', 's', 'e', 'r', 'v', 'i', 'c', 'e', 's',
		0x07, '_', 'd', 'n', 's', '-', 's', 'd',
		0x04, '_', 'u', 'd', 'p',
		0x05, 'l', 'o', 'c', 'a', 'l',
		0x00,
		0x00, 0x0c, // Type PTR
		0x00, 0x01, // Class IN
	}

	_, _ = conn.WriteTo(query, addr)
	_ = conn.SetReadDeadline(time.Now().Add(timeout))

	var services []MDNSService
	buf := make([]byte, 2048)

	for {
		select {
		case <-ctx.Done():
			return services, nil
		default:
		}

		n, src, err := conn.ReadFrom(buf)
		if err != nil {
			break
		}

		if n > 12 {
			services = append(services, MDNSService{
				Service:   "_services._dns-sd._udp.local",
				Host:      src.String(),
				IPAddress: src.(*net.UDPAddr).IP.String(),
				Port:      src.(*net.UDPAddr).Port,
			})
		}
	}

	return services, nil
}
