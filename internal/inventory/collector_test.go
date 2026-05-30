package inventory

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/kextant/kextant-agent/internal/config"
	"github.com/kextant/kextant-agent/pkg/types"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	extensionsfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestCollectBuildsManifestFromClusterInventory(t *testing.T) {
	cfg := inventoryTestConfig()
	cfg.HelmMetadataEnabled = true

	kubeClient := k8sfake.NewSimpleClientset(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kube-system"}},
		&corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "ip-10-0-1-10", Labels: map[string]string{"app.kubernetes.io/name": "node", "private": "omit"}},
			Status: corev1.NodeStatus{NodeInfo: corev1.NodeSystemInfo{
				KubeletVersion: "v1.33.2",
				OSImage:        "Amazon Linux 2023",
				Architecture:   "amd64",
				KernelVersion:  "6.1",
			}},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "loki",
				Namespace: "default",
				Labels: map[string]string{
					"app.kubernetes.io/name": "loki",
					"private":                "should-not-appear",
				},
				Annotations: map[string]string{
					"meta.helm.sh/release-name": "loki",
					"private":                   "should-not-appear",
				},
			},
			Spec: appsv1.DeploymentSpec{Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{
				Name:  "loki",
				Image: "harbor.internal.io/dockerhub/grafana/loki:3.1.1@sha256:abc123",
			}}}}},
		},
		helmSecret(t, "default"),
	)
	discovery, ok := kubeClient.Discovery().(*fake.FakeDiscovery)
	require.True(t, ok)
	discovery.FakedServerVersion = &version.Info{Major: "1", Minor: "33", GitVersion: "v1.33.2", Platform: "linux/amd64"}

	extClient := extensionsfake.NewSimpleClientset(&extensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: "certificates.cert-manager.io", Labels: map[string]string{"app.kubernetes.io/name": "cert-manager"}},
		Spec: extensionsv1.CustomResourceDefinitionSpec{
			Group:    "cert-manager.io",
			Scope:    extensionsv1.NamespaceScoped,
			Versions: []extensionsv1.CustomResourceDefinitionVersion{{Name: "v1"}},
		},
	})

	collector := NewWithClients(cfg, nil, kubeClient, extClient)
	manifest, err := collector.Collect(context.Background(), "0.1.0", "abc")
	require.NoError(t, err)

	require.Equal(t, types.ManifestSchemaVersion, manifest.SchemaVersion)
	require.Equal(t, "0.1.0", manifest.AgentVersion)
	require.Equal(t, "abc", manifest.BuildCommit)
	require.Equal(t, "cluster-1", manifest.ClusterID)
	require.Equal(t, "prod-us-east-1", manifest.ClusterName)
	require.Equal(t, "v1.33.2", manifest.Kubernetes.ServerVersion)

	require.Len(t, manifest.Nodes, 1)
	require.Equal(t, "v1.33.2", manifest.Nodes[0].KubeletVersion)
	require.Equal(t, map[string]string{"app.kubernetes.io/name": "node"}, manifest.Nodes[0].Labels)

	require.Len(t, manifest.Components, 1)
	require.Len(t, manifest.ComponentSummaries, 1)
	require.Len(t, manifest.ComponentSummaries[0].Locations, 1)
	component := manifest.Components[0]
	require.Equal(t, "Deployment", component.Kind)
	require.Equal(t, "default", component.Namespace)
	require.Equal(t, "loki", component.Name)
	require.Equal(t, "harbor.internal.io", component.ImageRegistry)
	require.Equal(t, "dockerhub/grafana/loki", component.ImageRepository)
	require.Equal(t, "3.1.1", component.ImageTag)
	require.Equal(t, "sha256:abc123", component.ImageDigest)
	require.Equal(t, map[string]string{"app.kubernetes.io/name": "loki"}, component.Labels)
	require.Equal(t, map[string]string{"meta.helm.sh/release-name": "loki"}, component.Annotations)

	require.Len(t, manifest.HelmReleases, 1)
	require.Equal(t, "loki", manifest.HelmReleases[0].Name)
	require.Equal(t, "loki", manifest.HelmReleases[0].Chart)
	require.Equal(t, "6.7.3", manifest.HelmReleases[0].ChartVersion)
	require.Equal(t, "3.1.1", manifest.HelmReleases[0].AppVersion)

	require.Len(t, manifest.CRDs, 1)
	require.Equal(t, "cert-manager.io", manifest.CRDs[0].Group)
	require.Equal(t, []string{"v1"}, manifest.CRDs[0].Versions)

	encoded, err := json.Marshal(manifest)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "should-not-appear")
	require.NotContains(t, string(encoded), "supersecret")
}

func TestCollectHonorsRedaction(t *testing.T) {
	cfg := inventoryTestConfig()
	cfg.RedactionNamespaceNames = config.RedactionHashed
	cfg.RedactionWorkloadNames = config.RedactionHashed
	cfg.RedactionLabels = config.MetadataNone
	cfg.RedactionAnnotations = config.MetadataNone

	kubeClient := k8sfake.NewSimpleClientset(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "payments"}},
		&corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node-a"}, Status: corev1.NodeStatus{NodeInfo: corev1.NodeSystemInfo{KubeletVersion: "v1.33.2"}}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "payments", Labels: map[string]string{"app": "api"}}, Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "api", Image: "example.com/api:v1"}}}},
	)
	extClient := extensionsfake.NewSimpleClientset()

	manifest, err := NewWithClients(cfg, nil, kubeClient, extClient).Collect(context.Background(), "0.1.0", "abc")
	require.NoError(t, err)
	require.Len(t, manifest.Components, 1)
	require.NotEqual(t, "payments", manifest.Components[0].Namespace)
	require.Contains(t, manifest.Components[0].Namespace, "sha256:")
	require.NotEqual(t, "api", manifest.Components[0].Name)
	require.Empty(t, manifest.Components[0].Labels)
	require.Empty(t, manifest.Components[0].Annotations)
	require.NotEqual(t, "node-a", manifest.Nodes[0].Name)
}

func TestHelmMetadataCanBeDisabled(t *testing.T) {
	cfg := inventoryTestConfig()
	cfg.HelmMetadataEnabled = false
	kubeClient := k8sfake.NewSimpleClientset(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
		helmSecret(t, "default"),
	)
	manifest, err := NewWithClients(cfg, nil, kubeClient, extensionsfake.NewSimpleClientset()).Collect(context.Background(), "0.1.0", "abc")
	require.NoError(t, err)
	require.Empty(t, manifest.HelmReleases)
}

func inventoryTestConfig() *config.Config {
	return &config.Config{
		ClusterName:                  "prod-us-east-1",
		ClusterID:                    "cluster-1",
		ExcludeNamespaces:            []string{"kube-system", "kube-public", "kube-node-lease", "kextant"},
		RedactionNamespaceNames:      config.RedactionPlain,
		RedactionWorkloadNames:       config.RedactionPlain,
		RedactionLabels:              config.MetadataAllowlist,
		RedactionAnnotations:         config.MetadataAllowlist,
		RedactionLabelAllowlist:      config.DefaultLabelAllowlist,
		RedactionAnnotationAllowlist: config.DefaultAnnotationAllowlist,
	}
}

func helmSecret(t *testing.T, namespace string) *corev1.Secret {
	t.Helper()
	updatedAt := time.Date(2026, 5, 29, 12, 0, 0, 0, time.UTC)
	payload := map[string]any{
		"name":      "loki",
		"namespace": namespace,
		"info": map[string]any{
			"status":        "deployed",
			"last_deployed": updatedAt.Format(time.RFC3339),
		},
		"chart": map[string]any{
			"metadata": map[string]any{
				"name":       "loki",
				"version":    "6.7.3",
				"appVersion": "3.1.1",
			},
		},
		"ignored_secret_value": "supersecret",
	}
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	_, err = writer.Write(payloadBytes)
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	encoded := base64.StdEncoding.EncodeToString(compressed.Bytes())

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sh.helm.release.v1.loki.v1",
			Namespace: namespace,
			Labels: map[string]string{
				"owner":  "helm",
				"name":   "loki",
				"status": "deployed",
			},
		},
		Type: helmReleaseSecretType,
		Data: map[string][]byte{"release": []byte(encoded)},
	}
}
