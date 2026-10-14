# ⚡ PingForge

> **Advanced Network Diagnostics & Telemetry Toolkit** — High-throughput TCP/UDP connectivity analysis, multi-resolver DNS/DoH/DoT auditing, microsecond latency & RFC 3550 jitter measurement, SSL/TLS security scorecards, WebSocket testing, Path MTU discovery, interactive live TUI dashboards, and historical SLA reports.

[![CI Status](https://github.com/sujanv/PingForge/actions/workflows/ci.yml/badge.svg)](https://github.com/sujanv/PingForge/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/sujanv/pingforge)](https://goreportcard.com/report/github.com/sujanv/pingforge)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)](https://golang.org)

---

## 🚀 Overview

**PingForge** is an enterprise-grade CLI network engineering toolkit designed for SREs, DevOps engineers, and network architects. Engineered in Go with modern socket programming, timeout-bounded contexts, worker pools, and zero-allocation statistical calculators.

### ✨ Comprehensive Feature Matrix

| Feature | Subcommand | Key Capabilities |
| :--- | :--- | :--- |
| **TCP Diagnostics** | `pingforge tcp` | Handshake latency, Keep-Alive, TCP_NODELAY, TLS negotiation, banner grabbing |
| **DNS Resolution** | `pingforge dns` | Multi-resolver lookups (Cloudflare, Google, Quad9, OpenDNS), A/AAAA/MX/TXT/NS/CNAME/SRV, propagation checks |
| **DNS-over-HTTPS** | `pingforge doh` | RFC 8484 wire format & JSON API, query latency phase timing (DNS, TCP, TLS, TTFB) |
| **DNS-over-TLS** | `pingforge dot` | RFC 7858 on port 853, TLS handshake timing, wire protocol parsing |
| **Latency & Jitter** | `pingforge ping` | Round-trip time (RTT), packet loss, sample standard deviation, **RFC 3550 Jitter**, P50/P90/P95/P99 percentiles |
| **Live TUI Monitor** | `pingforge live` | Full-screen real-time terminal dashboard, rolling packet loss gauge, live ASCII sparklines |
| **TLS Security Audit** | `pingforge tls-audit` | Multi-version inspection (TLS 1.0-1.3), cipher suite audit, cert chain hierarchy, ALPN, security grade (A+ to F) |
| **HTTP/HTTPS Health** | `pingforge health` | Granular phase timing via `httptrace`, regex assertions, HTTP status validation, latency budgets |
| **HTTP/2 Multiplexing** | `pingforge http2` | RFC 7540 HTTP/2 probe, stream header analysis, ALPN verification, multiplex latency |
| **HTTP/3 & QUIC Probe** | `pingforge http3` | RFC 9114 HTTP/3 probe, Alt-Svc advertisement verification, UDP 443 handshake RTT |
| **Continuous MTR Trace** | `pingforge mtr`   | Multi-cycle traceroute engine, statistical hop-by-hop packet loss, jitter standard deviation |
| **OCSP Stapling Audit** | `pingforge ocsp`  | RFC 6066 OCSP response verification, certificate revocation inspection, CRL distributions |
| **Full-Mesh Latency**   | `pingforge mesh`  | Concurrent $N \times N$ peer latency mesh builder for multi-datacenter network profiling |
| **NTP Time Sync**       | `pingforge ntp`   | RFC 5905 Network Time Protocol, stratum, clock offset, round-trip delay, and server dispersion |
| **Bufferbloat & Loaded Latency** | `pingforge bufferbloat` | Baseline vs concurrent loaded latency inflation, bufferbloat grading (A+ to F) |
| **WHOIS & RDAP Resolver**| `pingforge whois` | RFC 3912 TCP port 43 WHOIS queries, RDAP REST JSON fallback, registrar and expiry discovery |
| **SNMP v2c Device Walk**| `pingforge snmp`  | Lightweight BER ASN.1 encoder, MIB-II system status (sysDescr, sysUpTime, sysName) |
| **Dual-Stack Happy Eyeballs** | `pingforge dualstack` | Concurrent RFC 8305 IPv4 vs IPv6 race, latency delta comparator, protocol winner detection |
| **Bulk TLS Expiry Auditor** | `pingforge cert-check` | Batch SSL/TLS certificate validity scanner, expiration warnings, CRITICAL/WARNING/OK alerting |
| **mDNS Service Discovery** | `pingforge mdns` | Multicast DNS ZeroConf discovery on 224.0.0.251:5353, local LAN hostnames and services |
| **OpenTelemetry & InfluxDB** | `pingforge report` | RFC-compliant OpenTelemetry JSON Metrics & InfluxDB 2.x Line Protocol exporters |
| **WebSocket Probe** | `pingforge ws` | RFC 6455 HTTP upgrade handshake timing, Ping/Pong frame RTT, subprotocol validation |
| **DNSSEC Validation** | `pingforge dnssec` | Cryptographic signature verification, Authenticated Data (AD), RRSIG, DNSKEY, DS records |
| **IP Geolocation & ASN**| `pingforge geo` | IP Geolocation, ASN, organization, coordinates, timezone, and network intelligence |
| **BGP Route Visibility**| `pingforge bgp` | BGP route status, origin ASN, route visibility %, upstream transit provider analysis |
| **ARP & Neighbor Table**| `pingforge arp` | Cross-platform local ARP and IPv4 Neighbor cache scanner, MAC address resolution |
| **Packet Capture (PCAP)**| `pingforge pcap` | Stream synthetic network probes into Wireshark-compatible PCAP files or replay recorded packets |
| **UDP Reachability** | `pingforge udp` | Custom payload transmission, response latency, port unreachable handling |
| **Path MTU Discovery** | `pingforge mtu` | Socket buffer threshold scanning (576 to 9000 bytes), estimated TCP MSS detection |
| **Throughput Speed** | `pingforge throughput` | Sustained TCP streaming & HTTP download speed measurement in MB/s and Mbps |
| **Port Scanner** | `pingforge scan` | High-concurrency worker pool port scanner, service detection, banner grabbing |
| **Route Tracer** | `pingforge trace` | Hop-by-hop latency and route inspection |
| **Batch Probing** | `pingforge batch` | Parallel multi-target scanning via files (`-f targets.txt`) or IPv4 CIDR blocks (`--cidr 192.168.1.0/28`) |
| **Historical Reports** | `pingforge report` | Persistent time-series storage, SLA uptime %, ASCII sparklines, CSV/JSON/HTML/PostScript |
| **Anomaly Detection** | `pingforge report` | Z-score statistical outlier detection ($z \ge 2.5$) and latency spike alerting |
| **Monitoring Daemon** | `pingforge daemon` | Background monitoring loop, native Prometheus `/metrics` exporter, and `/healthz` endpoints |

---

## 🏗️ Architecture

```
                                  ┌─────────────────────────────────────┐
                                  │        PingForge CLI Engine         │
                                  │           (Cobra + Viper)           │
                                  └──────────────────┬──────────────────┘
                                                     │
               ┌─────────────────────────────────────┼─────────────────────────────────────┐
               ▼                                     ▼                                     ▼
     ┌────────────────────┐                ┌────────────────────┐                ┌────────────────────┐
     │   Worker Pool      │                │   Storage Engine   │                │  Output Formatter  │
     │  (Channels, sync,  │                │(Append JSON Ledger,│                │(ANSI Tables, JSON, │
     │   Rate Limiting)   │                │ SLA, Z-Score Anom) │                │ YAML, NDJSON, GFM) │
     └─────────┬──────────┘                └─────────┬──────────┘                └─────────┬──────────┘
               │                                     │                                     │
   ┌───────────┴───────────┐                         │                                     │
   ▼                       ▼                         │                                     │
┌──────────────┐    ┌──────────────┐                 ▼                                     ▼
│  net.Dialer  │    │ DoH / DoT /  │          ┌───────────────┐                     ┌───────────────┐
│ TCP/UDP/Ping │    │ DNS Resolver │          │ HTML / Live   │                     │  /metrics     │
│  MTU / Speed │    │ 53 & 853     │          │ TUI Dashboard │                     │ Prometheus    │
└──────────────┘    └──────────────┘          └───────────────┘                     └───────────────┘
   ▲                       ▲
   │                       │
   └───────────┬───────────┘
               │
    ┌──────────┴──────────┐
    │  TLS & WebSockets   │
    │  Handshake, Ciphers,│
    │  Ping/Pong Frames   │
    └──────────┬──────────┘
               │
    ┌──────────┴──────────────────────────┐
    │  v5 Advanced Telemetry Engines      │
    │  • NTP (RFC 5905) Stratum / Jitter  │
    │  • Bufferbloat Latency Inflation    │
    │  • SNMP v2c MIB-II Network Walker   │
    │  • Dual-Stack (RFC 8305) Comparator │
    │  • Bulk TLS Expiry & Alert Monitor  │
    │  • Multicast DNS (mDNS) ZeroConf    │
    │  • InfluxDB & OTLP Metric Exporters │
    └─────────────────────────────────────┘
```

---

## 📦 Installation

### Pre-requisites
- Go 1.22 or higher (`go version`)

### Build from Source
```bash
git clone https://github.com/sujanv/PingForge.git
cd PingForge
make build
./bin/pingforge --help
```

### Install Globally
```bash
make install
# or
go install github.com/sujanv/pingforge/cmd/pingforge@latest
```

---

## 🛠️ Detailed CLI Usage & Command Reference

### 1. TCP Connectivity & Socket Tuning (`pingforge tcp`)
```bash
# Basic TCP handshake check
pingforge tcp 1.1.1.1:443

# Handshake with TLS certificate inspection
pingforge tcp google.com:443 --tls

# Grab service banner (e.g. SSH, FTP, Redis)
pingforge tcp github.com:22 --banner
```

### 2. DNS-over-HTTPS (DoH) Queries (`pingforge doh`)
Query encrypted DNS-over-HTTPS providers with granular phase timing (DNS, TCP, TLS, TTFB):
```bash
# Query via Cloudflare (default)
pingforge doh google.com

# Query via Google DoH with MX records
pingforge doh github.com --provider google --type MX

# Query via Quad9 or AdGuard
pingforge doh wikipedia.org --provider quad9 --type AAAA
```

### 3. DNS-over-TLS (DoT) Queries (`pingforge dot`)
Query RFC 7858 DNS-over-TLS on port 853:
```bash
# Query Cloudflare DoT (1.1.1.1:853)
pingforge dot google.com

# Query Quad9 DoT (9.9.9.9:853)
pingforge dot cloudflare.com --resolver 9.9.9.9:853
```

### 4. SSL/TLS Security Audit (`pingforge tls-audit`)
Performs a comprehensive security evaluation, checking for deprecated protocol versions (TLS 1.0, 1.1), certificate chains, ALPN negotiation, and security scorecard:
```bash
# Audit TLS configuration on port 443
pingforge tls-audit google.com:443

# Output audit as machine-readable JSON
pingforge tls-audit github.com:443 --json
```

### 5. Latency, Jitter & Percentiles (`pingforge ping`)
Rapid consecutive connection probes with statistical analysis:
```bash
# 5 consecutive probes to Google DNS
pingforge ping 8.8.8.8 -p 53 -c 5

# High-frequency jitter & percentile probe
pingforge ping 1.1.1.1 -p 443 -c 10 -i 100ms
```

### 6. Live Interactive Terminal Monitor (`pingforge live`)
Full-screen refreshing terminal dashboard with real-time rolling sparkline latency graph:
```bash
# Live monitor 1.1.1.1 every 500ms
pingforge live 1.1.1.1:443 -i 500ms
```

### 7. WebSocket Handshake & Ping-Pong RTT (`pingforge ws`)
Validates RFC 6455 WebSocket connectivity:
```bash
# Test WebSocket endpoint
pingforge ws wss://echo.websocket.events

# Test with custom subprotocols
pingforge ws wss://echo.websocket.events --subprotocol chat,superchat
```

### 8. UDP Reachability Probe (`pingforge udp`)
Sends UDP probe packets with custom payloads:
```bash
# Send UDP probe to DNS port
pingforge udp 8.8.8.8:53 --payload "HELLO-DNS"
```

### 9. Path MTU Discovery (`pingforge mtu`)
Discovers optimal packet transmission sizes without fragmentation:
```bash
# Probe path MTU
pingforge mtu 1.1.1.1
```

### 10. Throughput & Download Speed (`pingforge throughput`)
Measures sustained download bandwidth:
```bash
# Benchmark throughput over 3 seconds
pingforge throughput https://www.google.com -d 3s
```

### 11. Concurrent Port Scanner (`pingforge scan`)
Fast multi-threaded port scanner with service identification:
```bash
# Scan default common ports
pingforge scan 127.0.0.1

# Scan custom port range with banner grabbing
pingforge scan 127.0.0.1 -p 80,443,8000-8080 --banner
```

### 12. Batch Probing & CIDR Expansion (`pingforge batch`)
Concurrent diagnostics across targets or entire subnets:
```bash
# Batch probe targets from text file
pingforge batch -f examples/targets.txt --type TCP

# Scan IPv4 CIDR subnet in parallel
pingforge batch --cidr 192.168.1.0/28 --type TCP
```

### 13. Historical SLA Reports & Anomaly Detection (`pingforge report`)
Generate comprehensive uptime and latency SLA reports:
```bash
# Terminal report with ASCII sparklines
pingforge report --since 24h

# Export interactive standalone HTML report
pingforge report --format html --out report.html

# Export as CSV or JSON
pingforge report --format csv --out report.csv
pingforge report --format json --out report.json
```

### 14. Prometheus Monitoring Daemon (`pingforge daemon`)
Continuous background probing with native metrics exposition:
```bash
# Run daemon polling every 15 seconds
pingforge daemon -l :9090 -i 15s -f examples/targets.txt
```
Endpoints exposed:
- `http://localhost:9090/metrics` — Prometheus metrics (`pingforge_probe_duration_seconds`, `pingforge_probe_success`, `pingforge_uptime_ratio`)
- `http://localhost:9090/healthz` — Liveness health probe
- `http://localhost:9090/api/summary` — JSON historical summary

### 15. IP Geolocation & ASN Intelligence (`pingforge geo`)
Lookup geographic coordinates, city, region, Autonomous System (ASN), and ISP details:
```bash
# Lookup IP geolocation and ASN
pingforge geo 8.8.8.8

# JSON output for automated scripting
pingforge geo 1.1.1.1 --json
```

### 16. BGP Route Visibility & Origin ASN (`pingforge bgp`)
Inspect global Border Gateway Protocol (BGP) routing status, prefix visibility, and upstream transit providers:
```bash
# Check BGP route visibility for an IPv4 address
pingforge bgp 1.1.1.1

# Query by Autonomous System Number (ASN)
pingforge bgp AS13335
```

### 17. Local ARP / Neighbor Cache (`pingforge arp`)
Discover active network hardware, MAC addresses, and network interface bindings on the local subnet:
```bash
# Display system ARP cache table
pingforge arp

# Output as JSON
pingforge arp --json
```

### 18. HTTP/2 Multiplexing & ALPN Probe (`pingforge http2`)
Validate HTTP/2 binary framing, Application-Layer Protocol Negotiation (ALPN), and stream response headers:
```bash
# Probe HTTP/2 protocol support and ALPN
pingforge http2 https://www.google.com

# Inspect HTTP/2 stream headers
pingforge http2 https://cloudflare.com --json
```

### 19. DNSSEC Cryptographic Chain Validation (`pingforge dnssec`)
Verify DNSSEC cryptographic trust chains, Authenticated Data (AD) flags, RRSIG, and DNSKEY records:
```bash
# Validate domain DNSSEC signatures
pingforge dnssec cloudflare.com

# Check signed domain
pingforge dnssec iana.org
```

### 20. Wireshark-Compatible Packet Capture (`pingforge pcap`)
Stream synthetic diagnostic probes directly into standard `.pcap` captures for analysis in Wireshark or tcpdump:
```bash
# Capture 25 synthetic probe frames
pingforge pcap trace.pcap -c 25

# Inspect captured packets in Wireshark
wireshark trace.pcap
```

### 21. HTTP/3 (QUIC) Capability & Alt-Svc Audit (`pingforge http3`)
Verify RFC 9114 HTTP/3 protocol support, Alt-Svc advertisement headers, and UDP 443 handshake RTT:
```bash
# Audit HTTP/3 support on endpoint
pingforge http3 https://www.google.com

# Inspect draft protocol version and Alt-Svc headers
pingforge http3 https://cloudflare.com --json
```

### 22. Continuous MTR Diagnostics (`pingforge mtr`)
Execute multi-cycle traceroutes measuring hop-by-hop packet loss, average latency, and jitter standard deviation:
```bash
# Run 5 traceroute cycles to target host
pingforge mtr 1.1.1.1 --cycles 5

# Trace path to specific TCP port
pingforge mtr github.com -p 443 --cycles 3
```

### 23. SSL/TLS OCSP Stapling & Revocation (`pingforge ocsp`)
Inspect RFC 6066 Online Certificate Status Protocol (OCSP) stapled responses and CRL distribution points:
```bash
# Verify OCSP stapling on HTTPS endpoint
pingforge ocsp google.com

# Audit custom port
pingforge ocsp mydomain.com -p 8443
```

### 24. Full-Mesh Peer Latency Matrix (`pingforge mesh`)
Run concurrent pairwise latency diagnostics across multi-node peer networks to identify routing asymmetries:
```bash
# Profile full-mesh latency between nodes
pingforge mesh 10.0.0.1,10.0.0.2,10.0.0.3

# Output mesh matrix as JSON
pingforge mesh node1.internal,node2.internal,node3.internal --json
```

### 25. RFC 5905 Network Time Protocol (NTP) Diagnostics (`pingforge ntp`)
Query Stratum-1 and public NTP servers (pool.ntp.org, time.google.com, time.cloudflare.com) to measure clock offset, round-trip delay, and stratum dispersion:
```bash
# Query default NTP pool
pingforge ntp pool.ntp.org

# Query specific server with custom timeout and port
pingforge ntp time.google.com -p 123 -t 3s --json
```

### 26. Bufferbloat & Loaded Latency Grading (`pingforge bufferbloat`)
Measure latency inflation and packet buffering under heavy sustained download saturation, grading network buffer health from A+ to F:
```bash
# Run 10-second bufferbloat benchmark against HTTP target
pingforge bufferbloat http://speedtest.tele2.net/10MB.zip -d 10s

# Inspect loaded latency delta and bufferbloat grade
pingforge bufferbloat https://proof.ovh.net/files/10Mb.dat --concurrency 8 --json
```

### 27. RFC 3912 TCP WHOIS & RDAP Domain Auditing (`pingforge whois`)
Perform high-speed raw TCP WHOIS lookups on port 43 with automated fallback to ICANN Registration Data Access Protocol (RDAP) JSON APIs:
```bash
# Query domain WHOIS registration
pingforge whois github.com

# Query Autonomous System Number (ASN) details
pingforge whois AS15169 --json
```

### 28. SNMP v2c Network Equipment Telemetry (`pingforge snmp`)
Query network switches, routers, and edge firewalls via lightweight BER-encoded SNMP v2c UDP probes to extract MIB-II system description, uptime, and identity:
```bash
# Query local SNMP agent with default 'public' community
pingforge snmp 192.168.1.1 -c public

# Query specific OID on custom port
pingforge snmp router.lan:161 -c secret -o 1.3.6.1.2.1.1.1.0 --json
```

### 29. RFC 8305 Dual-Stack Happy Eyeballs Race (`pingforge dualstack`)
Execute simultaneous IPv4 (A) and IPv6 (AAAA) connection races with configurable connection head starts to diagnose IPv6 routing penalties and protocol preference:
```bash
# Compare IPv4 vs IPv6 connection latency
pingforge dualstack google.com -p 443

# Output detailed dual-stack delta and winning protocol
pingforge dualstack cloudflare.com -p 80 --json
```

### 30. Bulk TLS Certificate Expiration Auditor (`pingforge cert-check`)
Audit batches of HTTPS endpoints and internal services for upcoming TLS certificate expiration, classifying risk levels as CRITICAL, WARNING, or OK:
```bash
# Check single target with 30-day warning threshold
pingforge cert-check github.com:443 --days 30

# Bulk audit targets from list
pingforge cert-check -f examples/targets.txt --days 14 --json
```

### 31. Multicast DNS (mDNS) ZeroConf Discovery (`pingforge mdns`)
Send UDP multicast queries across the local LAN on `224.0.0.251:5353` to discover ZeroConf services, Apple Bonjour devices, and local printers:
```bash
# Discover local workstation services
pingforge mdns _workstation._tcp.local.

# Discover local HTTP/HTTPS services
pingforge mdns _http._tcp.local. -t 3s --json
```

---

## 📋 Output Formats

PingForge supports first-class machine-readable formats for CI/CD pipelines:

```bash
# JSON output
pingforge tcp 1.1.1.1:443 --json

# YAML output
pingforge health https://github.com --yaml

# Newline-Delimited JSON (NDJSON)
pingforge batch -f examples/targets.txt --ndjson

# GitHub Flavored Markdown Table
pingforge batch -f examples/targets.txt --format markdown
```

---

## 🧪 Testing & Benchmarks

Run the complete unit test suite with race detection:
```bash
make test
```

Run micro-benchmarks:
```bash
make benchmark
```

---


---

## ❓ Troubleshooting & FAQs

### Why does `pingforge bufferbloat` require an HTTP endpoint?
Bufferbloat measures latency degradation under saturated network throughput. PingForge streams concurrent payloads to an HTTP/HTTPS endpoint to artificially fill network router queues, continuously probing TCP latency in parallel to compute buffer inflation.

### Can `pingforge snmp` query Cisco, Juniper, and Arista switches?
Yes! PingForge implements standard RFC 1157 / RFC 1901 SNMP v2c GetRequest PDUs using BER encoding. It works out-of-the-box with any network appliance exposing standard MIB-II system OIDs (`1.3.6.1.2.1.1.x`).

### How does `pingforge dualstack` implement Happy Eyeballs RFC 8305?
It fires concurrent connection requests to both IPv4 (`A`) and IPv6 (`AAAA`) addresses of the target domain. An optional connection headstart (default: 50ms) is granted to IPv6 to prioritize next-generation networking without penalizing user experience if IPv6 routing is impaired.

## 📄 License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
