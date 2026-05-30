package delivery

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/kextant/kextant-agent/pkg/types"
)

func TestLogDelivery_Multiline(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	report := &types.Report{
		ClusterName:   "test-cluster",
		GeneratedAt:   time.Now(),
		Score:         85,
		Grade:         "Good",
		CriticalCount: 1,
		WarningCount:  2,
		InfoCount:     1,
		Findings: []types.Finding{
			{
				ID:       "RES004",
				Title:    "Missing memory limits",
				Severity: types.SeverityCritical,
				Category: "resources",
			},
		},
	}

	// Test multiline format
	delivery := NewLogDelivery(logger, true)
	err := delivery.Send(report)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	t.Logf("Multiline output:\n%s", output)

	// Should have multiple log entries
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 5 { // At least summary + some report lines
		t.Errorf("Expected multiple log lines, got %d", len(lines))
	}

	// Should NOT contain escaped newlines in individual lines
	for i, line := range lines {
		if strings.Contains(line, "\\n") {
			t.Errorf("Line %d contains escaped newline: %s", i, line)
		}
	}
}

func TestLogDelivery_SingleLine(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	report := &types.Report{
		ClusterName:   "test-cluster",
		GeneratedAt:   time.Now(),
		Score:         85,
		Grade:         "Good",
		CriticalCount: 1,
		WarningCount:  2,
		InfoCount:     1,
		Findings: []types.Finding{
			{
				ID:       "RES004",
				Title:    "Missing memory limits",
				Severity: types.SeverityCritical,
				Category: "resources",
			},
		},
	}

	// Test single-line format
	delivery := NewLogDelivery(logger, false)
	err := delivery.Send(report)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	t.Logf("Single-line output:\n%s", output)

	// Should have exactly one log entry
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 1 {
		t.Errorf("Expected single log line, got %d", len(lines))
	}

	// Should contain the report field with escaped newlines
	if !strings.Contains(output, `"report"`) {
		t.Error("Expected 'report' field in structured log")
	}

	// The report field should contain escaped newlines
	if !strings.Contains(output, "\\n") {
		t.Error("Expected escaped newlines (\\n) in single-line format")
	}
}
