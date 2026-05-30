package delivery

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/kextant/kextant-agent/internal/report"
	"github.com/kextant/kextant-agent/pkg/types"
)

type SlackDelivery struct {
	webhookURL string
	logger     *slog.Logger
}

func NewSlackDelivery(webhookURL string, logger *slog.Logger) *SlackDelivery {
	return &SlackDelivery{
		webhookURL: webhookURL,
		logger:     logger,
	}
}

func (s *SlackDelivery) Send(r *types.Report) error {
	s.logger.Info("sending report to Slack", "cluster", r.ClusterName)

	payload, err := report.GenerateSlackBlocks(r)
	if err != nil {
		return fmt.Errorf("failed to generate Slack payload: %w", err)
	}

	return s.deliverWithRetry(payload)
}

func (s *SlackDelivery) deliverWithRetry(payload []byte) error {
	delays := []time.Duration{1 * time.Second, 5 * time.Second, 30 * time.Second}

	var lastErr error
	for i, delay := range delays {
		if err := s.sendWebhook(payload); err != nil {
			lastErr = err
			s.logger.Error("Slack delivery failed",
				"attempt", i+1,
				"max_attempts", len(delays),
				"error", err)
			if i < len(delays)-1 {
				time.Sleep(delay)
			}
			continue
		}
		s.logger.Info("successfully delivered report to Slack")
		return nil
	}

	return fmt.Errorf("Slack delivery failed after %d attempts: %w", len(delays), lastErr)
}

func (s *SlackDelivery) sendWebhook(payload []byte) error {
	resp, err := http.Post(s.webhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Slack webhook returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
