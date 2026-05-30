package inventory

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseImageReference(t *testing.T) {
	tests := []struct {
		name       string
		image      string
		registry   string
		repository string
		tag        string
		digest     string
	}{
		{
			name:       "docker hub implicit registry with tag",
			image:      "grafana/loki:3.1.1",
			repository: "grafana/loki",
			tag:        "3.1.1",
		},
		{
			name:       "private registry mirror with tag and digest",
			image:      "harbor.internal.io/dockerhub/grafana/loki:3.1.1@sha256:abc123",
			registry:   "harbor.internal.io",
			repository: "dockerhub/grafana/loki",
			tag:        "3.1.1",
			digest:     "sha256:abc123",
		},
		{
			name:       "localhost port",
			image:      "localhost:5000/team/app:v1",
			registry:   "localhost:5000",
			repository: "team/app",
			tag:        "v1",
		},
		{
			name:       "digest only",
			image:      "quay.io/org/app@sha256:def456",
			registry:   "quay.io",
			repository: "org/app",
			digest:     "sha256:def456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := parseImageReference(tt.image)
			require.Equal(t, tt.registry, parsed.Registry)
			require.Equal(t, tt.repository, parsed.Repository)
			require.Equal(t, tt.tag, parsed.Tag)
			require.Equal(t, tt.digest, parsed.Digest)
		})
	}
}
