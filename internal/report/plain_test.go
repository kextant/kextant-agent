package report

import (
	"testing"
	"time"

	"github.com/kextant/kextant-agent/pkg/types"
)

func TestGeneratePlainText(t *testing.T) {
	// Create a sample report
	report := &types.Report{
		ClusterName:   "test-cluster",
		GeneratedAt:   time.Date(2026, 1, 23, 9, 0, 0, 0, time.UTC),
		Score:         75,
		Grade:         "Good",
		CriticalCount: 2,
		WarningCount:  3,
		InfoCount:     1,
		Findings: []types.Finding{
			{
				ID:          "RES004",
				Title:       "Missing memory limits",
				Description: "Container has no memory limit defined",
				Severity:    types.SeverityCritical,
				Category:    "resources",
				Affected: []types.AffectedResource{
					{
						Kind:      "Pod",
						Namespace: "production",
						Name:      "api-server-abc123",
					},
					{
						Kind:      "Pod",
						Namespace: "production",
						Name:      "worker-xyz789",
					},
				},
				Risk: "Unbounded memory usage can lead to node OOM kills",
				Fix:  "kubectl set resources pod/api-server-abc123 --limits=memory=512Mi",
			},
			{
				ID:          "SEC002",
				Title:       "Privileged container",
				Description: "Container is running in privileged mode",
				Severity:    types.SeverityCritical,
				Category:    "security",
				Affected: []types.AffectedResource{
					{
						Kind:      "Pod",
						Namespace: "kube-system",
						Name:      "debug-pod",
					},
				},
				Risk: "Privileged containers can compromise the entire node",
				Fix:  "Remove securityContext.privileged: true from pod spec",
			},
			{
				ID:          "PRB001",
				Title:       "Missing readiness probe",
				Description: "Container has no readiness probe configured",
				Severity:    types.SeverityWarning,
				Category:    "probes",
				Affected: []types.AffectedResource{
					{
						Kind:      "Deployment",
						Namespace: "staging",
						Name:      "frontend",
					},
				},
				Risk: "Traffic may be routed to unhealthy pods",
				Fix:  "Add readinessProbe to container spec",
			},
			{
				ID:          "IMG001",
				Title:       "Using :latest tag",
				Description: "Image uses the :latest tag",
				Severity:    types.SeverityWarning,
				Category:    "images",
				Affected: []types.AffectedResource{
					{
						Kind:      "Deployment",
						Namespace: "development",
						Name:      "test-app",
					},
				},
				Risk: "Deployments become unpredictable",
				Fix:  "Pin to specific image tag (e.g., nginx:1.21.0)",
			},
			{
				ID:          "REP001",
				Title:       "Single replica deployment",
				Description: "Deployment has only 1 replica",
				Severity:    types.SeverityWarning,
				Category:    "replicas",
				Affected: []types.AffectedResource{
					{
						Kind:      "Deployment",
						Namespace: "production",
						Name:      "payment-service",
					},
				},
				Risk: "No high availability",
				Fix:  "kubectl scale deployment/payment-service --replicas=3",
			},
			{
				ID:          "NS001",
				Title:       "No resource quotas",
				Description: "Namespace has no ResourceQuota defined",
				Severity:    types.SeverityInfo,
				Category:    "namespace",
				Affected: []types.AffectedResource{
					{
						Kind:      "Namespace",
						Namespace: "development",
						Name:      "development",
					},
				},
				Risk: "Namespace can consume unlimited resources",
				Fix:  "Create ResourceQuota for the namespace",
			},
		},
	}

	plainText := GeneratePlainText(report)

	// Just verify it doesn't panic and produces output
	if plainText == "" {
		t.Error("Expected non-empty plain text output")
	}

	// Print for manual inspection
	t.Logf("\n%s\n", plainText)
}
