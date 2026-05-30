# Kextant Roadmap to a Marketable Product

## Guiding rule

Build the smallest product that proves teams will pay for live-cluster version intelligence. The first sellable artifact is a high-quality report, not a full dashboard.

## Phase 0: Validate positioning and prepare launch surface

**Timeline:** 1 week

### Outcomes

- Landing page explains version intelligence, not generic health checks.
- Waitlist/email capture is live.
- Example report is published with realistic findings.
- Top 100 component KB list is drafted.

### Tasks

- [ ] Update public messaging: "Kubernetes version intelligence reports".
- [ ] Create a sample report from the real production audit findings.
- [ ] Publish landing page at `kextant.com`.
- [ ] Add waitlist form and analytics.
- [ ] Identify 10-20 DevOps/platform contacts for beta interviews.
- [ ] Decide initial backend stack and repo split.

### Exit criteria

- 25 qualified waitlist signups or 5 direct beta commitments.
- Clear install-to-report demo narrative.

## Phase 1: Agent inventory foundation

**Timeline:** 2 weeks

### Outcomes

- The current agent can export a JSON inventory manifest.
- The manifest includes images, Helm releases, CRDs, nodes, and Kubernetes version.
- No cloud dependency required.

### Tasks in `kextant-agent`

- [ ] Add `pkg/types/manifest.go`.
- [ ] Add `internal/inventory` collector.
- [ ] Add node, CRD, workload, and Helm metadata collection.
- [ ] Add local command: `kextant-agent inventory` or `scan --manifest`.
- [ ] Add redaction configuration.
- [ ] Update RBAC for nodes and CRDs.
- [ ] Make Helm secret read opt-in or clearly documented.
- [ ] Add tests for manifest serialization and redaction.
- [ ] Update README/install docs.

### Exit criteria

- Run against a real cluster and produce a stable manifest.
- Manifest contains enough data to identify at least 50 common components manually.
- No secret values are present in manifest output.

## Phase 2: Static KB and offline report demo

**Timeline:** 2-3 weeks

### Outcomes

- A small curated KB can generate a compelling report from a manifest.
- This can be demoed before building full SaaS plumbing.

### Tasks

- [ ] Define KB schema.
- [ ] Seed top 100 Kubernetes components.
- [ ] Add known reference findings from the production audit.
- [ ] Build matching engine prototype.
- [ ] Generate JSON and HTML report from manifest + static KB.
- [ ] Add confidence scoring and unknown component handling.
- [ ] Create sample Slack/email renderings.

### Exit criteria

- Report identifies Kiam, External Secrets, kube-state-metrics, ClickHouse, Cluster Autoscaler, Jaeger, Grafana Agent/Alloy, Kubernetes Dashboard, Loki, Prometheus, Argo CD, Istio, etc. when present.
- Top findings have clear action text and source links.
- False positives are visibly controlled by confidence scoring.

## Phase 3: Cloud ingestion and report delivery

**Timeline:** 3-4 weeks

### Outcomes

- Beta users can connect a cluster and receive cloud-generated reports.

### Tasks in new `kextant-cloud` repo

- [ ] Create organization/user/cluster data model.
- [ ] Create API key based cluster authentication.
- [ ] Implement `POST /v1/manifests`.
- [ ] Store manifests and reports.
- [ ] Run matching/report worker asynchronously.
- [ ] Send email reports.
- [ ] Send Slack reports.
- [ ] Add basic admin view for KB corrections.

### Tasks in `kextant-agent`

- [ ] Add cloud upload config.
- [ ] Add upload client with retry/backoff.
- [ ] Add install docs for beta clusters.

### Exit criteria

- New beta user can install agent and receive first report in under 10 minutes.
- Cloud report matches offline report for same manifest.
- Report delivery works through both Slack and email.

## Phase 4: Self-serve onboarding and payments

**Timeline:** 2-3 weeks

### Outcomes

- Product can charge money without manual intervention.

### Tasks in `kextant-web`

- [ ] Create landing page + app shell.
- [ ] Add signup/login.
- [ ] Add cluster creation and install command page.
- [ ] Add latest report page.
- [ ] Add report history page.
- [ ] Add delivery target configuration.
- [ ] Add Stripe Checkout and Billing Portal.
- [ ] Gate Pro features by subscription tier.

### Exit criteria

- User can sign up, install, get report, and pay.
- Free vs Pro gates are enforced.
- Billing events update subscription state.

## Phase 5: Paid beta

**Timeline:** 4-6 weeks

### Outcomes

- Validate willingness to pay and report usefulness.
- Improve KB coverage based on real clusters.

### Tasks

- [ ] Invite 10-20 beta users.
- [ ] Offer founder pricing in exchange for feedback.
- [ ] Run weekly customer interviews.
- [ ] Track unknown components and add the most common to KB.
- [ ] Add report delta/trend section.
- [ ] Add source links for all critical findings.
- [ ] Publish 2-3 technical blog posts.

### Exit criteria

- 10 paying customers or 25 paid clusters.
- 80%+ identification rate across beta manifests for common components.
- At least 3 testimonials or case-study quotes.
- Churn reasons and objections are documented.

## Phase 6: Public launch

**Timeline:** after paid beta exit criteria

### Outcomes

- Public launch with confidence in pricing, report quality, and onboarding.

### Tasks

- [ ] Publish public docs and Helm install guide.
- [ ] Launch on Product Hunt and Hacker News.
- [ ] Post in r/kubernetes, r/devops, CNCF Slack, relevant Discords.
- [ ] Publish "State of Kubernetes Version Drift" article using anonymized beta data.
- [ ] Add comparison pages vs Renovate, Dependabot, Snyk, Fairwinds, Kubescape.
- [ ] Add AWS Marketplace evaluation if EKS traction is strong.

### Exit criteria

- $2k MRR.
- Repeatable self-serve onboarding.
- Support load is manageable nights/weekends.

## Backlog after launch

### Product

- CVE correlation per component version.
- Multi-cluster fleet report.
- Report trends and remediation progress.
- ArgoCD/Flux PR suggestions.
- Policy/admission webhook for EOL blockers.
- Cost optimization module.
- Compliance mapping.

### Technical

- Automated KB ingestion from GitHub Releases, Artifact Hub, endoflife.date, NVD/GHSA.
- mTLS cluster credentials.
- Registry digest matching.
- Public API for reports.
- SSO/SAML.
- SOC2 readiness.

### Go-to-market

- Component-specific SEO pages.
- Partner program for DevOps consultancies.
- AWS Marketplace listing.
- Case studies and benchmark reports.

## Launch readiness checklist

- [ ] Agent is installable with Helm.
- [ ] Agent manifest is documented and safe.
- [ ] Cloud onboarding takes less than 10 minutes.
- [ ] Reports are readable in Slack and email.
- [ ] Paid subscription works end to end.
- [ ] Free tier limits are enforced.
- [ ] Critical findings include source links.
- [ ] Unknown components are tracked for KB improvement.
- [ ] Terms/privacy policy exist.
- [ ] Support email and feedback path exist.
- [ ] Monitoring/alerts exist for ingestion and report failures.

## Highest-risk assumptions to test first

1. Users are willing to upload a sanitized manifest.
2. The top 100-200 component KB is enough to make reports valuable.
3. Users will pay at least $39/cluster/month.
4. Report-first UX is sufficient without a rich dashboard.
5. KB maintenance can be automated enough to keep support burden low.
