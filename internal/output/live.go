package output

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/sujanv/pingforge/internal/core"
	"github.com/sujanv/pingforge/internal/netprobe"
)

// LiveMonitor renders real-time updating terminal diagnostics.
type LiveMonitor struct {
	w       io.Writer
	history []float64
}

// NewLiveMonitor creates a new live monitor.
func NewLiveMonitor() *LiveMonitor {
	return &LiveMonitor{
		w:       os.Stdout,
		history: make([]float64, 0, 60),
	}
}

// Run streams live pings with ANSI refresh.
func (m *LiveMonitor) Run(ctx context.Context, host string, port int, interval time.Duration) error {
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	if port <= 0 {
		port = 80
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	sent := 0
	received := 0
	var sumLat float64
	minLat := math.MaxFloat64
	maxLat := -math.MaxFloat64
	var currentJitter float64
	var prevTransit float64

	fmt.Print("\033[2J") // Clear terminal

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nMonitor terminated.")
			return nil
		case <-ticker.C:
			sent++
			probeCtx, cancel := context.WithTimeout(ctx, interval)
			res := netprobe.ProbeTCP(probeCtx, host, port, netprobe.DefaultTCPConfig(), interval)
			cancel()

			if res.Status == core.StatusUp {
				received++
				ms := res.LatencyMs
				sumLat += ms
				if ms < minLat {
					minLat = ms
				}
				if ms > maxLat {
					maxLat = ms
				}
				if received > 1 {
					d := math.Abs(ms - prevTransit)
					currentJitter += (d - currentJitter) / 16.0
				}
				prevTransit = ms
				m.history = append(m.history, ms)
				if len(m.history) > 50 {
					m.history = m.history[1:]
				}
			}

			lossRate := 0.0
			if sent > 0 {
				lossRate = (float64(sent-received) / float64(sent)) * 100.0
			}
			avgLat := 0.0
			if received > 0 {
				avgLat = sumLat / float64(received)
			}

			// Redraw screen
			fmt.Print("\033[H")
			fmt.Printf("⚡ %s - Real-time Network Telemetry [%s:%d]\n", color.CyanString("PINGFORGE LIVE"), host, port)
			fmt.Printf("   Packets: %s sent, %s recv | Loss: %s | Jitter: %s\n\n",
				color.WhiteString("%d", sent),
				color.GreenString("%d", received),
				formatLoss(lossRate),
				color.MagentaString("%.2f ms", currentJitter),
			)
			fmt.Printf("   Min: %.2f ms  |  Avg: %.2f ms  |  Max: %.2f ms  |  Latest: %.2f ms\n",
				minLat, avgLat, maxLat, res.LatencyMs)
			fmt.Printf("   History: %s\n\n", renderSparkline(m.history))
			fmt.Println("   Press Ctrl+C to stop monitoring.")
		}
	}
}

func formatLoss(rate float64) string {
	str := fmt.Sprintf("%.1f%%", rate)
	if rate == 0 {
		return color.GreenString(str)
	}
	if rate < 10 {
		return color.YellowString(str)
	}
	return color.RedString(str)
}

func renderSparkline(values []float64) string {
	if len(values) == 0 {
		return ""
	}
	sparkChars := []rune{' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	minVal := math.MaxFloat64
	maxVal := -math.MaxFloat64
	for _, v := range values {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	span := maxVal - minVal
	var sb strings.Builder
	for _, v := range values {
		if span <= 0 {
			sb.WriteRune(sparkChars[0])
			continue
		}
		idx := int(((v - minVal) / span) * float64(len(sparkChars)-1))
		if idx >= len(sparkChars) {
			idx = len(sparkChars) - 1
		}
		if idx < 0 {
			idx = 0
		}
		sb.WriteRune(sparkChars[idx])
	}
	return sb.String()
}
