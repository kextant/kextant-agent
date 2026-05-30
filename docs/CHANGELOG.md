# Changelog

`docs/PRODUCT_REQUIREMENTS.md` is authoritative.

## v0.1.0 - Initial prototype

### Added

- Read-only Kubernetes health scanner preserved from MVP.
- Versioned inventory manifest schema `1.0.0`.
- Inventory collection for Kubernetes version, nodes, workloads, pods, Helm release metadata, and CRDs.
- Deduplicated component summaries while preserving affected locations.
- Local `inventory` command with optional `--output` and `--upload`.
- `scan --manifest`, `scan --manifest-output`, and `scan --upload` flows.
- Kextant Cloud upload client with API key auth, TLS, retry behavior, and optional mTLS certificate files.
- Redaction controls for namespace names, workload names, labels, and annotations.
- Helm chart with Deployment and CronJob modes.
- Optional Helm Secret RBAC in Helm chart.
- GitHub Actions CI/CD with linting, unit tests, coverage threshold, container smoke tests, and GHCR image publishing.
- Containerized application build using Go and Alpine.

### Changed

- Documentation is consolidated under `docs/` with `docs/PRODUCT_REQUIREMENTS.md` as the priority document.
- Raw deployment manifests use immutable image tag `0.1.0` by default.
- Health server now uses explicit HTTP timeouts.

### Removed

- Redundant legacy root-level planning/specification docs.
- GitLab CI configuration in favor of GitHub Actions.
