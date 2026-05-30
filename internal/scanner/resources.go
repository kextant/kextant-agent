package scanner

import (
	"context"
	"fmt"

	"github.com/kextant/kextant-agent/pkg/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Scanner) checkResources(ctx context.Context, namespaces []string) ([]types.Finding, error) {
	var findings []types.Finding

	for _, ns := range namespaces {
		pods, err := s.client.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		for _, pod := range pods.Items {
			findings = append(findings, s.checkPodResources(ctx, pod)...)
		}
	}

	return findings, nil
}

func (s *Scanner) checkPodResources(ctx context.Context, pod corev1.Pod) []types.Finding {
	var findings []types.Finding

	for _, container := range pod.Spec.Containers {
		// RES001: Missing CPU requests
		if container.Resources.Requests.Cpu().IsZero() {
			if s.exemptions.IsExemptPod(ctx, &pod, "RES001") {
				continue
			}
			findings = append(findings, types.Finding{
				ID:          "RES001",
				Title:       "Missing CPU requests",
				Description: fmt.Sprintf("Container %s in pod %s has no CPU request defined", container.Name, pod.Name),
				Severity:    types.SeverityWarning,
				Category:    "resources",
				Affected: []types.AffectedResource{{
					Kind:      "Pod",
					Namespace: pod.Namespace,
					Name:      fmt.Sprintf("%s/%s", pod.Name, container.Name),
				}},
				Risk: "Without CPU requests, the scheduler cannot make informed decisions about pod placement, potentially leading to CPU starvation",
				Fix: fmt.Sprintf(`kubectl set resources pod %s -n %s --requests=cpu=100m --container=%s`,
					pod.Name, pod.Namespace, container.Name),
			})
		}

		// RES002: Missing CPU limits
		if container.Resources.Limits.Cpu().IsZero() {
			if s.exemptions.IsExemptPod(ctx, &pod, "RES002") {
				continue
			}
			findings = append(findings, types.Finding{
				ID:          "RES002",
				Title:       "Missing CPU limits",
				Description: fmt.Sprintf("Container %s in pod %s has no CPU limit defined", container.Name, pod.Name),
				Severity:    types.SeverityWarning,
				Category:    "resources",
				Affected: []types.AffectedResource{{
					Kind:      "Pod",
					Namespace: pod.Namespace,
					Name:      fmt.Sprintf("%s/%s", pod.Name, container.Name),
				}},
				Risk: "Without CPU limits, a container can consume excessive CPU resources, impacting other workloads",
				Fix: fmt.Sprintf(`kubectl set resources pod %s -n %s --limits=cpu=500m --container=%s`,
					pod.Name, pod.Namespace, container.Name),
			})
		}

		// RES003: Missing memory requests
		if container.Resources.Requests.Memory().IsZero() {
			if s.exemptions.IsExemptPod(ctx, &pod, "RES003") {
				continue
			}
			findings = append(findings, types.Finding{
				ID:          "RES003",
				Title:       "Missing memory requests",
				Description: fmt.Sprintf("Container %s in pod %s has no memory request defined", container.Name, pod.Name),
				Severity:    types.SeverityWarning,
				Category:    "resources",
				Affected: []types.AffectedResource{{
					Kind:      "Pod",
					Namespace: pod.Namespace,
					Name:      fmt.Sprintf("%s/%s", pod.Name, container.Name),
				}},
				Risk: "Without memory requests, the scheduler cannot make informed decisions about pod placement",
				Fix: fmt.Sprintf(`kubectl set resources pod %s -n %s --requests=memory=128Mi --container=%s`,
					pod.Name, pod.Namespace, container.Name),
			})
		}

		// RES004: Missing memory limits
		if container.Resources.Limits.Memory().IsZero() {
			if s.exemptions.IsExemptPod(ctx, &pod, "RES004") {
				continue
			}
			findings = append(findings, types.Finding{
				ID:          "RES004",
				Title:       "Missing memory limits",
				Description: fmt.Sprintf("Container %s in pod %s has no memory limit defined", container.Name, pod.Name),
				Severity:    types.SeverityCritical,
				Category:    "resources",
				Affected: []types.AffectedResource{{
					Kind:      "Pod",
					Namespace: pod.Namespace,
					Name:      fmt.Sprintf("%s/%s", pod.Name, container.Name),
				}},
				Risk: "Without memory limits, a container can consume all available memory, potentially causing node OOM kills",
				Fix: fmt.Sprintf(`kubectl set resources pod %s -n %s --limits=memory=256Mi --container=%s`,
					pod.Name, pod.Namespace, container.Name),
			})
		}

		// RES005: CPU limit equals request
		if !container.Resources.Requests.Cpu().IsZero() &&
			!container.Resources.Limits.Cpu().IsZero() &&
			container.Resources.Requests.Cpu().Cmp(*container.Resources.Limits.Cpu()) == 0 {
			if s.exemptions.IsExemptPod(ctx, &pod, "RES005") {
				continue
			}
			findings = append(findings, types.Finding{
				ID:          "RES005",
				Title:       "CPU limit equals request",
				Description: fmt.Sprintf("Container %s in pod %s has CPU limit equal to request", container.Name, pod.Name),
				Severity:    types.SeverityInfo,
				Category:    "resources",
				Affected: []types.AffectedResource{{
					Kind:      "Pod",
					Namespace: pod.Namespace,
					Name:      fmt.Sprintf("%s/%s", pod.Name, container.Name),
				}},
				Risk: "Pod has no burstable CPU capacity and cannot handle temporary load spikes",
				Fix:  "Consider setting CPU limit higher than request to allow burst capacity",
			})
		}
	}

	return findings
}
