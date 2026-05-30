package report

import (
	"github.com/kextant/kextant-agent/pkg/types"
)

func CalculateScore(findings []types.Finding) int {
	// Start with perfect score
	score := 100

	// Deduct points based on severity
	for _, f := range findings {
		switch f.Severity {
		case types.SeverityCritical:
			score -= 10
		case types.SeverityWarning:
			score -= 3
		case types.SeverityInfo:
			score--
		}
	}

	// Ensure score doesn't go below 0
	if score < 0 {
		score = 0
	}

	return score
}

func GetGrade(score int) string {
	switch {
	case score >= 90:
		return "Excellent"
	case score >= 75:
		return "Good"
	case score >= 60:
		return "Fair"
	case score >= 40:
		return "Poor"
	default:
		return "Critical"
	}
}

func GetGradeEmoji(grade string) string {
	switch grade {
	case "Excellent":
		return "✅"
	case "Good":
		return "🟢"
	case "Fair":
		return "🟡"
	case "Poor":
		return "🟠"
	case "Critical":
		return "🔴"
	default:
		return "⚪"
	}
}
