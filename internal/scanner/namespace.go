package scanner

import (
	"context"
	"fmt"

	"github.com/kextant/kextant-agent/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Scanner) checkNamespaces(ctx context.Context, namespaces []string) ([]types.Finding, error) {
	var findings []types.Finding

	for _, ns := range namespaces {
		// Get namespace object for exemption checks
		nsObj, err := s.client.CoreV1().Namespaces().Get(ctx, ns, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		// NS001: No resource quotas
		quotas, err := s.client.CoreV1().ResourceQuotas(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		if len(quotas.Items) == 0 {
			if !s.exemptions.IsExemptNamespace(nsObj, "NS001") {
				findings = append(findings, types.Finding{
					ID:          "NS001",
					Title:       "No resource quotas",
					Description: fmt.Sprintf("Namespace %s has no ResourceQuota defined", ns),
					Severity:    types.SeverityInfo,
					Category:    "namespace",
					Affected: []types.AffectedResource{{
						Kind:      "Namespace",
						Namespace: ns,
						Name:      ns,
					}},
					Risk: "Without resource quotas, a single namespace can consume all cluster resources",
					Fix:  "Create ResourceQuota to limit resource consumption in the namespace",
				})
			}
		}

		// NS002: No limit ranges
		limitRanges, err := s.client.CoreV1().LimitRanges(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		if len(limitRanges.Items) == 0 {
			if !s.exemptions.IsExemptNamespace(nsObj, "NS002") {
				findings = append(findings, types.Finding{
					ID:          "NS002",
					Title:       "No limit ranges",
					Description: fmt.Sprintf("Namespace %s has no LimitRange defined", ns),
					Severity:    types.SeverityInfo,
					Category:    "namespace",
					Affected: []types.AffectedResource{{
						Kind:      "Namespace",
						Namespace: ns,
						Name:      ns,
					}},
					Risk: "Without limit ranges, pods can be created without resource limits",
					Fix:  "Create LimitRange to set default resource limits for containers",
				})
			}
		}

		// NS003: No network policies
		netPolicies, err := s.client.NetworkingV1().NetworkPolicies(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		if len(netPolicies.Items) == 0 {
			if !s.exemptions.IsExemptNamespace(nsObj, "NS003") {
				findings = append(findings, types.Finding{
					ID:          "NS003",
					Title:       "No network policies",
					Description: fmt.Sprintf("Namespace %s has no NetworkPolicy defined", ns),
					Severity:    types.SeverityWarning,
					Category:    "namespace",
					Affected: []types.AffectedResource{{
						Kind:      "Namespace",
						Namespace: ns,
						Name:      ns,
					}},
					Risk: "Without network policies, all pods can communicate freely, increasing security risk",
					Fix:  "Create NetworkPolicy to restrict pod-to-pod communication",
				})
			}
		}
	}

	return findings, nil
}
