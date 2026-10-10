package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sujanv/pingforge/internal/config"
	"github.com/sujanv/pingforge/internal/core"
)

// AlertPayload defines the webhook payload for triggered alerts.
type AlertPayload struct {
	Timestamp string  `json:"timestamp"`
	Target    string  `json:"target"`
	Status    string  `json:"status"`
	LatencyMs float64 `json:"latency_ms"`
	Message   string  `json:"message"`
	Threshold float64 `json:"threshold_ms,omitempty"`
}

// CheckAndNotify evaluates a probe result and fires a webhook if thresholds are breached.
func CheckAndNotify(ctx context.Context, cfg config.AlertConfig, res *core.ProbeResult) error {
	if !cfg.Enabled || cfg.WebhookURL == "" {
		return nil
	}

	shouldAlert := false
	alertMsg := ""

	if res.Status != core.StatusUp {
		shouldAlert = true
		alertMsg = fmt.Sprintf("Target %s is %s: %s", res.Target.String(), res.Status, res.Message)
	} else if cfg.LatencyThresholdMs > 0 && res.LatencyMs > cfg.LatencyThresholdMs {
		shouldAlert = true
		alertMsg = fmt.Sprintf("Target %s latency %.2fms exceeded threshold %.2fms",
			res.Target.String(), res.LatencyMs, cfg.LatencyThresholdMs)
	}

	if !shouldAlert {
		return nil
	}

	payload := AlertPayload{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Target:    res.Target.String(),
		Status:    string(res.Status),
		LatencyMs: res.LatencyMs,
		Message:   alertMsg,
		Threshold: cfg.LatencyThresholdMs,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.WebhookURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "PingForge-Alerts/1.0")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("dispatch webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook responded with HTTP %d", resp.StatusCode)
	}

	return nil
}
