package report

import (
	"time"

	"github.com/kextant/kextant-agent/pkg/types"
)

func Generate(clusterName string, findings []types.Finding) *types.Report {
	now := time.Now()

	report := &types.Report{
		ClusterName: clusterName,
		GeneratedAt: now,
		Findings:    findings,
	}

	// Count findings by severity
	for _, f := range findings {
		switch f.Severity {
		case types.SeverityCritical:
			report.CriticalCount++
		case types.SeverityWarning:
			report.WarningCount++
		case types.SeverityInfo:
			report.InfoCount++
		}
	}

	// Calculate score
	report.Score = CalculateScore(findings)
	report.Grade = GetGrade(report.Score)

	// Calculate total and passed checks
	report.TotalChecks = len(findings)
	report.PassedChecks = 0 // In MVP, we only report failures

	return report
}

func GetFindingsBySeverity(findings []types.Finding, severity types.Severity) []types.Finding {
	var result []types.Finding
	for _, f := range findings {
		if f.Severity == severity {
			result = append(result, f)
		}
	}
	return result
}

func GetFindingsByCategory(findings []types.Finding, category string) []types.Finding {
	var result []types.Finding
	for _, f := range findings {
		if f.Category == category {
			result = append(result, f)
		}
	}
	return result
}
