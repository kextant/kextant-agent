package scanner

import (
	"context"
	"fmt"
	"reflect"

	"github.com/kextant/kextant-agent/pkg/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Scanner) checkProbes(ctx context.Context, namespaces []string) ([]types.Finding, error) {
	var findings []types.Finding

	for _, ns := range namespaces {
		pods, err := s.client.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		for _, pod := range pods.Items {
			findings = append(findings, s.checkPodProbes(ctx, pod)...)
		}
	}

	return findings, nil
}

func (s *Scanner) checkPodProbes(ctx context.Context, pod corev1.Pod) []types.Finding {
	var findings []types.Finding

	for _, container := range pod.Spec.Containers {
		// PRB001: Missing readiness probe
		if container.ReadinessProbe == nil {
			if s.exemptions.IsExemptPod(ctx, &pod, "PRB001") {
				continue
			}
			findings = append(findings, types.Finding{
				ID:          "PRB001",
				Title:       "Missing readiness probe",
				Description: fmt.Sprintf("Container %s in pod %s has no readiness probe", container.Name, pod.Name),
				Severity:    types.SeverityWarning,
				Category:    "probes",
				Affected: []types.AffectedResource{{
					Kind:      "Pod",
					Namespace: pod.Namespace,
					Name:      fmt.Sprintf("%s/%s", pod.Name, container.Name),
				}},
				Risk: "Without a readiness probe, the container may receive traffic before it's ready, causing errors",
				Fix:  "Add a readiness probe to ensure the container is ready before receiving traffic",
			})
		}

		// PRB002: Missing liveness probe
		if container.LivenessProbe == nil {
			if s.exemptions.IsExemptPod(ctx, &pod, "PRB002") {
				continue
			}
			findings = append(findings, types.Finding{
				ID:          "PRB002",
				Title:       "Missing liveness probe",
				Description: fmt.Sprintf("Container %s in pod %s has no liveness probe", container.Name, pod.Name),
				Severity:    types.SeverityWarning,
				Category:    "probes",
				Affected: []types.AffectedResource{{
					Kind:      "Pod",
					Namespace: pod.Namespace,
					Name:      fmt.Sprintf("%s/%s", pod.Name, container.Name),
				}},
				Risk: "Without a liveness probe, Kubernetes cannot detect and restart unhealthy containers",
				Fix:  "Add a liveness probe to enable automatic recovery from deadlocks",
			})
		}

		// PRB003: Missing startup probe
		if container.StartupProbe == nil {
			if s.exemptions.IsExemptPod(ctx, &pod, "PRB003") {
				continue
			}
			findings = append(findings, types.Finding{
				ID:          "PRB003",
				Title:       "Missing startup probe",
				Description: fmt.Sprintf("Container %s in pod %s has no startup probe", container.Name, pod.Name),
				Severity:    types.SeverityInfo,
				Category:    "probes",
				Affected: []types.AffectedResource{{
					Kind:      "Pod",
					Namespace: pod.Namespace,
					Name:      fmt.Sprintf("%s/%s", pod.Name, container.Name),
				}},
				Risk: "For slow-starting applications, liveness probes may kill the container during startup",
				Fix:  "Add a startup probe for applications that take time to initialize",
			})
		}

		// PRB004: Identical liveness/readiness probes
		if container.LivenessProbe != nil && container.ReadinessProbe != nil {
			if reflect.DeepEqual(container.LivenessProbe, container.ReadinessProbe) {
				if s.exemptions.IsExemptPod(ctx, &pod, "PRB004") {
					continue
				}
				findings = append(findings, types.Finding{
					ID:          "PRB004",
					Title:       "Identical liveness/readiness probes",
					Description: fmt.Sprintf("Container %s in pod %s has identical liveness and readiness probes", container.Name, pod.Name),
					Severity:    types.SeverityInfo,
					Category:    "probes",
					Affected: []types.AffectedResource{{
						Kind:      "Pod",
						Namespace: pod.Namespace,
						Name:      fmt.Sprintf("%s/%s", pod.Name, container.Name),
					}},
					Risk: "Liveness and readiness probes serve different purposes and should typically check different conditions",
					Fix:  "Use readiness probes for traffic-ready checks and liveness probes for deadlock detection",
				})
			}
		}
	}

	return findings
}
