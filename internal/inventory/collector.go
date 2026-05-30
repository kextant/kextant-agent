package inventory

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/kextant/kextant-agent/internal/config"
	"github.com/kextant/kextant-agent/pkg/types"
	corev1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	extensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const helmReleaseSecretType corev1.SecretType = "helm.sh/release.v1"

type Collector struct {
	client           kubernetes.Interface
	extensionsClient extensionsclient.Interface
	cfg              *config.Config
	logger           *slog.Logger
	redactor         redactor
}

func New(cfg *config.Config, logger *slog.Logger) (*Collector, error) {
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{},
		)
		k8sConfig, err = kubeconfig.ClientConfig()
		if err != nil {
			return nil, err
		}
	}

	client, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	extClient, err := extensionsclient.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	return NewWithClients(cfg, logger, client, extClient), nil
}

func NewWithClients(cfg *config.Config, logger *slog.Logger, client kubernetes.Interface, extClient extensionsclient.Interface) *Collector {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &Collector{
		client:           client,
		extensionsClient: extClient,
		cfg:              cfg,
		logger:           logger,
		redactor:         newRedactor(cfg),
	}
}

func (c *Collector) Collect(ctx context.Context, agentVersion, buildCommit string) (*types.Manifest, error) {
	manifest := &types.Manifest{
		SchemaVersion: types.ManifestSchemaVersion,
		AgentVersion:  agentVersion,
		BuildCommit:   buildCommit,
		ClusterID:     c.cfg.ClusterID,
		ClusterName:   c.cfg.ClusterName,
		ScanTimestamp: time.Now().UTC(),
		Redaction: types.ManifestRedaction{
			NamespaceNames:      c.cfg.RedactionNamespaceNames,
			WorkloadNames:       c.cfg.RedactionWorkloadNames,
			Labels:              c.cfg.RedactionLabels,
			Annotations:         c.cfg.RedactionAnnotations,
			LabelAllowlist:      c.cfg.RedactionLabelAllowlist,
			AnnotationAllowlist: c.cfg.RedactionAnnotationAllowlist,
		},
	}

	manifest.Kubernetes = c.collectKubernetesVersion()

	nodes, err := c.collectNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("collect nodes: %w", err)
	}
	manifest.Nodes = nodes

	namespaces, err := c.getNamespacesToScan(ctx)
	if err != nil {
		return nil, fmt.Errorf("list namespaces: %w", err)
	}

	components, err := c.collectComponents(ctx, namespaces)
	if err != nil {
		return nil, fmt.Errorf("collect components: %w", err)
	}
	manifest.Components = components
	manifest.ComponentSummaries = summarizeComponents(components)

	if c.cfg.HelmMetadataEnabled {
		releases, err := c.collectHelmReleases(ctx, namespaces)
		if err != nil {
			return nil, fmt.Errorf("collect helm releases: %w", err)
		}
		manifest.HelmReleases = releases
	}

	crds, err := c.collectCRDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("collect crds: %w", err)
	}
	manifest.CRDs = crds

	return manifest, nil
}

func (c *Collector) collectKubernetesVersion() types.KubernetesInfo {
	version, err := c.client.Discovery().ServerVersion()
	if err != nil {
		c.logger.Warn("failed to collect Kubernetes server version", "error", err)
		return types.KubernetesInfo{}
	}
	return types.KubernetesInfo{
		ServerVersion: version.GitVersion,
		Major:         version.Major,
		Minor:         version.Minor,
		GitVersion:    version.GitVersion,
		Platform:      version.Platform,
	}
}

func (c *Collector) collectNodes(ctx context.Context) ([]types.NodeInfo, error) {
	nodeList, err := c.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	nodes := make([]types.NodeInfo, 0, len(nodeList.Items))
	for _, node := range nodeList.Items {
		nodes = append(nodes, types.NodeInfo{
			Name:           c.redactor.workload(node.Name),
			KubeletVersion: node.Status.NodeInfo.KubeletVersion,
			OSImage:        node.Status.NodeInfo.OSImage,
			Architecture:   node.Status.NodeInfo.Architecture,
			KernelVersion:  node.Status.NodeInfo.KernelVersion,
			Labels:         c.redactor.labels(node.Labels),
			Annotations:    c.redactor.annotations(node.Annotations),
		})
	}
	return nodes, nil
}

func (c *Collector) getNamespacesToScan(ctx context.Context) ([]string, error) {
	if len(c.cfg.ScanNamespaces) > 0 {
		return c.cfg.ScanNamespaces, nil
	}

	nsList, err := c.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	excludeMap := map[string]struct{}{}
	for _, ns := range c.cfg.ExcludeNamespaces {
		excludeMap[ns] = struct{}{}
	}

	namespaces := make([]string, 0, len(nsList.Items))
	for _, ns := range nsList.Items {
		if _, excluded := excludeMap[ns.Name]; !excluded {
			namespaces = append(namespaces, ns.Name)
		}
	}
	return namespaces, nil
}

func (c *Collector) collectComponents(ctx context.Context, namespaces []string) ([]types.ComponentInstance, error) {
	var components []types.ComponentInstance
	for _, namespace := range namespaces {
		deployments, err := c.client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, deployment := range deployments.Items {
			components = append(components, c.componentsFromPodTemplate("Deployment", deployment.Namespace, deployment.Name, deployment.Labels, deployment.Annotations, deployment.OwnerReferences, deployment.Spec.Template)...)
		}

		statefulSets, err := c.client.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, statefulSet := range statefulSets.Items {
			components = append(components, c.componentsFromPodTemplate("StatefulSet", statefulSet.Namespace, statefulSet.Name, statefulSet.Labels, statefulSet.Annotations, statefulSet.OwnerReferences, statefulSet.Spec.Template)...)
		}

		daemonSets, err := c.client.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, daemonSet := range daemonSets.Items {
			components = append(components, c.componentsFromPodTemplate("DaemonSet", daemonSet.Namespace, daemonSet.Name, daemonSet.Labels, daemonSet.Annotations, daemonSet.OwnerReferences, daemonSet.Spec.Template)...)
		}

		replicaSets, err := c.client.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, replicaSet := range replicaSets.Items {
			components = append(components, c.componentsFromPodTemplate("ReplicaSet", replicaSet.Namespace, replicaSet.Name, replicaSet.Labels, replicaSet.Annotations, replicaSet.OwnerReferences, replicaSet.Spec.Template)...)
		}

		jobs, err := c.client.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, job := range jobs.Items {
			components = append(components, c.componentsFromPodTemplate("Job", job.Namespace, job.Name, job.Labels, job.Annotations, job.OwnerReferences, job.Spec.Template)...)
		}

		cronJobs, err := c.client.BatchV1().CronJobs(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, cronJob := range cronJobs.Items {
			components = append(components, c.componentsFromPodTemplate("CronJob", cronJob.Namespace, cronJob.Name, cronJob.Labels, cronJob.Annotations, cronJob.OwnerReferences, cronJob.Spec.JobTemplate.Spec.Template)...)
		}

		pods, err := c.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, pod := range pods.Items {
			components = append(components, c.componentsFromPodSpec("Pod", pod.Namespace, pod.Name, pod.Labels, pod.Annotations, pod.OwnerReferences, pod.Spec)...)
		}
	}
	return components, nil
}

func (c *Collector) componentsFromPodTemplate(kind, namespace, name string, labels, annotations map[string]string, owners []metav1.OwnerReference, template corev1.PodTemplateSpec) []types.ComponentInstance {
	mergedLabels := mergeStringMaps(labels, template.Labels)
	mergedAnnotations := mergeStringMaps(annotations, template.Annotations)
	return c.componentsFromPodSpec(kind, namespace, name, mergedLabels, mergedAnnotations, owners, template.Spec)
}

func (c *Collector) componentsFromPodSpec(kind, namespace, name string, labels, annotations map[string]string, owners []metav1.OwnerReference, spec corev1.PodSpec) []types.ComponentInstance {
	components := make([]types.ComponentInstance, 0, len(spec.InitContainers)+len(spec.Containers))
	for _, container := range spec.InitContainers {
		components = append(components, c.componentFromContainer(kind, namespace, name, "init", labels, annotations, owners, container))
	}
	for _, container := range spec.Containers {
		components = append(components, c.componentFromContainer(kind, namespace, name, "container", labels, annotations, owners, container))
	}
	return components
}

func (c *Collector) componentFromContainer(kind, namespace, name, containerType string, labels, annotations map[string]string, owners []metav1.OwnerReference, container corev1.Container) types.ComponentInstance {
	parsed := parseImageReference(container.Image)
	return types.ComponentInstance{
		Image:           container.Image,
		ImageRegistry:   parsed.Registry,
		ImageRepository: parsed.Repository,
		ImageTag:        parsed.Tag,
		ImageDigest:     parsed.Digest,
		Namespace:       c.redactor.namespace(namespace),
		Kind:            kind,
		Name:            c.redactor.workload(name),
		ContainerName:   c.redactor.workload(container.Name),
		ContainerType:   containerType,
		Labels:          c.redactor.labels(labels),
		Annotations:     c.redactor.annotations(annotations),
		OwnerReferences: convertOwnerReferences(owners, c.redactor),
	}
}

func (c *Collector) collectHelmReleases(ctx context.Context, namespaces []string) ([]types.HelmRelease, error) {
	var releases []types.HelmRelease
	for _, namespace := range namespaces {
		secrets, err := c.client.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, secret := range secrets.Items {
			if !isHelmReleaseSecret(secret) {
				continue
			}
			release := decodeHelmRelease(secret)
			release.Namespace = c.redactor.namespace(secret.Namespace)
			release.Name = c.redactor.workload(release.Name)
			releases = append(releases, release)
		}
	}
	return releases, nil
}

func (c *Collector) collectCRDs(ctx context.Context) ([]types.CRDInfo, error) {
	if c.extensionsClient == nil {
		return nil, nil
	}
	crdList, err := c.extensionsClient.ApiextensionsV1().CustomResourceDefinitions().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	crds := make([]types.CRDInfo, 0, len(crdList.Items))
	for _, crd := range crdList.Items {
		crds = append(crds, c.crdInfo(crd))
	}
	return crds, nil
}

func (c *Collector) crdInfo(crd extensionsv1.CustomResourceDefinition) types.CRDInfo {
	versions := make([]string, 0, len(crd.Spec.Versions))
	for _, version := range crd.Spec.Versions {
		versions = append(versions, version.Name)
	}
	return types.CRDInfo{
		Name:        crd.Name,
		Group:       crd.Spec.Group,
		Versions:    versions,
		Scope:       string(crd.Spec.Scope),
		Labels:      c.redactor.labels(crd.Labels),
		Annotations: c.redactor.annotations(crd.Annotations),
	}
}

type helmReleaseEnvelope struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Info      struct {
		Status       string    `json:"status"`
		LastDeployed time.Time `json:"last_deployed"`
	} `json:"info"`
	Chart struct {
		Metadata struct {
			Name       string `json:"name"`
			Version    string `json:"version"`
			AppVersion string `json:"appVersion"`
		} `json:"metadata"`
	} `json:"chart"`
}

func isHelmReleaseSecret(secret corev1.Secret) bool {
	return secret.Type == helmReleaseSecretType || secret.Labels["owner"] == "helm"
}

func decodeHelmRelease(secret corev1.Secret) types.HelmRelease {
	fallback := types.HelmRelease{
		Name:      firstNonEmpty(secret.Labels["name"], strings.TrimPrefix(secret.Name, "sh.helm.release.v1.")),
		Namespace: secret.Namespace,
		Status:    secret.Labels["status"],
	}
	if version := secret.Labels["version"]; version != "" {
		fallback.ChartVersion = version
	}

	payload := secret.Data["release"]
	if len(payload) == 0 {
		return fallback
	}
	decoded, err := decodeHelmPayload(payload)
	if err != nil {
		return fallback
	}

	var envelope helmReleaseEnvelope
	if err := json.Unmarshal(decoded, &envelope); err != nil {
		return fallback
	}

	return types.HelmRelease{
		Name:         firstNonEmpty(envelope.Name, fallback.Name),
		Namespace:    firstNonEmpty(envelope.Namespace, fallback.Namespace),
		Chart:        envelope.Chart.Metadata.Name,
		ChartVersion: envelope.Chart.Metadata.Version,
		AppVersion:   envelope.Chart.Metadata.AppVersion,
		Status:       firstNonEmpty(envelope.Info.Status, fallback.Status),
		UpdatedAt:    envelope.Info.LastDeployed,
	}
}

func decodeHelmPayload(payload []byte) ([]byte, error) {
	decoded := payload
	if base64Decoded, err := base64.StdEncoding.DecodeString(string(payload)); err == nil {
		decoded = base64Decoded
	}

	reader, err := gzip.NewReader(bytes.NewReader(decoded))
	if err != nil {
		return decoded, nil
	}
	defer reader.Close()

	uncompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return uncompressed, nil
}

func mergeStringMaps(base, override map[string]string) map[string]string {
	if len(base) == 0 && len(override) == 0 {
		return nil
	}
	result := map[string]string{}
	for key, value := range base {
		result[key] = value
	}
	for key, value := range override {
		result[key] = value
	}
	return result
}

func convertOwnerReferences(owners []metav1.OwnerReference, r redactor) []types.OwnerReference {
	if len(owners) == 0 {
		return nil
	}
	result := make([]types.OwnerReference, 0, len(owners))
	for _, owner := range owners {
		result = append(result, types.OwnerReference{
			APIVersion: owner.APIVersion,
			Kind:       owner.Kind,
			Name:       r.workload(owner.Name),
		})
	}
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
