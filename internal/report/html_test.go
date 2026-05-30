package report

import (
	"strings"
	"testing"
	"time"

	"github.com/kextant/kextant-agent/pkg/types"
)

func TestGenerateHTML_IncludesCheckIDs(t *testing.T) {
	r := &types.Report{
		ClusterName:   "test",
		GeneratedAt:   time.Now(),
		Score:         85,
		Grade:         "Good",
		CriticalCount: 1,
		WarningCount:  1,
		InfoCount:     0,
		Findings: []types.Finding{
			{
				ID:          "SEC003",
				Title:       "Root filesystem writable",
				Description: "Container has writable root filesystem",
				Severity:    types.SeverityWarning,
				Risk:        "Writable root filesystem allows malicious code to modify the container filesystem",
				Fix:         "Set readOnlyRootFilesystem: true in securityContext",
			},
			{
				ID:          "RES004",
				Title:       "Missing memory limits",
				Description: "Container has no memory limit defined",
				Severity:    types.SeverityCritical,
				Risk:        "Without memory limits, a container can consume all available memory",
				Fix:         "Set memory limits in resources",
			},
		},
	}

	html, err := GenerateHTML(r)
	if err != nil {
		t.Fatalf("GenerateHTML failed: %v", err)
	}

	// Verify check IDs are included in titles
	if !strings.Contains(html, "Root filesystem writable [SEC003]") {
		t.Error("Expected 'Root filesystem writable [SEC003]' in HTML, but it was not found")
	}

	if !strings.Contains(html, "Missing memory limits [RES004]") {
		t.Error("Expected 'Missing memory limits [RES004]' in HTML, but it was not found")
	}
}
