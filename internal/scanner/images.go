package scanner

import (
	"context"
	"fmt"
	"strings"

	"github.com/kextant/kextant-agent/pkg/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Scanner) checkImages(ctx context.Context, namespaces []string) ([]types.Finding, error) {
	var findings []types.Finding

	for _, ns := range namespaces {
		pods, err := s.client.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		for _, pod := range pods.Items {
			findings = append(findings, s.checkPodImages(ctx, pod)...)
		}
	}

	return findings, nil
}

func (s *Scanner) checkPodImages(ctx context.Context, pod corev1.Pod) []types.Finding {
	var findings []types.Finding

	for _, container := range pod.Spec.Containers {
		image := container.Image

		// IMG001: Latest tag used
		if strings.HasSuffix(image, ":latest") {
			if s.exemptions.IsExemptPod(ctx, &pod, "IMG001") {
				continue
			}
			findings = append(findings, types.Finding{
				ID:          "IMG001",
				Title:       "Latest tag used",
				Description: fmt.Sprintf("Container %s in pod %s uses :latest tag", container.Name, pod.Name),
				Severity:    types.SeverityWarning,
				Category:    "images",
				Affected: []types.AffectedResource{{
					Kind:      "Pod",
					Namespace: pod.Namespace,
					Name:      fmt.Sprintf("%s/%s", pod.Name, container.Name),
				}},
				Risk: "Using :latest tag makes deployments unpredictable and can cause unexpected behavior",
				Fix:  "Use specific version tags for reproducible deployments",
			})
		}

		// IMG002: No image tag
		if !strings.Contains(image, ":") {
			if s.exemptions.IsExemptPod(ctx, &pod, "IMG002") {
				continue
			}
			findings = append(findings, types.Finding{
				ID:          "IMG002",
				Title:       "No image tag",
				Description: fmt.Sprintf("Container %s in pod %s has no tag specified", container.Name, pod.Name),
				Severity:    types.SeverityWarning,
				Category:    "images",
				Affected: []types.AffectedResource{{
					Kind:      "Pod",
					Namespace: pod.Namespace,
					Name:      fmt.Sprintf("%s/%s", pod.Name, container.Name),
				}},
				Risk: "Images without tags default to :latest, making deployments unpredictable",
				Fix:  "Always specify an explicit image tag",
			})
		}

		// IMG003: ImagePullPolicy not set
		if container.ImagePullPolicy == "" {
			if s.exemptions.IsExemptPod(ctx, &pod, "IMG003") {
				continue
			}
			findings = append(findings, types.Finding{
				ID:          "IMG003",
				Title:       "ImagePullPolicy not set",
				Description: fmt.Sprintf("Container %s in pod %s has no ImagePullPolicy set", container.Name, pod.Name),
				Severity:    types.SeverityInfo,
				Category:    "images",
				Affected: []types.AffectedResource{{
					Kind:      "Pod",
					Namespace: pod.Namespace,
					Name:      fmt.Sprintf("%s/%s", pod.Name, container.Name),
				}},
				Risk: "ImagePullPolicy defaults may not match your expectations",
				Fix:  "Explicitly set ImagePullPolicy to IfNotPresent or Always",
			})
		}

		// IMG004: ImagePullPolicy Always with tag
		if container.ImagePullPolicy == corev1.PullAlways &&
			!strings.HasSuffix(image, ":latest") &&
			strings.Contains(image, ":") {
			if s.exemptions.IsExemptPod(ctx, &pod, "IMG004") {
				continue
			}
			findings = append(findings, types.Finding{
				ID:          "IMG004",
				Title:       "ImagePullPolicy Always with tag",
				Description: fmt.Sprintf("Container %s in pod %s uses Always pull policy with a specific tag", container.Name, pod.Name),
				Severity:    types.SeverityInfo,
				Category:    "images",
				Affected: []types.AffectedResource{{
					Kind:      "Pod",
					Namespace: pod.Namespace,
					Name:      fmt.Sprintf("%s/%s", pod.Name, container.Name),
				}},
				Risk: "Using Always with a specific tag causes unnecessary image pulls and slower pod starts",
				Fix:  "Use IfNotPresent for tagged images to improve performance",
			})
		}
	}

	return findings
}
