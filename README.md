# Kextant Agent

Kextant Agent is a read-only Kubernetes health scanner and inventory collector for Kextant Version Intelligence.

The authoritative product requirements are in [`docs/PRODUCT_REQUIREMENTS.md`](docs/PRODUCT_REQUIREMENTS.md). That document takes priority over all other repository documentation.

## What this prototype provides

- Local Kubernetes health checks and reports
- Versioned JSON inventory manifest export
- Optional Kextant Cloud manifest upload
- Privacy/redaction controls
- Slack, email, and log delivery for local health reports
- Helm chart and raw Kubernetes manifests
- Containerized build and GitHub Actions CI/CD

## Quick start

```bash
# Build and test locally
make test-all

# Build the binary
make build

# Print version metadata
./kextant-agent version --json

# Validate config
./kextant-agent validate

# Run local health scan
./kextant-agent scan

# Export inventory manifest
./kextant-agent inventory
```

## Kubernetes install

```bash
helm install kextant charts/kextant-agent \
  --namespace kextant \
  --create-namespace
```

See:

- [`docs/INSTALLATION.md`](docs/INSTALLATION.md)
- [`docs/MANIFEST_SCHEMA.md`](docs/MANIFEST_SCHEMA.md)
- [`docs/PRIVACY.md`](docs/PRIVACY.md)
- [`docs/DEVELOPMENT.md`](docs/DEVELOPMENT.md)
- [`docs/CI_CD.md`](docs/CI_CD.md)
