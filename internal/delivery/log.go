package delivery

import (
	"log/slog"
	"strings"

	"github.com/kextant/kextant-agent/internal/report"
	"github.com/kextant/kextant-agent/pkg/types"
)

type LogDelivery struct {
	logger    *slog.Logger
	multiline bool
}

func NewLogDelivery(logger *slog.Logger, multiline bool) *LogDelivery {
	return &LogDelivery{
		logger:    logger,
		multiline: multiline,
	}
}

func (l *LogDelivery) Send(r *types.Report) error {
	plainText := report.GeneratePlainText(r)

	if l.multiline {
		// Output report across multiple log lines for readability
		l.logger.Info("Kextant health report",
			"cluster", r.ClusterName,
			"score", r.Score,
			"grade", r.Grade,
			"critical", r.CriticalCount,
			"warnings", r.WarningCount,
			"info", r.InfoCount)

		// Split the report into multiple log lines for readability
		lines := strings.Split(plainText, "\n")
		for _, line := range lines {
			if line != "" {
				l.logger.Info(line)
			}
		}

	} else {
		// Output as single structured log entry (JSON-friendly)
		l.logger.Info("Kextant health report",
			"cluster", r.ClusterName,
			"score", r.Score,
			"grade", r.Grade,
			"critical", r.CriticalCount,
			"warnings", r.WarningCount,
			"info", r.InfoCount,
			"report", plainText)
	}

	return nil
}
