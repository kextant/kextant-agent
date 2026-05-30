# Development Guide

`docs/PRODUCT_REQUIREMENTS.md` is authoritative. This document describes the v0.1.0 development workflow.

## Language and runtime

- Go 1.24.2
- Kubernetes client-go v0.29.x
- Container runtime: Docker-compatible OCI image

Go is used because it provides strong Kubernetes client support, low memory usage, fast startup, and simple static binaries.

## Local quality loop

```bash
make fmt
make vet
make lint
make test
make test-coverage
make docker-build
make docker-run-smoke
```

Run all local gates:

```bash
make test-all
```

## Test strategy

The prototype uses TDD-style package tests for new behavior:

- `internal/inventory`: manifest collection, image parsing, Helm metadata, redaction
- `internal/cloud`: manifest upload, auth headers, retry behavior
- `internal/config`: validation and defaults
- `internal/scanner`: health-check behavior against fake Kubernetes clients
- `internal/report`: score, filtering, Slack payloads, HTML/plain rendering
- `internal/delivery`: log, Slack, email message/auth helpers
- `internal/health`: health and readiness handlers

Functional container smoke testing verifies that the built image starts and supports:

- `version --json`
- `validate`

## Build

```bash
make build VERSION=0.1.0
```

The build injects:

- `main.version`
- `main.commit`
- `main.buildDate`

## Container build

```bash
make docker-build VERSION=0.1.0
make docker-run-smoke VERSION=0.1.0
```

## Inventory development

Generate a manifest from your current kubeconfig context:

```bash
make build
./kextant-agent inventory > manifest.json
```

With redaction:

```bash
REDACTION_NAMESPACE_NAMES=hashed \
REDACTION_WORKLOAD_NAMES=hashed \
REDACTION_LABELS=allowlist \
REDACTION_ANNOTATIONS=allowlist \
./kextant-agent inventory > manifest.json
```

## Cloud upload development

Use a local test server or staging endpoint:

```bash
CLUSTER_ID=dev-cluster \
KEXTANT_CLOUD_API_KEY=dev-key \
KEXTANT_CLOUD_ENDPOINT=http://localhost:8080 \
KEXTANT_CLOUD_TLS_SKIP_VERIFY=true \
./kextant-agent inventory --upload
```

`KEXTANT_CLOUD_TLS_SKIP_VERIFY=true` is only for local development.

## Release checklist

1. Update version in `charts/kextant-agent/Chart.yaml` and `values.yaml`.
2. Run `make test-all`.
3. Tag release as `v0.1.0` or later.
4. Push tag to GitHub.
5. GitHub Actions builds and publishes GHCR image.
