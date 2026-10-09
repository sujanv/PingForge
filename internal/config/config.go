package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config represents global configuration settings for PingForge.
type Config struct {
	Timeout      time.Duration `json:"timeout" yaml:"timeout"`
	Concurrency  int           `json:"concurrency" yaml:"concurrency"`
	OutputFormat string        `json:"output_format" yaml:"output_format"` // table, json, yaml, csv
	Verbose      bool          `json:"verbose" yaml:"verbose"`
	NoColor      bool          `json:"no_color" yaml:"no_color"`

	Storage StorageConfig `json:"storage" yaml:"storage"`
	TCP     TCPConfig     `json:"tcp" yaml:"tcp"`
	DNS     DNSConfig     `json:"dns" yaml:"dns"`
	Ping    PingConfig    `json:"ping" yaml:"ping"`
	Health  HealthConfig  `json:"health" yaml:"health"`
	Alerts  AlertConfig   `json:"alerts" yaml:"alerts"`
}

// StorageConfig holds options for persistent historical probe reports.
type StorageConfig struct {
	Enabled        bool   `json:"enabled" yaml:"enabled"`
	FilePath       string `json:"file_path" yaml:"file_path"`
	RetentionDays  int    `json:"retention_days" yaml:"retention_days"`
	MaxRecordsFile int    `json:"max_records" yaml:"max_records"`
}

// TCPConfig holds default options for TCP connectivity checks.
type TCPConfig struct {
	KeepAlive  time.Duration `json:"keep_alive" yaml:"keep_alive"`
	Retries    int           `json:"retries" yaml:"retries"`
	ReadBanner bool          `json:"read_banner" yaml:"read_banner"`
	TLS        bool          `json:"tls" yaml:"tls"`
	NoDelay    bool          `json:"no_delay" yaml:"no_delay"`
}

// DNSConfig holds default options for DNS resolution diagnostics.
type DNSConfig struct {
	Resolvers   []string      `json:"resolvers" yaml:"resolvers"`
	Timeout     time.Duration `json:"timeout" yaml:"timeout"`
	RecordTypes []string      `json:"record_types" yaml:"record_types"`
}

// PingConfig holds options for latency and jitter measurement.
type PingConfig struct {
	Count    int           `json:"count" yaml:"count"`
	Interval time.Duration `json:"interval" yaml:"interval"`
	Port     int           `json:"port" yaml:"port"`
}

// HealthConfig holds options for HTTP/HTTPS endpoint health probing.
type HealthConfig struct {
	FollowRedirects bool          `json:"follow_redirects" yaml:"follow_redirects"`
	MaxRedirects    int           `json:"max_redirects" yaml:"max_redirects"`
	ExpectedStatus  []int         `json:"expected_status" yaml:"expected_status"`
	UserAgent       string        `json:"user_agent" yaml:"user_agent"`
	LatencyBudget   time.Duration `json:"latency_budget" yaml:"latency_budget"`
}

// AlertConfig holds thresholds for latency and failure alerts.
type AlertConfig struct {
	Enabled            bool    `json:"enabled" yaml:"enabled"`
	LatencyThresholdMs float64 `json:"latency_threshold_ms" yaml:"latency_threshold_ms"`
	FailureThreshold   int     `json:"failure_threshold" yaml:"failure_threshold"`
	WebhookURL         string  `json:"webhook_url" yaml:"webhook_url"`
}

// DefaultConfig returns safe production defaults.
func DefaultConfig() *Config {
	defaultStorage := "history.json"

	return &Config{
		Timeout:      5 * time.Second,
		Concurrency:  10,
		OutputFormat: "table",
		Verbose:      false,
		NoColor:      false,
		Storage: StorageConfig{
			Enabled:        true,
			FilePath:       defaultStorage,
			RetentionDays:  30,
			MaxRecordsFile: 50000,
		},
		TCP: TCPConfig{
			KeepAlive:  30 * time.Second,
			Retries:    1,
			ReadBanner: false,
			TLS:        false,
			NoDelay:    true,
		},
		DNS: DNSConfig{
			Resolvers: []string{
				"system",
				"1.1.1.1:53", // Cloudflare
				"8.8.8.8:53", // Google
				"9.9.9.9:53", // Quad9
			},
			Timeout:     3 * time.Second,
			RecordTypes: []string{"A", "AAAA", "CNAME", "MX", "TXT", "NS"},
		},
		Ping: PingConfig{
			Count:    5,
			Interval: 200 * time.Millisecond,
			Port:     80,
		},
		Health: HealthConfig{
			FollowRedirects: true,
			MaxRedirects:    5,
			ExpectedStatus:  []int{200, 201, 204},
			UserAgent:       "PingForge/1.0 (+https://github.com/sujanv/PingForge)",
			LatencyBudget:   1500 * time.Millisecond,
		},
		Alerts: AlertConfig{
			Enabled:            false,
			LatencyThresholdMs: 250.0,
			FailureThreshold:   3,
			WebhookURL:         "",
		},
	}
}

// LoadConfig attempts to read config from the given file, falling back to defaults.
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()
	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return cfg, nil
}

// SaveConfig writes the configuration to disk.
func SaveConfig(cfg *Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
