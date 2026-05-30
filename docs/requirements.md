# Kextant Product Requirements

## 1. Product goal

Deliver a marketable self-serve product that scans a Kubernetes cluster, identifies live third-party component version risk, and sends a prioritized report to Slack/email.

The existing health scanner remains part of the product, but the paid MVP centers on version intelligence.

## 2. MVP definition

The MVP is complete when a beta user can:

1. Create an account and cluster in Kextant Cloud.
2. Install the Kextant agent using Helm.
3. Run a scan with read-only permissions.
4. Submit a deterministic cluster inventory manifest to Kextant Cloud.
5. Receive a report identifying common outdated, EOL, deprecated, abandoned, or version-skewed components.
6. See upgrade/migration guidance for critical findings.
7. Configure Slack/email delivery.
8. Upgrade to a paid plan through Stripe.

## 3. Personas

| Persona | Need |
| --- | --- |
| Platform engineer | Know what cluster components need upgrades and which are risky |
| DevOps/SRE lead | Prioritize remediation across workloads and namespaces |
| Security engineer | Identify unsupported, vulnerable, or abandoned software in production |
| Engineering manager/CTO | Receive a readable risk summary without reading raw scanner output |

## 4. Scope

### In scope for paid MVP

- Existing local Kubernetes health checks.
- Agent inventory manifest for live cluster components.
- Cloud manifest ingestion.
- Knowledge base for the top 100-200 Kubernetes ecosystem components.
- Deterministic image/Helm/CRD-to-component matching.
- Severity-ranked version intelligence report.
- Slack and email delivery.
- Minimal account, cluster, report history, and billing UI.

### Out of scope for paid MVP

- Admission controller/policy blocking.
- Automated GitOps pull requests.
- Full compliance frameworks.
- Deep cost optimization.
- On-prem/air-gapped deployment.
- Comprehensive long-tail component coverage.
- Real-time dashboards or alert streams.

## 5. Functional requirements

### 5.1 Agent: existing health report capability

The current agent already supports many of these requirements and should continue to do so.

| ID | Requirement | Priority |
| --- | --- | --- |
| AG-H-001 | Run read-only Kubernetes health checks for resources, probes, replicas, images, security context, APIs, and namespaces | Must |
| AG-H-002 | Generate local reports in plain text, HTML email, Slack, and logs | Must |
| AG-H-003 | Support scheduled and one-shot scans | Must |
| AG-H-004 | Support global and resource-level exemptions | Must |
| AG-H-005 | Remain usable without Kextant Cloud | Must |

### 5.2 Agent: inventory manifest

| ID | Requirement | Priority |
| --- | --- | --- |
| AG-I-001 | Collect Kubernetes server version and node kubelet versions | Must |
| AG-I-002 | Collect workload images from Deployments, StatefulSets, DaemonSets, ReplicaSets, Jobs, CronJobs, and Pods | Must |
| AG-I-003 | Collect namespace, kind, name, labels, annotations, image, and image tag for each component occurrence | Must |
| AG-I-004 | Collect Helm release metadata from Helm v3 release secrets when enabled | Must |
| AG-I-005 | Collect CRDs with group, versions, and names to assist operator detection | Must |
| AG-I-006 | Deduplicate repeated images while preserving affected locations | Must |
| AG-I-007 | Export manifest to local JSON without cloud upload | Must |
| AG-I-008 | Submit manifest to Kextant Cloud using API key and TLS/mTLS | Must |
| AG-I-009 | Allow users to redact labels/annotations/namespaces through config | Should |
| AG-I-010 | Include agent version and manifest schema version | Must |

### 5.3 Agent configuration and RBAC

| ID | Requirement | Priority |
| --- | --- | --- |
| AG-C-001 | Add cloud endpoint, cluster ID, API key/client certificate, and upload toggle config | Must |
| AG-C-002 | Add RBAC for nodes, CRDs, and Helm secrets | Must |
| AG-C-003 | Make Helm secret scanning optional and documented | Must |
| AG-C-004 | Provide Helm chart installation path | Must |
| AG-C-005 | Keep resource footprint below 50m CPU and 128Mi memory for normal scans | Should |

### 5.4 Kextant Cloud: onboarding and tenancy

| ID | Requirement | Priority |
| --- | --- | --- |
| CL-O-001 | User can create an account | Must |
| CL-O-002 | User can create a cluster and receive install credentials | Must |
| CL-O-003 | User can rotate/revoke cluster credentials | Must |
| CL-O-004 | User can configure Slack webhook and email recipients | Must |
| CL-O-005 | User can view latest report and report history | Must |
| CL-O-006 | User can manage subscription through Stripe | Must |

### 5.5 Kextant Cloud: manifest ingestion

| ID | Requirement | Priority |
| --- | --- | --- |
| CL-I-001 | Accept manifest submissions over authenticated HTTPS | Must |
| CL-I-002 | Validate manifest schema version and size | Must |
| CL-I-003 | Store raw manifest encrypted at rest or in encrypted object storage | Must |
| CL-I-004 | Normalize image references and tags | Must |
| CL-I-005 | Trigger asynchronous report generation | Must |
| CL-I-006 | Reject submissions over tier rate limits | Must |

### 5.6 Knowledge base

| ID | Requirement | Priority |
| --- | --- | --- |
| KB-001 | Define canonical component schema: name, aliases, images, charts, project URL, status, versions, latest stable, EOL/deprecation data | Must |
| KB-002 | Seed KB with top 100 Kubernetes ecosystem components | Must |
| KB-003 | Track abandoned/deprecated replacements for known components | Must |
| KB-004 | Track Kubernetes version and node version support policy | Must |
| KB-005 | Track confidence score for each KB field | Must |
| KB-006 | Add manual admin workflow for correcting KB entries | Must |
| KB-007 | Add automated ingestion from GitHub Releases, Artifact Hub, endoflife.date, NVD/GHSA after manual seed | Should |

### 5.7 Matching engine

| ID | Requirement | Priority |
| --- | --- | --- |
| ME-001 | Match Helm releases to KB components using chart name/appVersion | Must |
| ME-002 | Match images to KB components using known canonical image patterns | Must |
| ME-003 | Match CRDs to known operators where possible | Must |
| ME-004 | Produce a confidence score for each match | Must |
| ME-005 | Mark unknown components as unknown rather than guessing | Must |
| ME-006 | Group duplicate component occurrences across namespaces/workloads | Must |
| ME-007 | Identify Kubernetes node/control-plane skew and EOL versions | Must |

### 5.8 Findings and severity

| ID | Requirement | Priority |
| --- | --- | --- |
| FI-001 | Detect abandoned projects | Must |
| FI-002 | Detect deprecated projects with replacement guidance | Must |
| FI-003 | Detect EOL versions | Must |
| FI-004 | Detect major/minor version gaps based on component policy | Must |
| FI-005 | Detect cluster autoscaler/Kubernetes version mismatch | Must |
| FI-006 | Detect multiple major versions of same component where relevant | Should |
| FI-007 | Correlate CVEs to affected component versions | Business tier / Should after MVP |
| FI-008 | Summarize breaking changes for major upgrade paths | Should |

### 5.9 Reports

| ID | Requirement | Priority |
| --- | --- | --- |
| RP-001 | Generate report summary with score, critical/warning/info counts, identified/unknown counts | Must |
| RP-002 | List findings in priority order with current version, latest version, status, risk, and action | Must |
| RP-003 | Include affected namespaces/workloads for each finding | Must |
| RP-004 | Include match confidence and clearly mark unverified data | Must |
| RP-005 | Provide Slack Block Kit report | Must |
| RP-006 | Provide HTML email report | Must |
| RP-007 | Provide JSON report through API | Must |
| RP-008 | Include trend delta compared with previous report | Should |

## 6. Manifest schema requirements

The agent manifest must include:

- `schema_version`
- `agent_version`
- `cluster_id`
- `cluster_name` or pseudonymous display name
- `scan_timestamp`
- Kubernetes server version
- Nodes with kubelet version, OS image, architecture
- Component occurrences with image, namespace, kind, name, labels, annotations
- Helm releases with name, namespace, chart, chart version, app version, status
- CRDs with group, names, versions
- Redaction metadata showing whether names/labels/annotations were redacted

The manifest must not include secret values, ConfigMap data values, environment variable values, or pod logs.

## 7. Non-functional requirements

| Requirement | Target |
| --- | --- |
| Agent permissions | Read-only; no mutation verbs |
| Agent resource usage | <50m CPU and <128Mi memory typical |
| Manifest max size | 10MB for MVP |
| Report latency | <30 seconds for 500 components |
| API uptime | 99.5% beta, 99.9% after paid launch |
| Data retention | Free 14 days, Pro 90 days, Business 1 year |
| Encryption in transit | TLS 1.3 where supported |
| Encryption at rest | Managed disk/object-store encryption |
| Tenant isolation | No cross-tenant data access paths |
| Auditability | Log manifest submissions, report access, credential rotation |

## 8. Acceptance criteria for marketable launch

- A fresh EKS cluster with common add-ons can be scanned in under 10 minutes from signup.
- A real production-like cluster report identifies at least 80% of top common components.
- The report detects reference findings such as Kiam abandoned, Kubernetes dashboard severely outdated, External Secrets behind, ClickHouse EOL, and Cluster Autoscaler/Kubernetes mismatch when present.
- A user can understand the top three actions without reading documentation.
- A user can pay and unlock Pro features without manual intervention.
- Unknown components are safely reported as unknown with no invented facts.

## 9. Open questions

1. Should the first cloud auth method be API key only, or API key plus mTLS from day one?
2. Should the first cloud backend use AWS serverless or a simpler managed app + Postgres deployment?
3. How aggressively should resource names be redacted by default?
4. Should Helm secret access be enabled by default or opt-in during install?
5. What exact component count belongs in the free tier: top 25, 50, or 100?
6. Which top 100 components are required for launch coverage?
