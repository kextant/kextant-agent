package scanner

import (
	"context"
	"fmt"

	"github.com/kextant/kextant-agent/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Scanner) checkReplicas(ctx context.Context, namespaces []string) ([]types.Finding, error) {
	var findings []types.Finding

	for _, ns := range namespaces {
		deployments, err := s.client.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		for _, deploy := range deployments.Items {
			// REP001: Single replica deployment
			if deploy.Spec.Replicas != nil && *deploy.Spec.Replicas == 1 {
				if s.exemptions.IsExemptDeployment(&deploy, "REP001") {
					continue
				}
				findings = append(findings, types.Finding{
					ID:          "REP001",
					Title:       "Single replica deployment",
					Description: fmt.Sprintf("Deployment %s has only 1 replica", deploy.Name),
					Severity:    types.SeverityWarning,
					Category:    "replicas",
					Affected: []types.AffectedResource{{
						Kind:      "Deployment",
						Namespace: deploy.Namespace,
						Name:      deploy.Name,
					}},
					Risk: "Single replica deployments have no high availability and will have downtime during updates",
					Fix: fmt.Sprintf(`kubectl scale deployment %s -n %s --replicas=3`,
						deploy.Name, deploy.Namespace),
				})
			}

			// REP003: Mismatched replicas
			if deploy.Spec.Replicas != nil && deploy.Status.AvailableReplicas != *deploy.Spec.Replicas {
				if s.exemptions.IsExemptDeployment(&deploy, "REP003") {
					continue
				}
				findings = append(findings, types.Finding{
					ID:          "REP003",
					Title:       "Mismatched replicas",
					Description: fmt.Sprintf("Deployment %s has %d desired replicas but only %d available", deploy.Name, *deploy.Spec.Replicas, deploy.Status.AvailableReplicas),
					Severity:    types.SeverityWarning,
					Category:    "replicas",
					Affected: []types.AffectedResource{{
						Kind:      "Deployment",
						Namespace: deploy.Namespace,
						Name:      deploy.Name,
					}},
					Risk: "Deployment is not running at desired capacity, which may indicate resource constraints or pod failures",
					Fix: fmt.Sprintf(`kubectl describe deployment %s -n %s`,
						deploy.Name, deploy.Namespace),
				})
			}
		}

		// REP002: No PodDisruptionBudget
		pdbs, err := s.client.PolicyV1().PodDisruptionBudgets(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		// Create a map of PDBs by selector
		pdbMap := make(map[string]bool)
		for _, pdb := range pdbs.Items {
			pdbMap[pdb.Name] = true
		}

		// Check each deployment for PDB
		for _, deploy := range deployments.Items {
			// Skip single-replica deployments (they can't have PDBs anyway)
			if deploy.Spec.Replicas != nil && *deploy.Spec.Replicas <= 1 {
				continue
			}

			// Simple check: does a PDB with the same name exist?
			// In a more sophisticated version, we'd match by label selector
			if !pdbMap[deploy.Name] {
				if s.exemptions.IsExemptDeployment(&deploy, "REP002") {
					continue
				}
				findings = append(findings, types.Finding{
					ID:          "REP002",
					Title:       "No PodDisruptionBudget",
					Description: fmt.Sprintf("Deployment %s has no PodDisruptionBudget defined", deploy.Name),
					Severity:    types.SeverityWarning,
					Category:    "replicas",
					Affected: []types.AffectedResource{{
						Kind:      "Deployment",
						Namespace: deploy.Namespace,
						Name:      deploy.Name,
					}},
					Risk: "Without a PDB, cluster operations like node drains can disrupt all replicas simultaneously",
					Fix:  "Create a PodDisruptionBudget to ensure availability during voluntary disruptions",
				})
			}
		}
	}

	return findings, nil
}
