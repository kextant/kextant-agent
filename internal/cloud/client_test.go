package cloud

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kextant/kextant-agent/internal/config"
	"github.com/kextant/kextant-agent/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestUploadManifestSendsExpectedRequest(t *testing.T) {
	var authHeader string
	var received types.Manifest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/manifests", r.URL.Path)
		require.Equal(t, http.MethodPost, r.Method)
		authHeader = r.Header.Get("Authorization")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&received))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"manifest_id":"manifest-1","report_job_id":"job-1","status":"accepted"}`))
	}))
	defer server.Close()

	client, err := New(testConfig(server.URL), nil)
	require.NoError(t, err)

	response, err := client.UploadManifest(context.Background(), &types.Manifest{
		SchemaVersion: types.ManifestSchemaVersion,
		ClusterID:     "cluster-1",
		ClusterName:   "prod",
	})
	require.NoError(t, err)
	require.Equal(t, "Bearer secret", authHeader)
	require.Equal(t, "cluster-1", received.ClusterID)
	require.Equal(t, "manifest-1", response.ManifestID)
	require.Equal(t, "job-1", response.ReportJobID)
}

func TestUploadManifestDoesNotRetryClientError(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		http.Error(w, "bad key", http.StatusUnauthorized)
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	cfg.KextantCloudUploadRetries = 3
	client, err := New(cfg, nil)
	require.NoError(t, err)

	_, err = client.UploadManifest(context.Background(), &types.Manifest{})
	require.ErrorContains(t, err, "401")
	require.Equal(t, 1, attempts)
}

func TestUploadManifestRetriesServerError(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		if attempts == 1 {
			http.Error(w, "try again", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	cfg.KextantCloudUploadRetries = 1
	client, err := New(cfg, nil)
	require.NoError(t, err)

	response, err := client.UploadManifest(context.Background(), &types.Manifest{})
	require.NoError(t, err)
	require.Equal(t, "accepted", response.Status)
	require.Equal(t, 2, attempts)
}

func testConfig(endpoint string) *config.Config {
	return &config.Config{
		ClusterID:                 "cluster-1",
		KextantCloudEndpoint:      endpoint,
		KextantCloudAPIKey:        "secret",
		KextantCloudTLSSkipVerify: true,
		KextantCloudUploadRetries: 0,
	}
}
