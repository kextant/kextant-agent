package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("CLUSTER_NAME", "")
	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, "kubernetes-cluster", cfg.ClusterName)
	require.False(t, cfg.KextantCloudEnabled)
	require.Equal(t, RedactionPlain, cfg.RedactionNamespaceNames)
	require.Equal(t, MetadataAllowlist, cfg.RedactionLabels)
	require.Contains(t, cfg.ExcludeNamespaces, "kube-system")
	require.Contains(t, cfg.RedactionLabelAllowlist, "app.kubernetes.io/name")
}

func TestValidateCloudUploadRequiresCredentials(t *testing.T) {
	cfg := &Config{
		KextantCloudEndpoint: "https://api.kextant.com",
	}
	err := cfg.ValidateCloudUpload()
	require.ErrorContains(t, err, "CLUSTER_ID")

	cfg.ClusterID = "cluster-1"
	err = cfg.ValidateCloudUpload()
	require.ErrorContains(t, err, "KEXTANT_CLOUD_API_KEY")

	cfg.KextantCloudAPIKey = "secret"
	require.NoError(t, cfg.ValidateCloudUpload())
}

func TestValidateCloudUploadRequiresHTTPSUnlessDevelopmentOverride(t *testing.T) {
	cfg := &Config{
		ClusterID:                 "cluster-1",
		KextantCloudEndpoint:      "http://localhost:8080",
		KextantCloudAPIKey:        "secret",
		KextantCloudUploadRetries: 1,
	}
	require.ErrorContains(t, cfg.ValidateCloudUpload(), "https")

	cfg.KextantCloudTLSSkipVerify = true
	require.NoError(t, cfg.ValidateCloudUpload())
}

func TestValidateRejectsInvalidRedactionModes(t *testing.T) {
	cfg := &Config{
		RedactionNamespaceNames: "bogus",
		RedactionWorkloadNames:  RedactionPlain,
		RedactionLabels:         MetadataAllowlist,
		RedactionAnnotations:    MetadataAllowlist,
	}
	require.ErrorContains(t, cfg.Validate(), "REDACTION_NAMESPACE_NAMES")
}
