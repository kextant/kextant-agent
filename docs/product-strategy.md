# Kextant Product Strategy

## Strategic thesis

Kextant should become a **Kubernetes Version Intelligence platform** with health reporting as the adoption wedge.

The original health-report concept is useful, but the stronger opportunity is the live-cluster version-drift problem: Kubernetes clusters accumulate dozens of operators, charts, databases, controllers, observability tools, and security components. Teams rarely know which versions are EOL, abandoned, severely behind, or risky to upgrade until a manual audit exposes the problem.

Kextant's differentiated promise:

> Weekly actionable Kubernetes version intelligence reports. No dashboard babysitting. No manual spreadsheet. Know what is outdated, unsupported, abandoned, vulnerable, and what to do next.

## What we should build

### Core product

A report-first SaaS that combines:

1. **In-cluster deterministic agent**
   - Read-only scanner running in the customer's cluster.
   - Collects health findings and a structured inventory manifest.
   - No LLM or opaque analysis inside the customer cluster.

2. **Version intelligence knowledge base**
   - Canonical mapping of Kubernetes ecosystem components to current versions, support status, EOL dates, CVEs, deprecations, replacements, and upgrade paths.
   - Curated first, automated over time.

3. **Matching and prioritization engine**
   - Maps live images, Helm releases, CRDs, and Kubernetes/node versions to known components.
   - Produces deterministic severity-ranked findings.

4. **Report delivery**
   - Slack/email reports that are readable by platform teams and engineering leadership.
   - Minimal dashboard only for setup, history, billing, and drill-down.

### Product modules

| Module | Role | Monetization |
| --- | --- | --- |
| Local health checks | Free wedge, trust builder, immediate value | Free/OSS |
| Inventory manifest | Required data capture for version intelligence | Free agent capability |
| Version currency and EOL | Main paid value | Pro |
| Abandoned/deprecated project detection | Main paid value | Pro/Business |
| Upgrade paths and breaking changes | Differentiator | Pro/Business |
| CVE correlation by component version | Security-budget justification | Business/Enterprise |
| Multi-cluster fleet view | Expansion | Business/Enterprise |
| Admission/policy gates and GitOps PRs | Advanced automation | Enterprise |
| Cost optimization | Future upsell, not launch wedge | Business/Enterprise later |

## Ideal customer profile

### Launch ICP

Small-to-mid Kubernetes teams, especially AWS EKS users:

- 5-50 engineers.
- 1-10 Kubernetes clusters.
- No dedicated platform security team.
- Uses Helm, ArgoCD/Flux, operators, observability stacks, ingress controllers, databases, and cloud add-ons.
- Has felt pain from surprise upgrades, EOL software, or neglected internal platform dependencies.
- Wants passive reports instead of another dashboard.

### Buyer and users

| Person | Motivation |
| --- | --- |
| Platform lead | Wants confidence that cluster software is maintained and upgradeable |
| DevOps/SRE engineer | Wants prioritized remediation steps without manual research |
| Security engineer | Wants evidence of patch posture and unsupported software exposure |
| Engineering manager/CTO | Wants leadership-readable risk summaries and fewer surprise incidents |

## Jobs to be done

1. **When I inherit or operate a cluster**, I want to know what third-party software is running so I can understand my risk surface.
2. **When upstream projects release new versions**, I want to know if I am behind and whether the upgrade matters.
3. **When a component is EOL or abandoned**, I want to know before it becomes an incident or audit finding.
4. **When planning upgrades**, I want a safe path, known breaking changes, and migration guidance.
5. **When reporting to leadership**, I want a concise digest with severity and business impact.

## Positioning

### Category

Kubernetes Version Intelligence / Cluster Dependency Intelligence.

### One-liner

Kextant finds outdated, EOL, abandoned, and risky Kubernetes components running in your cluster and sends prioritized upgrade reports to Slack/email.

### Longer positioning

Kextant is for Kubernetes teams that do not have time to manually track every Helm chart, operator, controller, and database version across clusters. Unlike source dependency bots or vulnerability scanners, Kextant scans what is actually running, understands Kubernetes ecosystem components, and explains the upgrade path in a weekly report.

## Competitive differentiation

| Competitor type | What they do | Kextant wedge |
| --- | --- | --- |
| Renovate/Dependabot | Source dependency PRs | Kextant scans live clusters, including Helm/operator installs and drift outside source repos |
| Snyk/Trivy | CVE and image vulnerability scanning | Kextant adds EOL, abandoned-project, release currency, upgrade path, and breaking-change context |
| Fairwinds/Polaris | Kubernetes configuration best practices | Kextant focuses on version intelligence plus passive reports |
| Kubecost/OpenCost | Cost visibility | Kextant may add cost later, but launches on version risk |
| endoflife.date | Manual EOL lookup | Kextant automates discovery, mapping, reports, and prioritization |

## Marketable product definition

A first marketable product is ready when a user can:

1. Install the Kextant agent with a one-command Helm install.
2. Register a cluster to Kextant Cloud.
3. Receive a Slack/email report within five minutes.
4. See at least 80% of common components identified in a realistic EKS cluster.
5. Get severity-ranked findings for EOL, abandoned, deprecated, and significantly outdated components.
6. Understand exactly what to do next for each critical finding.
7. Start a paid subscription without a sales call.

The product does **not** need a complex dashboard to be marketable. The report is the product.

## Non-goals for launch

- Full enterprise compliance suite.
- Admission controller that blocks deploys.
- Automated upgrade PRs.
- Deep cloud cost optimization.
- Air-gapped/on-prem server.
- Perfect coverage of every obscure component.
- Real-time alerting for every release.

## Success metrics

### Product quality

- 80%+ identification rate for top common components in beta clusters.
- Less than 5% high-severity false positives.
- Report generated in under 30 seconds for a 500-component manifest.
- Agent install to first report in under 10 minutes.

### Business

- 100 qualified waitlist signups before paid launch.
- 20 beta clusters scanned.
- 10 paying customers or 25 paid clusters within 60 days of paid launch.
- $2k MRR milestone from self-serve customers.

### Adoption

- 200 free agent installs or one-time audits.
- 20%+ free-to-trial conversion from clusters with at least one critical finding.
- 5%+ trial-to-paid conversion initially; improve toward 10-15% with onboarding.

## Assumptions to validate

1. Teams will upload a sanitized cluster inventory manifest to a SaaS if the agent is open and read-only.
2. Version intelligence reports create more willingness to pay than generic health reports.
3. $39-$99 per cluster per month is acceptable for small/mid teams.
4. A curated top-100/top-200 component KB is enough for a compelling first product.
5. Slack/email report-first UX is a differentiator, not just a missing dashboard feature.
