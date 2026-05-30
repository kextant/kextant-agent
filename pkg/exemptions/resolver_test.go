package exemptions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestParseAnnotation(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		expected    []string
	}{
		{
			name:        "single check ID",
			annotations: map[string]string{AnnotationKey: "RES001"},
			expected:    []string{"RES001"},
		},
		{
			name:        "multiple check IDs",
			annotations: map[string]string{AnnotationKey: "RES001,PRB001,REP001"},
			expected:    []string{"RES001", "PRB001", "REP001"},
		},
		{
			name:        "whitespace trimming",
			annotations: map[string]string{AnnotationKey: " RES001 , PRB001 , REP001 "},
			expected:    []string{"RES001", "PRB001", "REP001"},
		},
		{
			name:        "wildcard",
			annotations: map[string]string{AnnotationKey: "*"},
			expected:    []string{"*"},
		},
		{
			name:        "missing annotation",
			annotations: map[string]string{},
			expected:    []string{},
		},
		{
			name:        "empty annotation",
			annotations: map[string]string{AnnotationKey: ""},
			expected:    []string{},
		},
		{
			name:        "empty entries filtered",
			annotations: map[string]string{AnnotationKey: "RES001,,PRB001"},
			expected:    []string{"RES001", "PRB001"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAnnotation(tt.annotations)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsExempt_DirectMatch(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		checkID     string
		expected    bool
	}{
		{
			name:        "exact match",
			annotations: map[string]string{AnnotationKey: "RES001"},
			checkID:     "RES001",
			expected:    true,
		},
		{
			name:        "no match",
			annotations: map[string]string{AnnotationKey: "RES001"},
			checkID:     "RES002",
			expected:    false,
		},
		{
			name:        "wildcard matches all",
			annotations: map[string]string{AnnotationKey: "*"},
			checkID:     "RES001",
			expected:    true,
		},
		{
			name:        "multiple checks with match",
			annotations: map[string]string{AnnotationKey: "RES001,PRB001,REP001"},
			checkID:     "PRB001",
			expected:    true,
		},
		{
			name:        "multiple checks without match",
			annotations: map[string]string{AnnotationKey: "RES001,PRB001,REP001"},
			checkID:     "IMG001",
			expected:    false,
		},
		{
			name:        "no annotation",
			annotations: map[string]string{},
			checkID:     "RES001",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isExempt(tt.annotations, tt.checkID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsExemptPod_DirectAnnotation(t *testing.T) {
	client := fake.NewSimpleClientset()
	resolver := NewResolver(client, nil)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			Annotations: map[string]string{
				AnnotationKey: "RES001,PRB001",
			},
		},
	}

	ctx := context.Background()

	// Should be exempt for RES001
	assert.True(t, resolver.IsExemptPod(ctx, pod, "RES001"))

	// Should be exempt for PRB001
	assert.True(t, resolver.IsExemptPod(ctx, pod, "PRB001"))

	// Should NOT be exempt for RES002
	assert.False(t, resolver.IsExemptPod(ctx, pod, "RES002"))
}

func TestIsExemptPod_InheritedFromDeployment(t *testing.T) {
	// Create a deployment with exemptions
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "default",
			UID:       "deploy-uid-123",
			Annotations: map[string]string{
				AnnotationKey: "REP001,RES001",
			},
		},
	}

	// Create a replicaset owned by the deployment
	replicaset := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-rs",
			Namespace: "default",
			UID:       "rs-uid-456",
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "test-deployment",
					UID:        "deploy-uid-123",
				},
			},
		},
	}

	// Create a pod owned by the replicaset
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "ReplicaSet",
					Name:       "test-rs",
					UID:        "rs-uid-456",
				},
			},
		},
	}

	client := fake.NewSimpleClientset(deployment, replicaset, pod)
	resolver := NewResolver(client, nil)
	ctx := context.Background()

	// Should inherit REP001 from deployment
	assert.True(t, resolver.IsExemptPod(ctx, pod, "REP001"))

	// Should inherit RES001 from deployment
	assert.True(t, resolver.IsExemptPod(ctx, pod, "RES001"))

	// Should NOT be exempt for PRB001 (not in deployment)
	assert.False(t, resolver.IsExemptPod(ctx, pod, "PRB001"))
}

func TestIsExemptPod_CombinedExemptions(t *testing.T) {
	// Create a deployment with some exemptions
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "default",
			UID:       "deploy-uid-123",
			Annotations: map[string]string{
				AnnotationKey: "REP001",
			},
		},
	}

	replicaset := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-rs",
			Namespace: "default",
			UID:       "rs-uid-456",
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "test-deployment",
					UID:        "deploy-uid-123",
				},
			},
		},
	}

	// Create a pod with its own exemptions
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			Annotations: map[string]string{
				AnnotationKey: "RES001,PRB001",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "ReplicaSet",
					Name:       "test-rs",
					UID:        "rs-uid-456",
				},
			},
		},
	}

	client := fake.NewSimpleClientset(deployment, replicaset, pod)
	resolver := NewResolver(client, nil)
	ctx := context.Background()

	// Should have REP001 from deployment
	assert.True(t, resolver.IsExemptPod(ctx, pod, "REP001"))

	// Should have RES001 from pod itself
	assert.True(t, resolver.IsExemptPod(ctx, pod, "RES001"))

	// Should have PRB001 from pod itself
	assert.True(t, resolver.IsExemptPod(ctx, pod, "PRB001"))

	// Should NOT have IMG001 (not in either)
	assert.False(t, resolver.IsExemptPod(ctx, pod, "IMG001"))
}

func TestIsExemptPod_StandalonePod(t *testing.T) {
	// Pod without owner references
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "standalone-pod",
			Namespace: "default",
			Annotations: map[string]string{
				AnnotationKey: "RES001",
			},
		},
	}

	client := fake.NewSimpleClientset(pod)
	resolver := NewResolver(client, nil)
	ctx := context.Background()

	// Should have its own exemption
	assert.True(t, resolver.IsExemptPod(ctx, pod, "RES001"))

	// Should NOT have others
	assert.False(t, resolver.IsExemptPod(ctx, pod, "PRB001"))
}

func TestIsExemptPod_WildcardInheritance(t *testing.T) {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "default",
			UID:       "deploy-uid-123",
			Annotations: map[string]string{
				AnnotationKey: "*",
			},
		},
	}

	replicaset := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-rs",
			Namespace: "default",
			UID:       "rs-uid-456",
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "test-deployment",
					UID:        "deploy-uid-123",
				},
			},
		},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "ReplicaSet",
					Name:       "test-rs",
					UID:        "rs-uid-456",
				},
			},
		},
	}

	client := fake.NewSimpleClientset(deployment, replicaset, pod)
	resolver := NewResolver(client, nil)
	ctx := context.Background()

	// Should inherit wildcard and exempt everything
	assert.True(t, resolver.IsExemptPod(ctx, pod, "RES001"))
	assert.True(t, resolver.IsExemptPod(ctx, pod, "PRB001"))
	assert.True(t, resolver.IsExemptPod(ctx, pod, "ANY_CHECK"))
}

func TestIsExemptDeployment(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		checkID     string
		expected    bool
	}{
		{
			name:        "deployment with exemption",
			annotations: map[string]string{AnnotationKey: "REP001"},
			checkID:     "REP001",
			expected:    true,
		},
		{
			name:        "deployment without exemption",
			annotations: map[string]string{AnnotationKey: "REP001"},
			checkID:     "REP002",
			expected:    false,
		},
		{
			name:        "deployment with wildcard",
			annotations: map[string]string{AnnotationKey: "*"},
			checkID:     "REP001",
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewSimpleClientset()
			resolver := NewResolver(client, nil)

			deploy := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-deploy",
					Namespace:   "default",
					Annotations: tt.annotations,
				},
			}

			result := resolver.IsExemptDeployment(deploy, tt.checkID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsExemptNamespace(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		checkID     string
		expected    bool
	}{
		{
			name:        "namespace with exemption",
			annotations: map[string]string{AnnotationKey: "NS001"},
			checkID:     "NS001",
			expected:    true,
		},
		{
			name:        "namespace without exemption",
			annotations: map[string]string{AnnotationKey: "NS001"},
			checkID:     "NS002",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewSimpleClientset()
			resolver := NewResolver(client, nil)

			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-ns",
					Annotations: tt.annotations,
				},
			}

			result := resolver.IsExemptNamespace(ns, tt.checkID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCacheClearing(t *testing.T) {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "default",
			UID:       "deploy-uid-123",
			Annotations: map[string]string{
				AnnotationKey: "REP001",
			},
		},
	}

	replicaset := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-rs",
			Namespace: "default",
			UID:       "rs-uid-456",
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "test-deployment",
					UID:        "deploy-uid-123",
				},
			},
		},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "ReplicaSet",
					Name:       "test-rs",
					UID:        "rs-uid-456",
				},
			},
		},
	}

	client := fake.NewSimpleClientset(deployment, replicaset, pod)
	resolver := NewResolver(client, nil)
	ctx := context.Background()

	// First call - should populate cache
	result1 := resolver.IsExemptPod(ctx, pod, "REP001")
	assert.True(t, result1)

	// Verify cache is populated
	require.NotEmpty(t, resolver.cache)

	// Clear cache
	resolver.ClearCache()

	// Verify cache is empty
	assert.Empty(t, resolver.cache)

	// Second call after clear - should still work
	result2 := resolver.IsExemptPod(ctx, pod, "REP001")
	assert.True(t, result2)
}

func TestBrokenOwnerChain(t *testing.T) {
	// Pod references a ReplicaSet that doesn't exist
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "orphan-pod",
			Namespace: "default",
			Annotations: map[string]string{
				AnnotationKey: "RES001",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "ReplicaSet",
					Name:       "missing-rs",
					UID:        "missing-uid",
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pod)
	resolver := NewResolver(client, nil)
	ctx := context.Background()

	// Should still work with pod's own exemptions
	assert.True(t, resolver.IsExemptPod(ctx, pod, "RES001"))
	assert.False(t, resolver.IsExemptPod(ctx, pod, "PRB001"))
}

func TestGlobalExemptions(t *testing.T) {
	client := fake.NewSimpleClientset()

	// Test with specific global exemptions
	resolver := NewResolver(client, []string{"RES001", "PRB001"})
	ctx := context.Background()

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
	}

	// Should be exempt due to global exemptions
	assert.True(t, resolver.IsExemptPod(ctx, pod, "RES001"))
	assert.True(t, resolver.IsExemptPod(ctx, pod, "PRB001"))

	// Should NOT be exempt (not in global list)
	assert.False(t, resolver.IsExemptPod(ctx, pod, "RES002"))
}

func TestGlobalWildcardExemption(t *testing.T) {
	client := fake.NewSimpleClientset()

	// Test with wildcard global exemption
	resolver := NewResolver(client, []string{"*"})
	ctx := context.Background()

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
	}

	// Everything should be exempt
	assert.True(t, resolver.IsExemptPod(ctx, pod, "RES001"))
	assert.True(t, resolver.IsExemptPod(ctx, pod, "PRB001"))
	assert.True(t, resolver.IsExemptPod(ctx, pod, "ANY_CHECK"))
}

func TestGlobalExemptionPriority(t *testing.T) {
	client := fake.NewSimpleClientset()

	// Global exemptions should work even without resource annotations
	resolver := NewResolver(client, []string{"RES001"})

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deploy",
			Namespace: "default",
			// No annotations
		},
	}

	// Should still be exempt due to global config
	assert.True(t, resolver.IsExemptDeployment(deployment, "RES001"))
	assert.False(t, resolver.IsExemptDeployment(deployment, "RES002"))
}

func TestGlobalExemptionWithWhitespace(t *testing.T) {
	client := fake.NewSimpleClientset()

	// Test trimming whitespace in global exemptions
	resolver := NewResolver(client, []string{" RES001 ", "PRB001", "  REP001  "})
	ctx := context.Background()

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
	}

	// All should work despite whitespace
	assert.True(t, resolver.IsExemptPod(ctx, pod, "RES001"))
	assert.True(t, resolver.IsExemptPod(ctx, pod, "PRB001"))
	assert.True(t, resolver.IsExemptPod(ctx, pod, "REP001"))
}
