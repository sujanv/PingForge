package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sujanv/pingforge/internal/core"
)

// WebhookPayload represents rich formatted alert notification
type WebhookPayload struct {
	Title       string    `json:"title"`
	Event       string    `json:"event"`
	Target      string    `json:"target"`
	Severity    string    `json:"severity"` // INFO, WARNING, CRITICAL
	LatencyMs   float64   `json:"latency_ms"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	PingForgeVer string   `json:"pingforge_version"`
}

// DispatchWebhook sends an event payload to a configured endpoint (Slack, Discord, PagerDuty, generic)
func DispatchWebhook(ctx context.Context, webhookURL string, payload WebhookPayload, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	payload.PingForgeVer = "v4.0.0"

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("create webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "PingForge-Alert-Router/4.0")

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("post webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook endpoint returned status %d", resp.StatusCode)
	}

	return nil
}
