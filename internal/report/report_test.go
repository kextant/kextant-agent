package report

import (
	"encoding/json"
	"testing"

	"github.com/kextant/kextant-agent/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestGenerateCountsAndScore(t *testing.T) {
	findings := []types.Finding{
		{ID: "C", Severity: types.SeverityCritical, Category: "security"},
		{ID: "W", Severity: types.SeverityWarning, Category: "resources"},
		{ID: "I", Severity: types.SeverityInfo, Category: "images"},
	}
	report := Generate("prod", findings)
	require.Equal(t, "prod", report.ClusterName)
	require.Equal(t, 1, report.CriticalCount)
	require.Equal(t, 1, report.WarningCount)
	require.Equal(t, 1, report.InfoCount)
	require.Equal(t, 86, report.Score)
	require.Equal(t, "Good", report.Grade)
}

func TestGetGradeAndEmoji(t *testing.T) {
	require.Equal(t, "Excellent", GetGrade(95))
	require.Equal(t, "Good", GetGrade(80))
	require.Equal(t, "Fair", GetGrade(65))
	require.Equal(t, "Poor", GetGrade(45))
	require.Equal(t, "Critical", GetGrade(10))
	require.Equal(t, "✅", GetGradeEmoji("Excellent"))
	require.Equal(t, "⚪", GetGradeEmoji("unknown"))
}

func TestGetFindingsBySeverityAndCategory(t *testing.T) {
	findings := []types.Finding{
		{ID: "C", Severity: types.SeverityCritical, Category: "security"},
		{ID: "W", Severity: types.SeverityWarning, Category: "resources"},
	}
	require.Equal(t, []types.Finding{findings[0]}, GetFindingsBySeverity(findings, types.SeverityCritical))
	require.Equal(t, []types.Finding{findings[1]}, GetFindingsByCategory(findings, "resources"))
}

func TestGenerateSlackBlocks(t *testing.T) {
	report := &types.Report{
		ClusterName:   "prod",
		Score:         70,
		Grade:         "Fair",
		CriticalCount: 1,
		WarningCount:  4,
		Findings: []types.Finding{
			{Title: "critical", Severity: types.SeverityCritical, Risk: "risk", Affected: []types.AffectedResource{{Namespace: "default", Name: "api"}}},
			{Title: "warning-1", Severity: types.SeverityWarning},
			{Title: "warning-2", Severity: types.SeverityWarning},
			{Title: "warning-3", Severity: types.SeverityWarning},
			{Title: "warning-4", Severity: types.SeverityWarning},
		},
	}
	payload, err := GenerateSlackBlocks(report)
	require.NoError(t, err)
	var message SlackMessage
	require.NoError(t, json.Unmarshal(payload, &message))
	require.NotEmpty(t, message.Blocks)
	require.Contains(t, string(payload), "and 1 more warnings")
}
