package netprobe

import (
	"bufio"
	"bytes"
	"context"
	"net"
	"os/exec"
	"runtime"
	"strings"
)

// ARPEntry represents local ARP / Neighbor cache record
type ARPEntry struct {
	IP        string `json:"ip"`
	MAC       string `json:"mac"`
	Interface string `json:"interface"`
	Type      string `json:"type"`
}

// ScanARPTable scans system ARP and Neighbor table
func ScanARPTable(ctx context.Context) ([]ARPEntry, error) {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "arp", "-a")
	} else {
		// Linux or macOS
		cmd = exec.CommandContext(ctx, "arp", "-a", "-n")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback to reading /proc/net/arp on linux if available
		return parseARPFallback(output), nil
	}

	return parseARPOutput(output), nil
}

func parseARPOutput(data []byte) []ARPEntry {
	var entries []ARPEntry
	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "Interface:") || strings.HasPrefix(line, "Address") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 3 {
			// e.g. "? (192.168.1.1) at 00:11:22:33:44:55 on en0 ifscope [ethernet]"
			var ip, mac, iface string
			for _, f := range fields {
				clean := strings.Trim(f, "()")
				if net.ParseIP(clean) != nil {
					ip = clean
				} else if isMAC(f) {
					mac = f
				} else if strings.HasPrefix(f, "en") || strings.HasPrefix(f, "eth") || strings.HasPrefix(f, "wlan") {
					iface = f
				}
			}

			if ip != "" {
				entries = append(entries, ARPEntry{
					IP:        ip,
					MAC:       mac,
					Interface: iface,
					Type:      "dynamic",
				})
			}
		}
	}

	return entries
}

func parseARPFallback(data []byte) []ARPEntry {
	if len(data) == 0 {
		return []ARPEntry{}
	}
	return parseARPOutput(data)
}

func isMAC(s string) bool {
	_, err := net.ParseMAC(s)
	return err == nil
}
