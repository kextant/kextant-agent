package scanner

import (
	"context"
	"fmt"

	"github.com/kextant/kextant-agent/pkg/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Scanner) checkSecurity(ctx context.Context, namespaces []string) ([]types.Finding, error) {
	var findings []types.Finding

	for _, ns := range namespaces {
		pods, err := s.client.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		for _, pod := range pods.Items {
			findings = append(findings, s.checkPodSecurity(ctx, pod)...)
		}
	}

	return findings, nil
}

func (s *Scanner) checkPodSecurity(ctx context.Context, pod corev1.Pod) []types.Finding {
	var findings []types.Finding

	for _, container := range pod.Spec.Containers {
		sc := container.SecurityContext
		podSC := pod.Spec.SecurityContext

		// SEC001: Running as root
		runAsRoot := true
		if sc != nil && sc.RunAsNonRoot != nil && *sc.RunAsNonRoot {
			runAsRoot = false
		} else if sc != nil && sc.RunAsUser != nil && *sc.RunAsUser != 0 {
			runAsRoot = false
		} else if podSC != nil && podSC.RunAsNonRoot != nil && *podSC.RunAsNonRoot {
			runAsRoot = false
		} else if podSC != nil && podSC.RunAsUser != nil && *podSC.RunAsUser != 0 {
			runAsRoot = false
		}

		if runAsRoot {
			if s.exemptions.IsExemptPod(ctx, &pod, "SEC001") {
				continue
			}
			findings = append(findings, types.Finding{
				ID:          "SEC001",
				Title:       "Running as root",
				Description: fmt.Sprintf("Container %s in pod %s may run as root user", container.Name, pod.Name),
				Severity:    types.SeverityWarning,
				Category:    "security",
				Affected: []types.AffectedResource{{
					Kind:      "Pod",
					Namespace: pod.Namespace,
					Name:      fmt.Sprintf("%s/%s", pod.Name, container.Name),
				}},
				Risk: "Running as root increases security risk if the container is compromised",
				Fix:  "Set runAsNonRoot: true or runAsUser to a non-zero value in securityContext",
			})
		}

		// SEC002: Privileged container
		if sc != nil && sc.Privileged != nil && *sc.Privileged {
			if s.exemptions.IsExemptPod(ctx, &pod, "SEC002") {
				continue
			}
			findings = append(findings, types.Finding{
				ID:          "SEC002",
				Title:       "Privileged container",
				Description: fmt.Sprintf("Container %s in pod %s runs in privileged mode", container.Name, pod.Name),
				Severity:    types.SeverityCritical,
				Category:    "security",
				Affected: []types.AffectedResource{{
					Kind:      "Pod",
					Namespace: pod.Namespace,
					Name:      fmt.Sprintf("%s/%s", pod.Name, container.Name),
				}},
				Risk: "Privileged containers have unrestricted access to host resources and can compromise the entire node",
				Fix:  "Remove privileged: true from securityContext unless absolutely necessary",
			})
		}

		// SEC003: Root filesystem writable
		if sc == nil || sc.ReadOnlyRootFilesystem == nil || !*sc.ReadOnlyRootFilesystem {
			if s.exemptions.IsExemptPod(ctx, &pod, "SEC003") {
				continue
			}
			findings = append(findings, types.Finding{
				ID:          "SEC003",
				Title:       "Root filesystem writable",
				Description: fmt.Sprintf("Container %s in pod %s has writable root filesystem", container.Name, pod.Name),
				Severity:    types.SeverityWarning,
				Category:    "security",
				Affected: []types.AffectedResource{{
					Kind:      "Pod",
					Namespace: pod.Namespace,
					Name:      fmt.Sprintf("%s/%s", pod.Name, container.Name),
				}},
				Risk: "Writable root filesystem allows malicious code to modify the container filesystem",
				Fix:  "Set readOnlyRootFilesystem: true in securityContext and use volumes for writable directories",
			})
		}

		// SEC004: Privilege escalation allowed
		if sc == nil || sc.AllowPrivilegeEscalation == nil || *sc.AllowPrivilegeEscalation {
			if s.exemptions.IsExemptPod(ctx, &pod, "SEC004") {
				continue
			}
			findings = append(findings, types.Finding{
				ID:          "SEC004",
				Title:       "Privilege escalation allowed",
				Description: fmt.Sprintf("Container %s in pod %s allows privilege escalation", container.Name, pod.Name),
				Severity:    types.SeverityWarning,
				Category:    "security",
				Affected: []types.AffectedResource{{
					Kind:      "Pod",
					Namespace: pod.Namespace,
					Name:      fmt.Sprintf("%s/%s", pod.Name, container.Name),
				}},
				Risk: "Allowing privilege escalation enables processes to gain more privileges than their parent",
				Fix:  "Set allowPrivilegeEscalation: false in securityContext",
			})
		}

		// SEC005: All capabilities not dropped
		droppedAll := false
		if sc != nil && sc.Capabilities != nil {
			for _, cap := range sc.Capabilities.Drop {
				if cap == "ALL" {
					droppedAll = true
					break
				}
			}
		}

		if !droppedAll {
			if s.exemptions.IsExemptPod(ctx, &pod, "SEC005") {
				continue
			}
			findings = append(findings, types.Finding{
				ID:          "SEC005",
				Title:       "All capabilities not dropped",
				Description: fmt.Sprintf("Container %s in pod %s doesn't drop all capabilities", container.Name, pod.Name),
				Severity:    types.SeverityInfo,
				Category:    "security",
				Affected: []types.AffectedResource{{
					Kind:      "Pod",
					Namespace: pod.Namespace,
					Name:      fmt.Sprintf("%s/%s", pod.Name, container.Name),
				}},
				Risk: "Containers inherit Linux capabilities that may not be needed, increasing attack surface",
				Fix:  "Drop all capabilities and add back only what's needed: capabilities: { drop: [ALL] }",
			})
		}
	}

	return findings
}
