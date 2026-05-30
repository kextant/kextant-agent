# Kextant Agent Product Requirements

**Repository:** `github.com/kextant/kextant-agent`  
**Primary product role:** Trusted in-cluster scanner and inventory collector  
**Document status:** Authoritative product requirements  
**Last updated:** 2026-05-29

> **Priority notice:** This document takes priority over all other product, functional, technical, and roadmap documentation in this repository. If another document conflicts with this file, implement this file first and update the conflicting document later.

---

## 1. Product purpose

`kextant-agent` is the installable component that runs inside a customer's Kubernetes cluster. Its purpose is to earn trust, collect deterministic facts, and provide immediate local value without requiring Kextant Cloud.

The agent serves the whole Kextant product by being the data source for Kubernetes Version Intelligence. Kextant Cloud, the knowledge base, and the web app are only useful if the agent can safely and accurately answer:

1. What Kubernetes version and node versions are running?
2. What workload images are deployed?
3. What Helm charts and app versions are installed?
4. What CRDs exist, and which operators/components do they imply?
5. What common Kubernetes health and configuration issues exist locally?
6. Can the user inspect exactly what will be uploaded before enabling cloud analysis?

The agent must remain small, auditable, read-only, deterministic, and easy to remove.

---

## 2. Product strategy

### 2.1 Strategic role

The agent is not the main monetized asset. It is the adoption wedge and trust layer.

Paid value is created by Kextant Cloud and the Version Intelligence Knowledge Base, but the agent must make cloud value possible by collecting a complete enough inventory manifest. The agent should eventually be open source or source-available to reduce security objections.

### 2.2 Primary value propositions

For users who install only the agent:

- Scheduled Kubernetes health reports delivered through logs, Slack, or email.
- Local-only operation with no SaaS dependency.
- Read-only scanning of common operational risks.
- JSON inventory export for manual review.

For users connecting to Kextant Cloud:

- Accurate live-cluster inventory for version analysis.
- Secure upload of sanitized manifest data.
- Low-friction path from install to first version intelligence report.

### 2.3 Non-goals

The agent must not become:

- A cloud backend.
- A dashboard.
- A policy engine that mutates cluster resources.
- An LLM client.
- A vulnerability scanner that performs deep image layer analysis.
- A heavyweight daemon requiring persistent volumes or cluster-wide write access.

---

## 3. Target users

| User | What they need from this repo |
| --- | --- |
| Platform engineer | Quick install, predictable read-only RBAC, accurate inventory, local health report |
| DevOps/SRE engineer | Scheduled reports and actionable local findings |
| Security engineer | Proof of what data leaves the cluster and assurance that secret values are never collected |
| Engineering lead | Easy rollout across clusters and reliable delivery to Slack/email/Kextant Cloud |
| Kextant Cloud backend | Stable manifest schema and agent behavior that can be processed deterministically |

---

## 4. Required product outcomes

The repository is successful when:

1. A user can install the agent with Helm in under 5 minutes.
2. The agent runs with read-only Kubernetes permissions only.
3. A user can run a one-shot scan and see a local health report.
4. A user can export a complete inventory manifest as JSON without uploading it.
5. A user can enable cloud upload and receive a Kextant Cloud report within 10 minutes of signup.
6. The manifest contains enough metadata for Kextant Cloud to identify at least 80% of common EKS/Kubernetes ecosystem components when the knowledge base supports them.
7. The agent never uploads Kubernetes Secret values, ConfigMap values, environment variable values, mounted file contents, pod logs, or application data.
8. The agent can be removed cleanly by deleting its namespace or Helm release.

---

## 5. Scope

### 5.1 In scope

- Kubernetes health checks already present in the current agent.
- Local report generation: plain text, logs, Slack, and email.
- Scheduled and one-time scan modes.
- Inventory manifest generation.
- Optional cloud upload of manifest data.
- Redaction controls for sensitive metadata.
- Helm chart installation.
- Strict read-only RBAC.
- Tests for scanner behavior, report generation, manifest serialization, and redaction.

### 5.2 Out of scope

- User accounts, billing, and dashboard UI.
- Knowledge-base enrichment.
- Cross-tenant data storage.
- CVE database ingestion.
- Hosted report history.
- Admission webhooks or deployment blocking.
- Automated upgrade pull requests.
- Cost optimization calculations beyond local metadata collection.

---

## 6. Functional requirements

### 6.1 Installation and packaging

| ID | Requirement | Priority | Acceptance criteria |
| --- | --- | --- | --- |
| AG-INST-001 | Provide a Helm chart as the primary installation method | Must | `helm install kextant ...` deploys namespace, service account, RBAC, config, and workload |
| AG-INST-002 | Keep raw Kubernetes manifests available for advanced/manual installs | Must | `deploy/` manifests remain valid and documented |
| AG-INST-003 | Support Deployment and CronJob modes | Should | User can choose always-on scheduled agent or scan-on-schedule CronJob |
| AG-INST-004 | Provide uninstall instructions | Must | Documentation includes Helm uninstall and namespace cleanup steps |
| AG-INST-005 | Publish container images with immutable version tags | Must | Releases include `vX.Y.Z` image tags; `latest` is not required for production use |
| AG-INST-006 | Support amd64 and arm64 images | Should | Multi-arch image manifest is published for releases |

### 6.2 CLI behavior

| ID | Requirement | Priority | Acceptance criteria |
| --- | --- | --- | --- |
| AG-CLI-001 | Keep `scan` command for one-shot health scan | Must | Existing usage continues to work |
| AG-CLI-002 | Keep `scan --send` command for configured local delivery | Must | Sends to enabled log/Slack/email targets |
| AG-CLI-003 | Add `inventory` command or equivalent `scan --manifest` mode | Must | Prints manifest JSON to stdout without uploading |
| AG-CLI-004 | Add `inventory --output <file>` | Should | Writes manifest JSON to file path |
| AG-CLI-005 | Add `inventory --upload` or `scan --upload` | Must | Uploads manifest to Kextant Cloud when configured |
| AG-CLI-006 | Add `validate` coverage for cloud config and redaction config | Must | Misconfigured cloud endpoint/API key fails validation before runtime |
| AG-CLI-007 | Add `version --json` | Should | Emits agent version, build SHA, manifest schema version |

### 6.3 Existing local health checks

The current health-check categories remain part of the product and must not regress.

| ID | Requirement | Priority |
| --- | --- | --- |
| AG-HEALTH-001 | Resource request/limit checks remain available | Must |
| AG-HEALTH-002 | Probe checks remain available | Must |
| AG-HEALTH-003 | Replica and PDB checks remain available | Must |
| AG-HEALTH-004 | Image tag/pull-policy checks remain available | Must |
| AG-HEALTH-005 | Security context checks remain available | Must |
| AG-HEALTH-006 | Deprecated API checks remain available | Must |
| AG-HEALTH-007 | Namespace policy checks remain available | Must |
| AG-HEALTH-008 | Global and resource-level exemptions remain supported | Must |
| AG-HEALTH-009 | Health score remains deterministic for the same input | Must |

### 6.4 Inventory manifest collection

The agent must produce a versioned manifest that Kextant Cloud can process without additional cluster access.

| ID | Requirement | Priority | Acceptance criteria |
| --- | --- | --- | --- |
| AG-MAN-001 | Include `schema_version` | Must | Initial value is `1.0.0` |
| AG-MAN-002 | Include `agent_version` and build metadata | Must | Backend can distinguish agent capabilities |
| AG-MAN-003 | Include `cluster_id` when configured | Must | Cloud submissions are associated with one cluster |
| AG-MAN-004 | Include scan timestamp in UTC | Must | RFC3339 timestamp is present |
| AG-MAN-005 | Include Kubernetes server version | Must | `major`, `minor`, and full GitVersion are captured where available |
| AG-MAN-006 | Include nodes with kubelet version, OS image, architecture, and optional redacted name | Must | Node version skew can be analyzed |
| AG-MAN-007 | Include workload image occurrences | Must | Deployments, StatefulSets, DaemonSets, ReplicaSets, Jobs, CronJobs, and Pods are scanned |
| AG-MAN-008 | Preserve workload context | Must | Each occurrence includes namespace, kind, resource name, container name, image reference, labels, and annotations after redaction rules |
| AG-MAN-009 | Deduplicate component images while preserving all affected locations | Should | Cloud can show one finding with multiple affected workloads |
| AG-MAN-010 | Include Helm v3 release metadata when enabled | Must | Release name, namespace, chart, chart version, app version, and status are captured |
| AG-MAN-011 | Include CRD metadata | Must | CRD name, group, versions, and scope are captured |
| AG-MAN-012 | Include redaction metadata | Must | Manifest states which redaction modes were applied |
| AG-MAN-013 | Exclude secret values and release payload bodies | Must | Tests prove no `data` values are serialized |
| AG-MAN-014 | Support local-only manifest export | Must | User can review exact JSON before upload |

### 6.5 Cloud upload

| ID | Requirement | Priority | Acceptance criteria |
| --- | --- | --- | --- |
| AG-UP-001 | Add configuration for `KEXTANT_CLOUD_ENABLED` | Must | Upload is disabled by default unless explicitly enabled |
| AG-UP-002 | Add configuration for cloud API endpoint | Must | Defaults to production endpoint but can be overridden for staging/dev |
| AG-UP-003 | Add cluster-scoped API key support | Must | API key is read from Kubernetes Secret/env var |
| AG-UP-004 | Add TLS verification | Must | HTTPS is required unless explicit dev override is set |
| AG-UP-005 | Add retry with exponential backoff | Should | Temporary network errors do not lose scheduled upload immediately |
| AG-UP-006 | Log upload success/failure without leaking credentials | Must | Logs include status and report ID if returned, never API key |
| AG-UP-007 | Support server response containing report URL or report ID | Should | User can see where to retrieve cloud report |
| AG-UP-008 | Fail local health report gracefully if cloud upload fails | Must | Local scan/report still completes when cloud is unavailable |
| AG-UP-009 | Add mTLS support before enterprise launch | Should | Client cert/key can be mounted from Secret |

### 6.6 Redaction and privacy controls

| ID | Requirement | Priority | Acceptance criteria |
| --- | --- | --- | --- |
| AG-PRIV-001 | Never collect Secret values | Must | Unit tests and code review verify this |
| AG-PRIV-002 | Never collect ConfigMap values | Must | ConfigMaps are not serialized into manifest values |
| AG-PRIV-003 | Never collect env var values | Must | Container env vars are ignored or names-only if explicitly added later |
| AG-PRIV-004 | Allow namespace redaction | Should | Namespace names can be replaced with stable hashes |
| AG-PRIV-005 | Allow workload name redaction | Should | Resource names can be replaced with stable hashes |
| AG-PRIV-006 | Allow label and annotation allowlists | Must | Default upload includes only safe matching labels/annotations |
| AG-PRIV-007 | Print privacy summary after manifest generation | Should | CLI reports counts and redaction settings |

### 6.7 RBAC

| ID | Requirement | Priority | Acceptance criteria |
| --- | --- | --- | --- |
| AG-RBAC-001 | All permissions are read-only | Must | ClusterRole contains no create/update/patch/delete/watch unless explicitly justified; list/get are default |
| AG-RBAC-002 | Include read access to workloads currently scanned | Must | Existing health scans continue working |
| AG-RBAC-003 | Add read access to nodes | Must | Node version analysis works |
| AG-RBAC-004 | Add read access to CRDs | Must | Operator detection works |
| AG-RBAC-005 | Helm secret read is optional and documented | Must | Users can install without secret read at lower accuracy |
| AG-RBAC-006 | Provide minimal and full RBAC profiles | Should | Minimal excludes Helm secrets; full includes optional metadata needed for best matching |

---

## 7. Manifest schema requirements

The initial manifest should follow this logical structure:

```json
{
  "schema_version": "1.0.0",
  "agent_version": "0.2.0",
  "cluster_id": "cluster_uuid_or_empty_for_local",
  "cluster_name": "configured_display_name",
  "scan_timestamp": "2026-05-29T00:00:00Z",
  "redaction": {
    "namespace_names": "plain|hashed|omitted",
    "workload_names": "plain|hashed|omitted",
    "labels": "allowlist|all|none",
    "annotations": "allowlist|all|none"
  },
  "kubernetes": {
    "server_version": "v1.33.2"
  },
  "nodes": [],
  "components": [],
  "helm_releases": [],
  "crds": []
}
```

### 7.1 Component occurrence fields

Each workload image occurrence must include:

- `image`
- `image_registry` when parseable
- `image_repository` when parseable
- `image_tag` when parseable
- `image_digest` when present in the reference
- `namespace`
- `kind`
- `name`
- `container_name`
- `labels`
- `annotations`
- `owner_references` if safely available

### 7.2 Helm release fields

Each Helm release entry must include:

- `name`
- `namespace`
- `chart`
- `chart_version`
- `app_version`
- `status`
- `updated_at` if available

### 7.3 CRD fields

Each CRD entry must include:

- `name`
- `group`
- `versions`
- `scope`
- `labels`
- `annotations`

---

## 8. Non-functional requirements

| Area | Requirement | Target |
| --- | --- | --- |
| Resource usage | Typical scan footprint | <50m CPU, <128Mi memory for common clusters |
| Runtime | Scan duration | <60 seconds for 500 workloads in normal API conditions |
| Manifest size | Default max | <10MB unless explicitly configured |
| Compatibility | Kubernetes versions | Support currently maintained Kubernetes minor versions plus best-effort for n-2 |
| Reliability | Scheduled scans | Agent continues running after failed delivery/upload |
| Security | Privileges | Non-root container, read-only root filesystem where practical, dropped Linux capabilities |
| Observability | Logs | Structured logs with scan status, counts, and errors |
| Testing | CI | Unit tests and linting must pass before release |

---

## 9. Integration with other Kextant repos

| Repo | Agent responsibility toward repo |
| --- | --- |
| `kextant-cloud` | Submit valid manifests; handle API responses; respect rate limits and auth errors |
| `kextant-kb` | Provide data fields required for image/chart/CRD matching |
| `kextant-web` | Support install commands generated by the web app; expose version and config validation for onboarding |
| `kextant-infra` | Use endpoints, certificates, and release artifacts provisioned by infrastructure |

---

## 10. Release requirements

A release is acceptable only when:

1. Version is tagged in Git.
2. Container image is built and published with the same version tag.
3. Helm chart version and app version are updated.
4. Changelog includes user-visible changes and migration notes.
5. RBAC changes are called out clearly.
6. Manifest schema changes are backward compatible or documented with migration guidance.
7. Tests pass.
8. A local kind/minikube smoke test succeeds.

---

## 11. Marketable launch acceptance criteria

For the Kextant product to be marketable, this repo must deliver:

- Helm install path that works on EKS without manual YAML editing.
- One-shot inventory export.
- Optional cloud upload.
- Safe default redaction.
- Clear explanation of data collected and not collected.
- Existing health reports preserved.
- Stable manifest schema consumed by Kextant Cloud.
- Documentation that lets a cautious platform/security engineer approve installation.

---

## 12. Immediate implementation backlog

1. Add manifest types.
2. Add inventory collector for nodes, workloads, Helm releases, and CRDs.
3. Add local manifest command.
4. Add redaction configuration.
5. Add cloud upload client.
6. Add Helm chart.
7. Update RBAC profiles.
8. Add tests for privacy and manifest schema.
9. Update README with new install and privacy model.
