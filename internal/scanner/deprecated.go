package scanner

import (
	"context"
	"fmt"

	"github.com/kextant/kextant-agent/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Deprecated and removed API versions mapping
// Based on Kubernetes deprecation guide
var deprecatedAPIs = map[string]struct {
	resource   string
	deprecated string
	removed    string
	current    string
}{
	"extensions/v1beta1/Ingress": {
		resource:   "Ingress",
		deprecated: "1.14",
		removed:    "1.22",
		current:    "networking.k8s.io/v1",
	},
	"networking.k8s.io/v1beta1/Ingress": {
		resource:   "Ingress",
		deprecated: "1.19",
		removed:    "1.22",
		current:    "networking.k8s.io/v1",
	},
	"networking.k8s.io/v1beta1/IngressClass": {
		resource:   "IngressClass",
		deprecated: "1.19",
		removed:    "1.22",
		current:    "networking.k8s.io/v1",
	},
	"policy/v1beta1/PodDisruptionBudget": {
		resource:   "PodDisruptionBudget",
		deprecated: "1.21",
		removed:    "1.25",
		current:    "policy/v1",
	},
	"policy/v1beta1/PodSecurityPolicy": {
		resource:   "PodSecurityPolicy",
		deprecated: "1.21",
		removed:    "1.25",
		current:    "Pod Security Admission",
	},
	"batch/v1beta1/CronJob": {
		resource:   "CronJob",
		deprecated: "1.21",
		removed:    "1.25",
		current:    "batch/v1",
	},
	"autoscaling/v2beta1/HorizontalPodAutoscaler": {
		resource:   "HorizontalPodAutoscaler",
		deprecated: "1.22",
		removed:    "1.26",
		current:    "autoscaling/v2",
	},
	"autoscaling/v2beta2/HorizontalPodAutoscaler": {
		resource:   "HorizontalPodAutoscaler",
		deprecated: "1.23",
		removed:    "1.26",
		current:    "autoscaling/v2",
	},
}

func (s *Scanner) checkDeprecatedAPIs(ctx context.Context, namespaces []string) ([]types.Finding, error) {
	var findings []types.Finding

	for _, ns := range namespaces {
		// Check Ingresses
		ingresses, err := s.client.NetworkingV1().Ingresses(ns).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, ing := range ingresses.Items {
				apiVersion := ing.APIVersion
				if apiVersion != "networking.k8s.io/v1" {
					key := fmt.Sprintf("%s/Ingress", apiVersion)
					if info, ok := deprecatedAPIs[key]; ok {
						if s.exemptions.IsExemptResource(ing.Annotations, "API001") {
							continue
						}
						findings = append(findings, types.Finding{
							ID:          "API001",
							Title:       "Deprecated API version",
							Description: fmt.Sprintf("Ingress %s uses deprecated API version %s", ing.Name, apiVersion),
							Severity:    types.SeverityCritical,
							Category:    "deprecated-api",
							Affected: []types.AffectedResource{{
								Kind:      "Ingress",
								Namespace: ing.Namespace,
								Name:      ing.Name,
							}},
							Risk: fmt.Sprintf("This API version was deprecated in Kubernetes %s and removed in %s", info.deprecated, info.removed),
							Fix:  fmt.Sprintf("Migrate to %s API version", info.current),
						})
					}
				}
			}
		}

		// Check PodDisruptionBudgets
		pdbs, err := s.client.PolicyV1().PodDisruptionBudgets(ns).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, pdb := range pdbs.Items {
				apiVersion := pdb.APIVersion
				if apiVersion != "policy/v1" {
					key := fmt.Sprintf("%s/PodDisruptionBudget", apiVersion)
					if info, ok := deprecatedAPIs[key]; ok {
						if s.exemptions.IsExemptResource(pdb.Annotations, "API001") {
							continue
						}
						findings = append(findings, types.Finding{
							ID:          "API001",
							Title:       "Deprecated API version",
							Description: fmt.Sprintf("PodDisruptionBudget %s uses deprecated API version %s", pdb.Name, apiVersion),
							Severity:    types.SeverityCritical,
							Category:    "deprecated-api",
							Affected: []types.AffectedResource{{
								Kind:      "PodDisruptionBudget",
								Namespace: pdb.Namespace,
								Name:      pdb.Name,
							}},
							Risk: fmt.Sprintf("This API version was deprecated in Kubernetes %s and removed in %s", info.deprecated, info.removed),
							Fix:  fmt.Sprintf("Migrate to %s API version", info.current),
						})
					}
				}
			}
		}

		// Check CronJobs
		cronjobs, err := s.client.BatchV1().CronJobs(ns).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, cj := range cronjobs.Items {
				apiVersion := cj.APIVersion
				if apiVersion != "batch/v1" {
					key := fmt.Sprintf("%s/CronJob", apiVersion)
					if info, ok := deprecatedAPIs[key]; ok {
						if s.exemptions.IsExemptResource(cj.Annotations, "API001") {
							continue
						}
						findings = append(findings, types.Finding{
							ID:          "API001",
							Title:       "Deprecated API version",
							Description: fmt.Sprintf("CronJob %s uses deprecated API version %s", cj.Name, apiVersion),
							Severity:    types.SeverityCritical,
							Category:    "deprecated-api",
							Affected: []types.AffectedResource{{
								Kind:      "CronJob",
								Namespace: cj.Namespace,
								Name:      cj.Name,
							}},
							Risk: fmt.Sprintf("This API version was deprecated in Kubernetes %s and removed in %s", info.deprecated, info.removed),
							Fix:  fmt.Sprintf("Migrate to %s API version", info.current),
						})
					}
				}
			}
		}

		// Check HorizontalPodAutoscalers
		hpas, err := s.client.AutoscalingV2().HorizontalPodAutoscalers(ns).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, hpa := range hpas.Items {
				apiVersion := hpa.APIVersion
				if apiVersion != "autoscaling/v2" {
					key := fmt.Sprintf("%s/HorizontalPodAutoscaler", apiVersion)
					if info, ok := deprecatedAPIs[key]; ok {
						if s.exemptions.IsExemptResource(hpa.Annotations, "API001") {
							continue
						}
						findings = append(findings, types.Finding{
							ID:          "API001",
							Title:       "Deprecated API version",
							Description: fmt.Sprintf("HorizontalPodAutoscaler %s uses deprecated API version %s", hpa.Name, apiVersion),
							Severity:    types.SeverityCritical,
							Category:    "deprecated-api",
							Affected: []types.AffectedResource{{
								Kind:      "HorizontalPodAutoscaler",
								Namespace: hpa.Namespace,
								Name:      hpa.Name,
							}},
							Risk: fmt.Sprintf("This API version was deprecated in Kubernetes %s and removed in %s", info.deprecated, info.removed),
							Fix:  fmt.Sprintf("Migrate to %s API version", info.current),
						})
					}
				}
			}
		}
	}

	return findings, nil
}
