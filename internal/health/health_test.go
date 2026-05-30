package health

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeChecker struct {
	err error
}

func (f fakeChecker) HealthCheck(context.Context) error {
	return f.err
}

func TestHealthzHandler(t *testing.T) {
	server := NewServer(fakeChecker{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	server.healthzHandler(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "ok", response.Body.String())
}

func TestReadyzHandlerSuccess(t *testing.T) {
	server := NewServer(fakeChecker{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	response := httptest.NewRecorder()

	server.readyzHandler(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "ok", response.Body.String())
}

func TestReadyzHandlerFailure(t *testing.T) {
	server := NewServer(fakeChecker{err: errors.New("api unavailable")}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	response := httptest.NewRecorder()

	server.readyzHandler(response, request)

	require.Equal(t, http.StatusServiceUnavailable, response.Code)
	require.Equal(t, "kubernetes api unavailable", response.Body.String())
}
