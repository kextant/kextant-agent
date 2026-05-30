# Inventory Manifest Schema

`docs/PRODUCT_REQUIREMENTS.md` is authoritative. This document describes the v0.1.0 manifest emitted by `kextant-agent inventory`.

## Generate locally

```bash
./kextant-agent inventory > manifest.json
```

## Upload to Kextant Cloud

```bash
CLUSTER_ID=<cluster-id> \
KEXTANT_CLOUD_API_KEY=<cluster-api-key> \
./kextant-agent inventory --upload
```

## Top-level schema

```json
{
  "schema_version": "1.0.0",
  "agent_version": "0.1.0",
  "build_commit": "abc123",
  "cluster_id": "cluster-uuid",
  "cluster_name": "prod-us-east-1",
  "scan_timestamp": "2026-05-29T00:00:00Z",
  "redaction": {},
  "kubernetes": {},
  "nodes": [],
  "components": [],
  "component_summaries": [],
  "helm_releases": [],
  "crds": []
}
```

## Redaction

The `redaction` object records what privacy controls were applied:

- `namespace_names`: `plain`, `hashed`, or `omitted`
- `workload_names`: `plain`, `hashed`, or `omitted`
- `labels`: `all`, `allowlist`, or `none`
- `annotations`: `all`, `allowlist`, or `none`
- `label_allowlist`: configured safe label keys
- `annotation_allowlist`: configured safe annotation keys

## Kubernetes

The `kubernetes` object includes:

- `server_version`
- `major`
- `minor`
- `git_version`
- `platform`

## Nodes

Each node includes:

- `name`
- `kubelet_version`
- `os_image`
- `arch`
- `kernel_version`
- allowlisted `labels`
- allowlisted `annotations`

Node names are redacted using `REDACTION_WORKLOAD_NAMES` because they can reveal infrastructure details.

## Components

Each component occurrence represents one container image in a workload or pod:

- `image`
- `image_registry`
- `image_repository`
- `image_tag`
- `image_digest`
- `namespace`
- `kind`
- `name`
- `container_name`
- `container_type`: `container` or `init`
- allowlisted `labels`
- allowlisted `annotations`
- `owner_references`

Workload kinds collected in v0.1.0:

- Deployment
- StatefulSet
- DaemonSet
- ReplicaSet
- Job
- CronJob
- Pod

## Component summaries

`component_summaries` deduplicates repeated images while preserving all affected locations. Each summary includes the parsed image fields and a `locations` array with namespace, kind, workload name, container name, and container type.

This lets Kextant Cloud produce one finding per image/component while still showing every affected workload.

## Helm releases

Helm releases are collected only when `HELM_METADATA_ENABLED=true` and RBAC allows Secret reads.

Each release includes:

- `name`
- `namespace`
- `chart`
- `chart_version`
- `app_version`
- `status`
- `updated_at`

## CRDs

Each CRD includes:

- `name`
- `group`
- `versions`
- `scope`
- allowlisted `labels`
- allowlisted `annotations`

## Data explicitly not included

The manifest must not contain:

- Kubernetes Secret values
- ConfigMap values
- Environment variable values
- Volume contents
- Pod logs
- Application payloads
- Service account tokens
