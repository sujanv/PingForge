# ⚡ PingForge

> **Advanced Network Diagnostics Toolkit** for high-throughput TCP connectivity checks, multi-resolver DNS lookups, microsecond latency & RFC 3550 jitter measurement, HTTP/HTTPS endpoint health auditing, and historical SLA reports.

[![CI Status](https://github.com/sujanv/PingForge/actions/workflows/ci.yml/badge.svg)](https://github.com/sujanv/PingForge/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/sujanv/pingforge)](https://goreportcard.com/report/github.com/sujanv/pingforge)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)](https://golang.org)

---

## 🚀 Overview

**PingForge** is an enterprise-grade CLI network engineering toolkit designed for DevOps, SREs, and network engineers. Engineered in Go with modern socket programming, timeout-bounded contexts, worker pools, and zero-allocation statistical calculators.

### ✨ Key Capabilities

- **🔌 TCP Handshake & Socket Diagnostics**: High-precision connect latency, TCP keep-alive, TCP_NODELAY tuning, TLS handshake verification, certificate expiration auditing, and banner grabbing.
- **🌐 Multi-Resolver DNS Auditing**: Query system and public DNS resolvers (Cloudflare `1.1.1.1`, Google `8.8.8.8`, Quad9 `9.9.9.9`, OpenDNS) in parallel across `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `NS`, and `SRV` records, with global propagation checking.
- **⏱️ Latency & RFC 3550 Jitter Statistics**: Microsecond round-trip time (RTT) measurements, packet loss tracking, standard deviation, RFC 3550 interarrival jitter, and percentiles (`P50`, `P90`, `P95`, `P99`).
- **🩺 HTTP/HTTPS Health Checks**: Granular timing breakdown (DNS lookup, TCP connect, TLS handshake, TTFB, total), regex body assertions, HTTP status validation, and latency budgets.
- **⚡ Concurrent Worker Pool & CIDR Probing**: Batch scanning across target lists or IPv4 CIDR blocks (`192.168.1.0/28`) with worker pools, rate limiting, and graceful cancellation.
- **📊 Historical Latency SLA Reports**: Persistent JSON time-series store, terminal ASCII sparklines (` ▂▃▄▅▆▇█`) and distribution histograms, CSV export, and standalone responsive HTML dashboards.
- **📡 Background Monitoring Daemon & Prometheus Exporter**: Continuous background probing daemon exposing native `/metrics` (Prometheus) and `/healthz` endpoints.

---

## 🏗️ Architecture

```
                                 ┌──────────────────────────────────┐
                                 │       PingForge CLI Engine       │
                                 │         (Cobra + Viper)          │
                                 └──────────────┬───────────────────┘
                                                │
                 ┌──────────────────────────────┼──────────────────────────────┐
                 ▼                              ▼                              ▼
      ┌────────────────────┐         ┌────────────────────┐         ┌────────────────────┐
      │   Worker Pool      │         │   Storage Engine   │         │  Output Formatter  │
      │  (Channels, sync,  │         │(Append JSON Ledger,│         │(Table, ANSI Color, │
      │   Rate Limiting)   │         │ Uptime SLA, Trends)│         │ JSON, YAML, HTML)  │
      └──────────┬─────────┘         └──────────┬─────────┘         └──────────┬─────────┘
                 │                              │                              │
     ┌───────────┴───────────┐                  │                              │
     ▼                       ▼                  │                              │
┌──────────────┐      ┌──────────────┐          │                              │
│  net.Dialer  │      │ net.Resolver │          ▼                              ▼
│  TCP/TLS/RTT │      │ DNS/Records  │   ┌───────────────┐              ┌───────────────┐
└──────────────┘      └──────────────┘   │ HTML / ASCII  │              │  /metrics     │
     ▲                       ▲           │ Latency Graph │              │ Prometheus    │
     │                       │           └───────────────┘              └───────────────┘
     └───────────┬───────────┘
                 │
      ┌──────────┴──────────┐
      │  httptrace / HTTP   │
      │  DNS, TCP, TLS, TTFB│
      └─────────────────────┘
```

---

## 📦 Installation

### From Source
```bash
go install github.com/sujanv/pingforge/cmd/pingforge@latest
```

### Build Locally
```bash
git clone https://github.com/sujanv/PingForge.git
cd PingForge
make build
./bin/pingforge --help
```

---

## 🛠️ Usage & Commands

### 1. TCP Connectivity & TLS Handshake
```bash
# Basic TCP connection check
pingforge tcp 1.1.1.1:443

# TCP check with TLS negotiation and certificate inspection
pingforge tcp google.com:443 --tls

# Grab service banner (e.g. SSH, Redis, FTP)
pingforge tcp github.com:22 --banner
```

### 2. DNS Queries & Global Propagation
```bash
# Query system and public resolvers
pingforge dns google.com

# Query specific resolver and record types
pingforge dns github.com -r 1.1.1.1:53 --type A,AAAA,MX,TXT

# Check global DNS propagation across worldwide resolvers
pingforge dns cloudflare.com --propagation
```

### 3. Latency & RFC 3550 Jitter Measurement
```bash
# Ping target with 5 probes
pingforge ping 8.8.8.8 -c 5

# High-frequency jitter & percentile probe
pingforge ping 1.1.1.1 -c 10 -i 100ms -p 443
```

### 4. HTTP/HTTPS Endpoint Health Check
```bash
# Basic health check
pingforge health https://httpbin.org/status/200

# Health check with status validation, body regex match, and latency budget
pingforge health https://httpbin.org/json --status 200 --match '"slideshow"' --budget 500ms
```

### 5. Concurrent Port Scanner
```bash
# Fast port scan for common services
pingforge scan 127.0.0.1

# Scan custom port range with banner grabbing
pingforge scan 127.0.0.1 -p 80,443,8000-8080 --banner
```

### 6. Batch Probing & CIDR Expansion
```bash
# Probe targets from file in parallel
pingforge batch -f examples/targets.txt --type TCP

# Scan CIDR subnet concurrently
pingforge batch --cidr 192.168.1.0/28 --type TCP
```

### 7. Historical SLA & Latency Reports
```bash
# Terminal report with ASCII sparklines
pingforge report --since 24h

# Export interactive standalone HTML report
pingforge report --format html --out report.html

# Export to CSV or JSON
pingforge report --format csv --out report.csv
pingforge report --format json
```

### 8. Prometheus Daemon Mode
```bash
# Launch background monitoring daemon
pingforge daemon -l :9090 -i 15s -f examples/targets.txt
# Metrics available at: http://localhost:9090/metrics
# Health status at: http://localhost:9090/healthz
```

---

## 📋 Output Formats

PingForge supports first-class machine-readable formats for CI/CD pipelines:

```bash
# JSON output
pingforge tcp 1.1.1.1:443 --json

# YAML output
pingforge health https://github.com --yaml
```

---

## 🧪 Testing & Benchmarks

Run unit tests and race detection:
```bash
make test
```

Run benchmarks:
```bash
make benchmark
```

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
