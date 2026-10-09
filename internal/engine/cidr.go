package engine

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/sujanv/pingforge/internal/core"
)

// ExpandCIDR expands an IPv4 CIDR prefix into a list of usable host IP strings.
func ExpandCIDR(cidrStr string) ([]string, error) {
	ip, ipNet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR %q: %w", cidrStr, err)
	}

	ip4 := ip.To4()
	if ip4 == nil {
		return []string{ip.String()}, nil
	}

	mask := binary.BigEndian.Uint32(ipNet.Mask)
	start := binary.BigEndian.Uint32(ip4) & mask
	total := ^mask + 1

	// Limit to max 4096 hosts per single CIDR probe to prevent memory explosion
	if total > 4096 {
		return nil, fmt.Errorf("CIDR %s contains %d hosts; maximum permitted per batch is 4096", cidrStr, total)
	}

	var ips []string
	for i := uint32(0); i < total; i++ {
		cur := start + i
		curIP := make(net.IP, 4)
		binary.BigEndian.PutUint32(curIP, cur)

		// For /30 or larger, exclude network and broadcast
		if total > 2 && (i == 0 || i == total-1) {
			continue
		}
		ips = append(ips, curIP.String())
	}

	return ips, nil
}

// ParseTargetString parses a target input (URL, CIDR, host:port, or hostname).
func ParseTargetString(s string) ([]core.Target, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	// Check if URL
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		u, err := url.Parse(s)
		if err != nil {
			return nil, err
		}
		port := 80
		if u.Scheme == "https" {
			port = 443
		}
		if u.Port() != "" {
			if p, err := strconv.Atoi(u.Port()); err == nil {
				port = p
			}
		}
		return []core.Target{
			{
				Host:     u.Hostname(),
				Port:     port,
				URL:      s,
				Protocol: u.Scheme,
			},
		}, nil
	}

	// Check if CIDR
	if strings.Contains(s, "/") {
		ips, err := ExpandCIDR(s)
		if err != nil {
			return nil, err
		}
		var targets []core.Target
		for _, ip := range ips {
			targets = append(targets, core.Target{
				Host:     ip,
				Port:     80,
				Protocol: "tcp",
			})
		}
		return targets, nil
	}

	// Check if host:port
	if strings.Contains(s, ":") {
		host, portStr, err := net.SplitHostPort(s)
		if err == nil {
			port, _ := strconv.Atoi(portStr)
			return []core.Target{
				{
					Host:     host,
					Port:     port,
					Protocol: "tcp",
				},
			}, nil
		}
	}

	// Plain hostname/IP
	return []core.Target{
		{
			Host:     s,
			Port:     80,
			Protocol: "tcp",
		},
	}, nil
}

// ReadTargetsFile reads lines from a file and parses targets.
func ReadTargetsFile(path string) ([]core.Target, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var targets []core.Target
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parsed, err := ParseTargetString(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}
		targets = append(targets, parsed...)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return targets, nil
}
