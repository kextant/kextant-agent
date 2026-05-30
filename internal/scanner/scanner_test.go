package scanner

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/kextant/kextant-agent/internal/config"
	"github.com/kextant/kextant-agent/pkg/exemptions"
	"github.com/kextant/kextant-agent/pkg/types"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
)

func TestScanRunsEnabledChecksAndFiltersNamespaces(t *testing.T) {
	s := testScanner(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kube-system"}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "bad", Namespace: "default"}, Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app", Image: "nginx:latest"}}}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "ignored", Namespace: "kube-system"}, Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "sys", Image: "nginx:latest"}}}},
	)

	findings, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, findings)
	for _, finding := range findings {
		for _, affected := range finding.Affected {
			require.NotEqual(t, "kube-system", affected.Namespace)
		}
	}
	require.Contains(t, findingIDs(findings), "IMG001")
}

func TestCheckPodResources(t *testing.T) {
	s := testScanner()
	pod := corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "default"}, Spec: corev1.PodSpec{Containers: []corev1.Container{{
		Name:  "app",
		Image: "example.com/app:v1",
	}}}}

	findings := s.checkPodResources(context.Background(), pod)
	require.ElementsMatch(t, []string{"RES001", "RES002", "RES003", "RES004"}, findingIDs(findings))

	pod.Spec.Containers[0].Resources = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("100m"), corev1.ResourceMemory: resource.MustParse("128Mi")},
		Limits:   corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("100m"), corev1.ResourceMemory: resource.MustParse("256Mi")},
	}
	findings = s.checkPodResources(context.Background(), pod)
	require.Equal(t, []string{"RES005"}, findingIDs(findings))
}

func TestCheckPodProbes(t *testing.T) {
	s := testScanner()
	probe := &corev1.Probe{ProbeHandler: corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{Path: "/healthz", Port: intstr.FromInt(8080)}}}
	pod := corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "default"}, Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app", Image: "app:v1"}}}}

	findings := s.checkPodProbes(context.Background(), pod)
	require.ElementsMatch(t, []string{"PRB001", "PRB002", "PRB003"}, findingIDs(findings))

	pod.Spec.Containers[0].ReadinessProbe = probe
	pod.Spec.Containers[0].LivenessProbe = probe
	pod.Spec.Containers[0].StartupProbe = probe
	findings = s.checkPodProbes(context.Background(), pod)
	require.Equal(t, []string{"PRB004"}, findingIDs(findings))
}

func TestCheckPodImages(t *testing.T) {
	s := testScanner()
	pod := corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "default"}, Spec: corev1.PodSpec{Containers: []corev1.Container{
		{Name: "latest", Image: "nginx:latest"},
		{Name: "missing", Image: "redis"},
		{Name: "port-missing", Image: "registry.internal:5000/myapp", ImagePullPolicy: corev1.PullAlways},
		{Name: "always", Image: "busybox:1.36", ImagePullPolicy: corev1.PullAlways},
	}}}
	findings := s.checkPodImages(context.Background(), pod)
	require.Contains(t, findingIDs(findings), "IMG001")
	require.Equal(t, 2, countFindingID(findings, "IMG002"))
	require.Contains(t, findingIDs(findings), "IMG003")
	require.Equal(t, 1, countFindingID(findings, "IMG004"))
}

func TestCheckPodSecurity(t *testing.T) {
	s := testScanner()
	privileged := true
	allowPrivilegeEscalation := true
	pod := corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "default"}, Spec: corev1.PodSpec{Containers: []corev1.Container{{
		Name:  "app",
		Image: "app:v1",
		SecurityContext: &corev1.SecurityContext{
			Privileged:               &privileged,
			AllowPrivilegeEscalation: &allowPrivilegeEscalation,
		},
	}}}}
	findings := s.checkPodSecurity(context.Background(), pod)
	require.ElementsMatch(t, []string{"SEC001", "SEC002", "SEC003", "SEC004", "SEC005"}, findingIDs(findings))

	nonRoot := true
	readOnly := true
	allowPrivilegeEscalation = false
	pod.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{
		RunAsNonRoot:             &nonRoot,
		ReadOnlyRootFilesystem:   &readOnly,
		AllowPrivilegeEscalation: &allowPrivilegeEscalation,
		Capabilities:             &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}},
	}
	findings = s.checkPodSecurity(context.Background(), pod)
	require.Empty(t, findings)
}

func TestCheckReplicas(t *testing.T) {
	replicasOne := int32(1)
	replicasThree := int32(3)
	s := testScanner(
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "single", Namespace: "default"}, Spec: appsv1.DeploymentSpec{Replicas: &replicasOne}, Status: appsv1.DeploymentStatus{AvailableReplicas: 1}},
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "default"}, Spec: appsv1.DeploymentSpec{Replicas: &replicasThree}, Status: appsv1.DeploymentStatus{AvailableReplicas: 2}},
		&policyv1.PodDisruptionBudget{ObjectMeta: metav1.ObjectMeta{Name: "other", Namespace: "default"}},
	)
	findings, err := s.checkReplicas(context.Background(), []string{"default"})
	require.NoError(t, err)
	require.Contains(t, findingIDs(findings), "REP001")
	require.Contains(t, findingIDs(findings), "REP002")
	require.Contains(t, findingIDs(findings), "REP003")
}

func TestCheckNamespaces(t *testing.T) {
	s := testScanner(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}})
	findings, err := s.checkNamespaces(context.Background(), []string{"default"})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"NS001", "NS002", "NS003"}, findingIDs(findings))
}

func TestCheckDeprecatedAPIs(t *testing.T) {
	s := testScanner()
	legacyIngress := &unstructured.Unstructured{}
	legacyIngress.SetAPIVersion("networking.k8s.io/v1beta1")
	legacyIngress.SetKind("Ingress")
	legacyIngress.SetNamespace("default")
	legacyIngress.SetName("legacy")
	s.dynamicClient = dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), deprecatedListKinds(), legacyIngress)

	findings, err := s.checkDeprecatedAPIs(context.Background(), []string{"default"})
	require.NoError(t, err)
	require.Contains(t, findingIDs(findings), "API001")
}

func TestHealthCheck(t *testing.T) {
	s := testScanner(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}})
	require.NoError(t, s.HealthCheck(context.Background()))
}

func testScanner(objects ...runtime.Object) *Scanner {
	client := fake.NewSimpleClientset(objects...)
	cfg := &config.Config{
		ClusterName:                  "test",
		ExcludeNamespaces:            []string{"kube-system"},
		ChecksResourcesEnabled:       true,
		ChecksProbesEnabled:          true,
		ChecksReplicasEnabled:        true,
		ChecksImagesEnabled:          true,
		ChecksSecurityContextEnabled: true,
		ChecksDeprecatedAPIEnabled:   true,
		ChecksNamespaceEnabled:       true,
	}
	return &Scanner{
		client:        client,
		dynamicClient: dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), deprecatedListKinds()),
		config:        cfg,
		logger:        slog.New(slog.NewTextHandler(io.Discard, nil)),
		exemptions:    exemptions.NewResolver(client, nil),
	}
}

func deprecatedListKinds() map[schema.GroupVersionResource]string {
	listKinds := make(map[schema.GroupVersionResource]string, len(deprecatedAPIResources))
	for _, apiResource := range deprecatedAPIResources {
		listKinds[schema.GroupVersionResource{Group: apiResource.group, Version: apiResource.version, Resource: apiResource.resource}] = apiResource.kind + "List"
	}
	return listKinds
}

func countFindingID(findings []types.Finding, id string) int {
	count := 0
	for _, finding := range findings {
		if finding.ID == id {
			count++
		}
	}
	return count
}

func findingIDs(findings []types.Finding) []string {
	ids := make([]string, 0, len(findings))
	for _, finding := range findings {
		ids = append(ids, finding.ID)
	}
	return ids
}
