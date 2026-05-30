# CI/CD

`docs/PRODUCT_REQUIREMENTS.md` is authoritative. This document describes the GitHub Actions pipeline for the v0.1.0 prototype.

## Workflow

The workflow lives at `.github/workflows/ci.yml`.

It runs on:

- pushes to `main`
- pushes to `cicd`
- version tags matching `v*.*.*`
- pull requests targeting `main`

## Jobs

### `lint-test`

Runs:

1. `gofmt` verification
2. `go vet ./...`
3. `golangci-lint`
4. race-enabled unit tests with coverage
5. enforces a minimum total coverage threshold of 60%
6. uploads `coverage.out` and `coverage.txt` as artifacts

### `container-functional`

Runs after lint/unit tests pass.

1. Builds the container image.
2. Runs functional smoke tests in the container:
   - `kextant-agent version --json`
   - `kextant-agent validate`
3. On push events, logs in to GitHub Container Registry.
4. Builds and publishes private multi-arch images (`linux/amd64`, `linux/arm64`) to `ghcr.io/kextant/kextant-agent`.

## Image publishing

Images are published to GitHub Container Registry using the repository's `GITHUB_TOKEN`.

Tags are generated from:

- branch name
- git tag
- commit SHA

The GitHub package should remain private until the product is intentionally made public.

## Required repository settings

- Actions enabled.
- Workflow permissions allow package writes.
- GHCR package visibility set to private unless explicitly changed.
- Branch protection should require the CI workflow before merging to `main`.

## Local parity

Run the local equivalent before pushing:

```bash
make test-all
```
