package cloud

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/kextant/kextant-agent/internal/config"
	"github.com/kextant/kextant-agent/pkg/types"
)

type Client struct {
	endpoint   string
	apiKey     string
	retries    int
	httpClient *http.Client
	logger     *slog.Logger
}

type UploadResponse struct {
	ManifestID  string `json:"manifest_id,omitempty"`
	ReportJobID string `json:"report_job_id,omitempty"`
	ReportID    string `json:"report_id,omitempty"`
	ReportURL   string `json:"report_url,omitempty"`
	Status      string `json:"status,omitempty"`
}

func New(cfg *config.Config, logger *slog.Logger) (*Client, error) {
	if err := cfg.ValidateCloudUpload(); err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	if cfg.KextantCloudTLSSkipVerify {
		tlsConfig.InsecureSkipVerify = true //nolint:gosec // Explicit development-only opt-in.
	}
	if cfg.KextantCloudMTLSCertFile != "" && cfg.KextantCloudMTLSKeyFile != "" {
		certificate, err := tls.LoadX509KeyPair(cfg.KextantCloudMTLSCertFile, cfg.KextantCloudMTLSKeyFile)
		if err != nil {
			return nil, fmt.Errorf("load mTLS certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{certificate}
	}

	baseTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("default HTTP transport is not *http.Transport")
	}
	transport := baseTransport.Clone()
	transport.TLSClientConfig = tlsConfig

	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	return &Client{
		endpoint: strings.TrimRight(cfg.KextantCloudEndpoint, "/"),
		apiKey:   cfg.KextantCloudAPIKey,
		retries:  cfg.KextantCloudUploadRetries,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
		logger: logger,
	}, nil
}

func (c *Client) UploadManifest(ctx context.Context, manifest *types.Manifest) (*UploadResponse, error) {
	attempts := c.retries + 1
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		response, err := c.uploadManifestOnce(ctx, manifest)
		if err == nil {
			return response, nil
		}
		lastErr = err
		if attempt == attempts || !isRetryable(err) {
			break
		}
		backoff := time.Duration(attempt) * 500 * time.Millisecond
		c.logger.Warn("manifest upload failed; retrying", "attempt", attempt, "error", err, "backoff", backoff)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
	}
	return nil, lastErr
}

func (c *Client) uploadManifestOnce(ctx context.Context, manifest *types.Manifest) (*UploadResponse, error) {
	body, err := json.Marshal(manifest)
	if err != nil {
		return nil, fmt.Errorf("marshal manifest: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/v1/manifests", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create manifest upload request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+c.apiKey)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "kextant-agent")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, retryableError{err: err}
	}
	defer response.Body.Close()

	limitedBody, _ := io.ReadAll(io.LimitReader(response.Body, 8192))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		err := fmt.Errorf("manifest upload failed with status %d: %s", response.StatusCode, strings.TrimSpace(string(limitedBody)))
		if response.StatusCode >= 500 || response.StatusCode == http.StatusTooManyRequests {
			return nil, retryableError{err: err}
		}
		return nil, err
	}

	if len(limitedBody) == 0 {
		return &UploadResponse{Status: "accepted"}, nil
	}
	var uploadResponse UploadResponse
	if err := json.Unmarshal(limitedBody, &uploadResponse); err != nil {
		return nil, fmt.Errorf("decode manifest upload response: %w", err)
	}
	return &uploadResponse, nil
}

type retryableError struct {
	err error
}

func (e retryableError) Error() string {
	return e.err.Error()
}

func (e retryableError) Unwrap() error {
	return e.err
}

func isRetryable(err error) bool {
	_, ok := err.(retryableError)
	return ok
}
