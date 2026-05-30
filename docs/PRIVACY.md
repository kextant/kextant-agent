# Privacy and Security Model

`docs/PRODUCT_REQUIREMENTS.md` is authoritative. This document describes how the v0.1.0 prototype limits data collection and cluster privileges.

## Design principles

- The agent is deterministic and read-only.
- Cloud upload is disabled by default.
- Users can export and inspect the exact JSON manifest before uploading.
- Unknown or sensitive data should be omitted rather than guessed or collected.

## RBAC

The agent uses `get` and `list` permissions only.

Required resources:

- Pods
- Nodes
- Namespaces
- Workload controllers
- Jobs/CronJobs
- ResourceQuotas
- LimitRanges
- NetworkPolicies
- PodDisruptionBudgets
- HorizontalPodAutoscalers
- CustomResourceDefinitions

Optional resource:

- Secrets, only for Helm release metadata when enabled in the Helm chart with `rbac.helmSecrets=true`.

## Helm Secrets

Helm stores release metadata in Kubernetes Secrets. When enabled, Kextant decodes only the Helm release payload to extract:

- release name
- namespace
- chart name
- chart version
- app version
- status
- last deployed time

Kextant does not serialize arbitrary Secret keys or values into the manifest.

## Redaction controls

| Environment variable | Values | Purpose |
| --- | --- | --- |
| `REDACTION_NAMESPACE_NAMES` | `plain`, `hashed`, `omitted` | Controls namespace names |
| `REDACTION_WORKLOAD_NAMES` | `plain`, `hashed`, `omitted` | Controls workload, node, container, and owner names |
| `REDACTION_LABELS` | `all`, `allowlist`, `none` | Controls labels |
| `REDACTION_ANNOTATIONS` | `all`, `allowlist`, `none` | Controls annotations |
| `REDACTION_LABEL_ALLOWLIST` | CSV | Safe label keys to include |
| `REDACTION_ANNOTATION_ALLOWLIST` | CSV | Safe annotation keys to include |

Hashed names use a stable SHA-256 prefix so reports can correlate repeated findings without exposing original names.

## Default metadata allowlists

Default labels:

- `app`
- `k8s-app`
- `name`
- `version`
- `app.kubernetes.io/name`
- `app.kubernetes.io/instance`
- `app.kubernetes.io/version`
- `app.kubernetes.io/component`
- `app.kubernetes.io/part-of`
- `app.kubernetes.io/managed-by`
- `helm.sh/chart`

Default annotations:

- `meta.helm.sh/release-name`
- `meta.helm.sh/release-namespace`

## Cloud authentication

Cloud upload requires:

- `CLUSTER_ID`
- `KEXTANT_CLOUD_API_KEY`
- HTTPS endpoint unless an explicit local development override is enabled

mTLS file configuration is available for future enterprise deployments:

- `KEXTANT_CLOUD_MTLS_CERT_FILE`
- `KEXTANT_CLOUD_MTLS_KEY_FILE`

## Container security

The supplied manifests and Helm chart run the agent as:

- non-root user
- no privilege escalation
- read-only root filesystem
- all Linux capabilities dropped
- modest CPU and memory requests/limits
