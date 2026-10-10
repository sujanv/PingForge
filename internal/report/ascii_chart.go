package report

import (
	"fmt"
	"math"
	"strings"
)

var sparkChars = []rune{' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// GenerateSparkline produces an ASCII sparkline representation of a slice of values.
func GenerateSparkline(values []float64) string {
	if len(values) == 0 {
		return ""
	}

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
		ratio := (v - minVal) / span
		idx := int(ratio * float64(len(sparkChars)-1))
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

// GenerateHistogram produces an ASCII frequency distribution.
func GenerateHistogram(values []float64, bins int) string {
	if len(values) == 0 {
		return ""
	}
	if bins <= 0 {
		bins = 6
	}

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
	if span <= 0 {
		return fmt.Sprintf("All samples equal to %.2f ms (Count: %d)\n", minVal, len(values))
	}

	binWidth := span / float64(bins)
	counts := make([]int, bins)
	maxCount := 0

	for _, v := range values {
		idx := int((v - minVal) / binWidth)
		if idx >= bins {
			idx = bins - 1
		}
		counts[idx]++
		if counts[idx] > maxCount {
			maxCount = counts[idx]
		}
	}

	var sb strings.Builder
	sb.WriteString("\n  Latency Distribution (Histogram):\n")
	barMaxLen := 30

	for i := 0; i < bins; i++ {
		low := minVal + float64(i)*binWidth
		high := low + binWidth
		barLen := 0
		if maxCount > 0 {
			barLen = int((float64(counts[i]) / float64(maxCount)) * float64(barMaxLen))
		}
		bar := strings.Repeat("█", barLen)
		sb.WriteString(fmt.Sprintf("    [%6.1f - %6.1f ms] : %-30s (%d)\n", low, high, bar, counts[i]))
	}

	return sb.String()
}
