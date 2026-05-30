package report

import (
	"fmt"
	"strings"

	"github.com/kextant/kextant-agent/pkg/types"
)

// GeneratePlainText generates a tightly formatted plain text report
// without colors or emoji, suitable for logging
func GeneratePlainText(r *types.Report) string {
	var sb strings.Builder

	// Header
	sb.WriteString("=== Kextant Health Report ===\n")
	sb.WriteString(fmt.Sprintf("Cluster: %s\n", r.ClusterName))
	sb.WriteString(fmt.Sprintf("Generated: %s\n", r.GeneratedAt.Format("2006-01-02 15:04:05 MST")))
	sb.WriteString(fmt.Sprintf("Score: %d/100 (%s)\n", r.Score, r.Grade))
	sb.WriteString(fmt.Sprintf("Findings: %d Critical, %d Warning, %d Info\n",
		r.CriticalCount, r.WarningCount, r.InfoCount))
	sb.WriteString("\n")

	// Group findings by severity
	criticalFindings := GetFindingsBySeverity(r.Findings, types.SeverityCritical)
	warningFindings := GetFindingsBySeverity(r.Findings, types.SeverityWarning)
	infoFindings := GetFindingsBySeverity(r.Findings, types.SeverityInfo)

	// Critical findings
	if len(criticalFindings) > 0 {
		sb.WriteString("--- CRITICAL ISSUES ---\n")
		for _, f := range criticalFindings {
			sb.WriteString(formatFinding(f))
		}
		sb.WriteString("\n")
	}

	// Warning findings
	if len(warningFindings) > 0 {
		sb.WriteString("--- WARNINGS ---\n")
		for _, f := range warningFindings {
			sb.WriteString(formatFinding(f))
		}
		sb.WriteString("\n")
	}

	// Info findings
	if len(infoFindings) > 0 {
		sb.WriteString("--- INFO ---\n")
		for _, f := range infoFindings {
			sb.WriteString(formatFinding(f))
		}
		sb.WriteString("\n")
	}

	// Summary
	if len(r.Findings) == 0 {
		sb.WriteString("No issues found. Cluster health is excellent.\n")
	}

	sb.WriteString("=== End Report ===")

	return sb.String()
}

func formatFinding(f types.Finding) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("\n[%s] %s\n", f.ID, f.Title))

	if f.Description != "" {
		sb.WriteString(fmt.Sprintf("  Description: %s\n", f.Description))
	}

	if len(f.Affected) > 0 {
		sb.WriteString("  Affected Resources:\n")
		for _, res := range f.Affected {
			sb.WriteString(fmt.Sprintf("    - %s/%s (%s)\n", res.Namespace, res.Name, res.Kind))
		}
	}

	if f.Risk != "" {
		sb.WriteString(fmt.Sprintf("  Risk: %s\n", f.Risk))
	}

	if f.Fix != "" {
		sb.WriteString(fmt.Sprintf("  Fix: %s\n", f.Fix))
	}

	return sb.String()
}
