package health

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}

type Server struct {
	checker HealthChecker
	logger  *slog.Logger
}

func NewServer(checker HealthChecker, logger *slog.Logger) *Server {
	return &Server{
		checker: checker,
		logger:  logger,
	}
}

func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.healthzHandler)
	mux.HandleFunc("/readyz", s.readyzHandler)

	s.logger.Info("starting health server", "addr", addr)
	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	return server.ListenAndServe()
}

func (s *Server) healthzHandler(w http.ResponseWriter, r *http.Request) {
	// Liveness probe - just check if the process is running.
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		s.logger.Error("failed to write health response", "error", err)
	}
}

func (s *Server) readyzHandler(w http.ResponseWriter, r *http.Request) {
	// Readiness probe - check if we can connect to K8s API
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := s.checker.HealthCheck(ctx); err != nil {
		s.logger.Error("readiness check failed", "error", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		if _, writeErr := w.Write([]byte("kubernetes api unavailable")); writeErr != nil {
			s.logger.Error("failed to write readiness failure response", "error", writeErr)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		s.logger.Error("failed to write readiness response", "error", err)
	}
}
