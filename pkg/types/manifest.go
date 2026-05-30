package types

import "time"

const ManifestSchemaVersion = "1.0.0"

type Manifest struct {
	SchemaVersion      string              `json:"schema_version"`
	AgentVersion       string              `json:"agent_version"`
	BuildCommit        string              `json:"build_commit,omitempty"`
	ClusterID          string              `json:"cluster_id,omitempty"`
	ClusterName        string              `json:"cluster_name"`
	ScanTimestamp      time.Time           `json:"scan_timestamp"`
	Redaction          ManifestRedaction   `json:"redaction"`
	Kubernetes         KubernetesInfo      `json:"kubernetes"`
	Nodes              []NodeInfo          `json:"nodes"`
	Components         []ComponentInstance `json:"components"`
	ComponentSummaries []ComponentSummary  `json:"component_summaries,omitempty"`
	HelmReleases       []HelmRelease       `json:"helm_releases"`
	CRDs               []CRDInfo           `json:"crds"`
}

type ManifestRedaction struct {
	NamespaceNames      string   `json:"namespace_names"`
	WorkloadNames       string   `json:"workload_names"`
	Labels              string   `json:"labels"`
	Annotations         string   `json:"annotations"`
	LabelAllowlist      []string `json:"label_allowlist,omitempty"`
	AnnotationAllowlist []string `json:"annotation_allowlist,omitempty"`
}

type KubernetesInfo struct {
	ServerVersion string `json:"server_version,omitempty"`
	Major         string `json:"major,omitempty"`
	Minor         string `json:"minor,omitempty"`
	GitVersion    string `json:"git_version,omitempty"`
	Platform      string `json:"platform,omitempty"`
}

type NodeInfo struct {
	Name           string            `json:"name"`
	KubeletVersion string            `json:"kubelet_version"`
	OSImage        string            `json:"os_image,omitempty"`
	Architecture   string            `json:"arch,omitempty"`
	KernelVersion  string            `json:"kernel_version,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
	Annotations    map[string]string `json:"annotations,omitempty"`
}

type ComponentInstance struct {
	Image           string            `json:"image"`
	ImageRegistry   string            `json:"image_registry,omitempty"`
	ImageRepository string            `json:"image_repository,omitempty"`
	ImageTag        string            `json:"image_tag,omitempty"`
	ImageDigest     string            `json:"image_digest,omitempty"`
	Namespace       string            `json:"namespace"`
	Kind            string            `json:"kind"`
	Name            string            `json:"name"`
	ContainerName   string            `json:"container_name"`
	ContainerType   string            `json:"container_type"`
	Labels          map[string]string `json:"labels,omitempty"`
	Annotations     map[string]string `json:"annotations,omitempty"`
	OwnerReferences []OwnerReference  `json:"owner_references,omitempty"`
}

type ComponentSummary struct {
	Image           string              `json:"image"`
	ImageRegistry   string              `json:"image_registry,omitempty"`
	ImageRepository string              `json:"image_repository,omitempty"`
	ImageTag        string              `json:"image_tag,omitempty"`
	ImageDigest     string              `json:"image_digest,omitempty"`
	Locations       []ComponentLocation `json:"locations"`
}

type ComponentLocation struct {
	Namespace     string `json:"namespace"`
	Kind          string `json:"kind"`
	Name          string `json:"name"`
	ContainerName string `json:"container_name"`
	ContainerType string `json:"container_type"`
}

type OwnerReference struct {
	APIVersion string `json:"api_version,omitempty"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
}

type HelmRelease struct {
	Name         string    `json:"name"`
	Namespace    string    `json:"namespace"`
	Chart        string    `json:"chart,omitempty"`
	ChartVersion string    `json:"chart_version,omitempty"`
	AppVersion   string    `json:"app_version,omitempty"`
	Status       string    `json:"status,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}

type CRDInfo struct {
	Name        string            `json:"name"`
	Group       string            `json:"group"`
	Versions    []string          `json:"versions"`
	Scope       string            `json:"scope"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}
