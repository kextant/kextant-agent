package scanner

import (
	"context"
	"fmt"

	"github.com/kextant/kextant-agent/pkg/types"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type deprecatedAPIResource struct {
	group      string
	version    string
	resource   string
	kind       string
	namespaced bool
	deprecated string
	removed    string
	current    string
}

// Deprecated and removed API resources that Kextant probes through their legacy API endpoints.
// A typed current-version client cannot preserve the API version a resource was applied through.
// Dynamic legacy endpoint probes detect resources only when the deprecated API is still served;
// removed APIs naturally return NotFound and are skipped.
var deprecatedAPIResources = []deprecatedAPIResource{
	{
		group:      "extensions",
		version:    "v1beta1",
		resource:   "ingresses",
		kind:       "Ingress",
		namespaced: true,
		deprecated: "1.14",
		removed:    "1.22",
		current:    "networking.k8s.io/v1",
	},
	{
		group:      "networking.k8s.io",
		version:    "v1beta1",
		resource:   "ingresses",
		kind:       "Ingress",
		namespaced: true,
		deprecated: "1.19",
		removed:    "1.22",
		current:    "networking.k8s.io/v1",
	},
	{
		group:      "networking.k8s.io",
		version:    "v1beta1",
		resource:   "ingressclasses",
		kind:       "IngressClass",
		deprecated: "1.19",
		removed:    "1.22",
		current:    "networking.k8s.io/v1",
	},
	{
		group:      "policy",
		version:    "v1beta1",
		resource:   "poddisruptionbudgets",
		kind:       "PodDisruptionBudget",
		namespaced: true,
		deprecated: "1.21",
		removed:    "1.25",
		current:    "policy/v1",
	},
	{
		group:      "policy",
		version:    "v1beta1",
		resource:   "podsecuritypolicies",
		kind:       "PodSecurityPolicy",
		deprecated: "1.21",
		removed:    "1.25",
		current:    "Pod Security Admission",
	},
	{
		group:      "batch",
		version:    "v1beta1",
		resource:   "cronjobs",
		kind:       "CronJob",
		namespaced: true,
		deprecated: "1.21",
		removed:    "1.25",
		current:    "batch/v1",
	},
	{
		group:      "autoscaling",
		version:    "v2beta1",
		resource:   "horizontalpodautoscalers",
		kind:       "HorizontalPodAutoscaler",
		namespaced: true,
		deprecated: "1.22",
		removed:    "1.26",
		current:    "autoscaling/v2",
	},
	{
		group:      "autoscaling",
		version:    "v2beta2",
		resource:   "horizontalpodautoscalers",
		kind:       "HorizontalPodAutoscaler",
		namespaced: true,
		deprecated: "1.23",
		removed:    "1.26",
		current:    "autoscaling/v2",
	},
}

func (s *Scanner) checkDeprecatedAPIs(ctx context.Context, namespaces []string) ([]types.Finding, error) {
	if s.dynamicClient == nil {
		s.logger.Warn("dynamic client unavailable; skipping deprecated API checks")
		return nil, nil
	}

	var findings []types.Finding
	for _, apiResource := range deprecatedAPIResources {
		resourceFindings, err := s.checkDeprecatedAPIResource(ctx, apiResource, namespaces)
		if err != nil {
			return nil, err
		}
		findings = append(findings, resourceFindings...)
	}
	return findings, nil
}

func (s *Scanner) checkDeprecatedAPIResource(ctx context.Context, apiResource deprecatedAPIResource, namespaces []string) ([]types.Finding, error) {
	gvr := schema.GroupVersionResource{
		Group:    apiResource.group,
		Version:  apiResource.version,
		Resource: apiResource.resource,
	}

	if apiResource.namespaced {
		var findings []types.Finding
		for _, namespace := range namespaces {
			items, err := s.dynamicClient.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
			if isDeprecatedAPIProbeUnavailable(err) {
				continue
			}
			if err != nil {
				return nil, fmt.Errorf("list %s %s in namespace %s: %w", apiResource.version, apiResource.resource, namespace, err)
			}
			for _, item := range items.Items {
				if s.exemptions.IsExemptResource(item.GetAnnotations(), "API001") {
					continue
				}
				findings = append(findings, deprecatedAPIFinding(apiResource, item.GetNamespace(), item.GetName()))
			}
		}
		return findings, nil
	}

	items, err := s.dynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
	if isDeprecatedAPIProbeUnavailable(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list %s %s: %w", apiResource.version, apiResource.resource, err)
	}

	findings := make([]types.Finding, 0, len(items.Items))
	for _, item := range items.Items {
		if s.exemptions.IsExemptResource(item.GetAnnotations(), "API001") {
			continue
		}
		findings = append(findings, deprecatedAPIFinding(apiResource, item.GetNamespace(), item.GetName()))
	}
	return findings, nil
}

func deprecatedAPIFinding(apiResource deprecatedAPIResource, namespace, name string) types.Finding {
	apiVersion := apiResource.version
	if apiResource.group != "" {
		apiVersion = apiResource.group + "/" + apiResource.version
	}

	return types.Finding{
		ID:          "API001",
		Title:       "Deprecated API version",
		Description: fmt.Sprintf("%s %s uses deprecated API version %s", apiResource.kind, name, apiVersion),
		Severity:    types.SeverityCritical,
		Category:    "deprecated-api",
		Affected: []types.AffectedResource{{
			Kind:      apiResource.kind,
			Namespace: namespace,
			Name:      name,
		}},
		Risk: fmt.Sprintf("This API version was deprecated in Kubernetes %s and removed in %s", apiResource.deprecated, apiResource.removed),
		Fix:  fmt.Sprintf("Migrate to %s API version", apiResource.current),
	}
}

func isDeprecatedAPIProbeUnavailable(err error) bool {
	return apierrors.IsNotFound(err) || apierrors.IsGone(err) || apierrors.IsMethodNotSupported(err)
}
