package types

import "time"

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityWarning  Severity = "warning"
	SeverityInfo     Severity = "info"
)

type Finding struct {
	ID          string             `json:"id"`          // e.g., "RES001"
	Title       string             `json:"title"`       // e.g., "Missing memory limits"
	Description string             `json:"description"` // Detailed explanation
	Severity    Severity           `json:"severity"`
	Category    string             `json:"category"` // e.g., "resources"
	Affected    []AffectedResource `json:"affected"`
	Risk        string             `json:"risk"` // Why it matters
	Fix         string             `json:"fix"`  // How to fix (kubectl command)
}

type AffectedResource struct {
	Kind      string `json:"kind"` // e.g., "Deployment"
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type Report struct {
	ClusterName   string    `json:"clusterName"`
	GeneratedAt   time.Time `json:"generatedAt"`
	Score         int       `json:"score"`
	Grade         string    `json:"grade"`
	TotalChecks   int       `json:"totalChecks"`
	PassedChecks  int       `json:"passedChecks"`
	CriticalCount int       `json:"criticalCount"`
	WarningCount  int       `json:"warningCount"`
	InfoCount     int       `json:"infoCount"`
	ExemptedCount int       `json:"exemptedCount"` // Reserved for future use
	Findings      []Finding `json:"findings"`
}
