package netprobe

import (
	"time"
	"github.com/sujanv/pingforge/internal/config"
)

// DefaultTCPConfig returns default tcp probe config.
func DefaultTCPConfig() config.TCPConfig {
	return config.TCPConfig{
		KeepAlive: 30 * time.Second,
		NoDelay:   true,
	}
}
