package exemptions

import (
	"context"
	"strings"
	"sync"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// AnnotationKey is the annotation key used to specify exempted checks
	AnnotationKey = "kextant.com/skip-checks"
)

// Resolver handles exemption logic for validation checks
type Resolver struct {
	client        kubernetes.Interface
	cache         map[string][]string
	mu            sync.RWMutex
	globalExempts map[string]bool
}

// NewResolver creates a new exemption resolver
func NewResolver(client kubernetes.Interface, globalExempts []string) *Resolver {
	exemptMap := make(map[string]bool)
	for _, checkID := range globalExempts {
		trimmed := strings.TrimSpace(checkID)
		if trimmed != "" {
			exemptMap[trimmed] = true
		}
	}

	return &Resolver{
		client:        client,
		cache:         make(map[string][]string),
		globalExempts: exemptMap,
	}
}

// ClearCache clears the internal owner reference cache
// Should be called at the start of each scan
func (r *Resolver) ClearCache() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache = make(map[string][]string)
}

// IsExemptPod checks if a pod is exempt from a specific check
// It considers global exemptions, the pod's direct annotations, and inherited exemptions from owners
func (r *Resolver) IsExemptPod(ctx context.Context, pod *corev1.Pod, checkID string) bool {
	// Check global exemptions first
	if r.globalExempts["*"] || r.globalExempts[checkID] {
		return true
	}

	// Check pod's direct annotations
	if isExempt(pod.Annotations, checkID) {
		return true
	}

	// Check inherited exemptions from owner chain
	if len(pod.OwnerReferences) > 0 {
		inherited := r.resolveOwnerExemptions(ctx, pod.OwnerReferences, pod.Namespace)
		for _, exemptedID := range inherited {
			if exemptedID == "*" || exemptedID == checkID {
				return true
			}
		}
	}

	return false
}

// IsExemptDeployment checks if a deployment is exempt from a specific check
func (r *Resolver) IsExemptDeployment(deployment *appsv1.Deployment, checkID string) bool {
	// Check global exemptions first
	if r.globalExempts["*"] || r.globalExempts[checkID] {
		return true
	}
	return isExempt(deployment.Annotations, checkID)
}

// IsExemptNamespace checks if a namespace is exempt from a specific check
func (r *Resolver) IsExemptNamespace(namespace *corev1.Namespace, checkID string) bool {
	// Check global exemptions first
	if r.globalExempts["*"] || r.globalExempts[checkID] {
		return true
	}
	return isExempt(namespace.Annotations, checkID)
}

// IsExemptResource checks if a resource with annotations is exempt from a specific check
func (r *Resolver) IsExemptResource(annotations map[string]string, checkID string) bool {
	// Check global exemptions first
	if r.globalExempts["*"] || r.globalExempts[checkID] {
		return true
	}
	return isExempt(annotations, checkID)
}

// isExempt checks if a checkID is exempted in the given annotations
func isExempt(annotations map[string]string, checkID string) bool {
	exemptions := parseAnnotation(annotations)
	for _, exemptedID := range exemptions {
		if exemptedID == "*" || exemptedID == checkID {
			return true
		}
	}
	return false
}

// parseAnnotation parses the exemption annotation into a list of check IDs
func parseAnnotation(annotations map[string]string) []string {
	value, exists := annotations[AnnotationKey]
	if !exists || value == "" {
		return []string{}
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// resolveOwnerExemptions resolves exemptions from the owner reference chain
// It follows the chain: Pod -> ReplicaSet -> Deployment
func (r *Resolver) resolveOwnerExemptions(ctx context.Context, ownerRefs []metav1.OwnerReference, namespace string) []string {
	var allExemptions []string

	for _, ownerRef := range ownerRefs {
		// Check cache first
		cacheKey := namespace + "/" + string(ownerRef.UID)

		r.mu.RLock()
		cached, found := r.cache[cacheKey]
		r.mu.RUnlock()

		if found {
			allExemptions = append(allExemptions, cached...)
			continue
		}

		var exemptions []string

		switch ownerRef.Kind {
		case "ReplicaSet":
			exemptions = r.getReplicaSetExemptions(ctx, ownerRef.Name, namespace)
		case "Deployment":
			exemptions = r.getDeploymentExemptions(ctx, ownerRef.Name, namespace)
		}

		// Cache the result
		r.mu.Lock()
		r.cache[cacheKey] = exemptions
		r.mu.Unlock()

		allExemptions = append(allExemptions, exemptions...)
	}

	return allExemptions
}

// getReplicaSetExemptions gets exemptions from a ReplicaSet and its owners
func (r *Resolver) getReplicaSetExemptions(ctx context.Context, name, namespace string) []string {
	rs, err := r.client.AppsV1().ReplicaSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		// If we can't get the ReplicaSet, just return empty
		return []string{}
	}

	// Get direct exemptions from ReplicaSet
	exemptions := parseAnnotation(rs.Annotations)

	// Get exemptions from ReplicaSet's owners (typically Deployment)
	if len(rs.OwnerReferences) > 0 {
		inherited := r.resolveOwnerExemptions(ctx, rs.OwnerReferences, namespace)
		exemptions = append(exemptions, inherited...)
	}

	return exemptions
}

// getDeploymentExemptions gets exemptions from a Deployment
func (r *Resolver) getDeploymentExemptions(ctx context.Context, name, namespace string) []string {
	deploy, err := r.client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		// If we can't get the Deployment, just return empty
		return []string{}
	}

	return parseAnnotation(deploy.Annotations)
}
