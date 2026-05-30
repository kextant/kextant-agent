package delivery

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kextant/kextant-agent/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestSlackDeliverySend(t *testing.T) {
	var payload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	delivery := NewSlackDelivery(server.URL, slog.New(slog.NewTextHandler(io.Discard, nil)))
	err := delivery.Send(&types.Report{
		ClusterName:   "prod",
		GeneratedAt:   time.Date(2026, 5, 29, 0, 0, 0, 0, time.UTC),
		Score:         90,
		Grade:         "Excellent",
		CriticalCount: 1,
		Findings: []types.Finding{{
			ID:       "SEC002",
			Title:    "Privileged container",
			Severity: types.SeverityCritical,
			Affected: []types.AffectedResource{{Namespace: "default", Name: "api/app"}},
		}},
	})
	require.NoError(t, err)
	require.Contains(t, payload, "blocks")
}
