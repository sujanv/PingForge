package netprobe

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/sujanv/pingforge/internal/config"
	"github.com/sujanv/pingforge/internal/core"
)

// PublicResolvers map human names to address:port
var PublicResolvers = map[string]string{
	"Cloudflare": "1.1.1.1:53",
	"Google":     "8.8.8.8:53",
	"Quad9":      "9.9.9.9:53",
	"OpenDNS":    "208.67.222.222:53",
}

// ProbeDNS performs multi-resolver and multi-record DNS diagnostics.
func ProbeDNS(ctx context.Context, domain string, cfg config.DNSConfig) *core.ProbeResult {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 3 * time.Second
	}

	res := &core.ProbeResult{
		ID:        fmt.Sprintf("dns-%s-%d", domain, time.Now().UnixNano()),
		Target:    core.Target{Host: domain, Protocol: "dns", Timeout: cfg.Timeout},
		Type:      core.ProbeDNS,
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	start := time.Now()
	dnsInfo := QueryResolver(ctx, domain, "system", cfg.RecordTypes, cfg.Timeout)
	totalLatency := time.Since(start)

	res.Latency = totalLatency
	res.LatencyMs = float64(totalLatency.Microseconds()) / 1000.0
	res.DNS = dnsInfo

	if dnsInfo.Error != "" {
		res.Status = core.StatusDown
		res.Error = dnsInfo.Error
		res.Message = fmt.Sprintf("DNS lookup failed: %s", dnsInfo.Error)
		return res
	}

	res.Status = core.StatusUp
	aRecords := dnsInfo.Records["A"]
	res.Message = fmt.Sprintf("Resolved %s -> %v in %.2fms", domain, aRecords, dnsInfo.QueryTimeMs)

	// Multi-resolver comparison
	resolverTimings := make(map[string]float64)
	for name, addr := range PublicResolvers {
		subInfo := QueryResolver(ctx, domain, addr, []string{"A"}, cfg.Timeout)
		if subInfo.Error == "" {
			resolverTimings[name] = subInfo.QueryTimeMs
		}
	}
	res.Details["resolver_latency_ms"] = resolverTimings

	return res
}

// QueryResolver queries a specific resolver for specified record types.
func QueryResolver(ctx context.Context, domain, resolverAddr string, recordTypes []string, timeout time.Duration) *core.DNSInfo {
	info := &core.DNSInfo{
		Resolver: resolverAddr,
		Records:  make(map[string][]string),
	}

	var resolver *net.Resolver
	if resolverAddr == "" || resolverAddr == "system" {
		resolver = net.DefaultResolver
	} else {
		resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: timeout}
				return d.DialContext(ctx, "udp", resolverAddr)
			},
		}
	}

	qStart := time.Now()
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if len(recordTypes) == 0 {
		recordTypes = []string{"A", "AAAA", "CNAME", "MX", "TXT"}
	}

	var firstErr error

	for _, recType := range recordTypes {
		switch strings.ToUpper(recType) {
		case "A", "AAAA":
			ips, err := resolver.LookupIPAddr(timeoutCtx, domain)
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			for _, ip := range ips {
				if ip.IP.To4() != nil && strings.ToUpper(recType) == "A" {
					info.Records["A"] = append(info.Records["A"], ip.IP.String())
				} else if ip.IP.To4() == nil && strings.ToUpper(recType) == "AAAA" {
					info.Records["AAAA"] = append(info.Records["AAAA"], ip.IP.String())
				}
			}

		case "CNAME":
			cname, err := resolver.LookupCNAME(timeoutCtx, domain)
			if err == nil && cname != "" && cname != domain+"." {
				info.Records["CNAME"] = []string{cname}
			}

		case "MX":
			mxs, err := resolver.LookupMX(timeoutCtx, domain)
			if err == nil {
				for _, mx := range mxs {
					info.Records["MX"] = append(info.Records["MX"], fmt.Sprintf("%s (pri %d)", mx.Host, mx.Pref))
				}
			}

		case "TXT":
			txts, err := resolver.LookupTXT(timeoutCtx, domain)
			if err == nil && len(txts) > 0 {
				info.Records["TXT"] = txts
			}

		case "NS":
			nss, err := resolver.LookupNS(timeoutCtx, domain)
			if err == nil {
				for _, ns := range nss {
					info.Records["NS"] = append(info.Records["NS"], ns.Host)
				}
			}
		}
	}

	info.QueryTimeMs = float64(time.Since(qStart).Microseconds()) / 1000.0

	if len(info.Records) == 0 && firstErr != nil {
		info.Error = firstErr.Error()
	}

	return info
}

// VerifyPropagation checks DNS resolution consistency across global resolvers concurrently.
func VerifyPropagation(ctx context.Context, domain string, timeout time.Duration) map[string]*core.DNSInfo {
	results := make(map[string]*core.DNSInfo)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for name, addr := range PublicResolvers {
		wg.Add(1)
		go func(rName, rAddr string) {
			defer wg.Done()
			info := QueryResolver(ctx, domain, rAddr, []string{"A", "CNAME"}, timeout)
			mu.Lock()
			results[rName] = info
			mu.Unlock()
		}(name, addr)
	}

	wg.Wait()
	return results
}
