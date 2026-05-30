# Kextant Monetization Plan

## Monetization principle

Do not monetize the agent itself. Monetize the intelligence that is expensive to build and maintain:

- Component identification across live clusters.
- Version, EOL, deprecation, and abandoned-project knowledge.
- Upgrade paths and breaking-change summaries.
- CVE correlation and compliance-oriented evidence.
- Report history, multi-cluster rollups, APIs, and enterprise controls.

This keeps adoption low-friction while creating a defensible paid product around the knowledge base.

## Packaging

### Recommended launch tiers

| Tier | Price | Target customer | Included |
| --- | --- | --- | --- |
| Free | $0 | Evaluation, hobby, small teams | OSS/local health reports, 1 cloud cluster, weekly limited version report, top 50 components, email only, 14-day report history |
| Pro | $39/cluster/mo annual or $49 month-to-month | Startups and small platform teams | Full version currency for top 500 components, EOL/deprecated/abandoned detection, upgrade path summaries, Slack + email, daily/weekly reports, 90-day history |
| Business | $99/cluster/mo annual or $129 month-to-month | Mid-market teams with security pressure | Everything in Pro plus CVE correlation, breaking-change detail, multi-cluster summary, API access, 1-year history, priority support |
| Enterprise | Custom; suggested minimum $500/mo or $199+/cluster/mo | Regulated or larger orgs | SSO/SAML, audit logs, custom component tracking, compliance mapping, custom retention, private/on-prem options, procurement/security review |

### Why this packaging

- **Free** drives installs and trust.
- **Pro** maps to the original $39/cluster/month plan and is enough to hit the first passive-income target.
- **Business** reserves security-budget features for buyers with higher willingness to pay.
- **Enterprise** avoids underpricing support-heavy customers.

The earlier $79 Enterprise idea can be used as a beta discount, but the final packaging should not put CVE correlation, compliance, SSO, or custom support into a low-price tier.

## What to charge for first

Charge for features that depend on Kextant Cloud and the knowledge base:

1. Full component coverage beyond the free limit.
2. EOL, deprecation, and abandoned-project findings.
3. Upgrade path and breaking-change intelligence.
4. Slack delivery and custom schedules.
5. Report history and trend tracking.
6. CVE correlation.
7. Multi-cluster rollups.
8. API/GitOps integrations.

Do **not** charge for basic local health checks at launch. They are the acquisition wedge.

## Funnel

### Acquisition offers

1. **Free one-time cluster audit**
   - `helm install` or CLI command generates a sanitized inventory.
   - User receives a sample version-risk report.
   - Report includes upgrade CTA when critical findings are found.

2. **Open-source agent**
   - Builds trust for read-only access.
   - Gives immediate local health-report value.
   - Provides distribution through GitHub, Helm, and community posts.

3. **Content marketing**
   - "The State of Version Drift in Kubernetes"
   - "How to find abandoned Helm charts in EKS"
   - "Kiam is dead: how to migrate to IRSA or EKS Pod Identity"
   - Component-specific SEO pages for EOL and migration guides.

4. **Consultancy/MSP partnerships**
   - DevOps consultancies can run Kextant audits for clients.
   - Offer partner codes or revenue share after product-market fit.

### Conversion mechanics

- Report shows identified risks and locked paid value:
  - Free: "12 outdated components detected; 3 critical details available in Pro."
  - Pro: "CVE correlation and multi-cluster impact available in Business."
- Upgrade should be self-serve through Stripe.
- Paid trial should start automatically after first report with critical findings.

## Revenue targets

### First target: $2k MRR

Possible paths:

| Mix | MRR |
| --- | --- |
| 52 Pro clusters at $39 | $2,028 |
| 21 Business clusters at $99 | $2,079 |
| 25 Pro clusters + 10 Business clusters | $1,965 |
| 5 small customers with 5 Pro clusters each + 5 Business clusters | $1,470 + $495 = $1,965 |

### Practical year-one target

- 200 free installs/audits.
- 40 paying customers.
- 100 paid clusters.
- Blended ARPA: $50-$75/cluster/month.
- Target MRR: $5k-$7.5k.

## Unit economics

Expected COGS per cluster should remain low if analysis is mostly batch/deterministic:

| Cost item | Target/month/cluster |
| --- | --- |
| API + worker compute | <$1.00 |
| Database/object storage | <$1.00 |
| Knowledge-base refresh amortized | $1.00-$3.00 |
| Email/Slack delivery | <$0.25 |
| LLM enrichment amortized | $0.50-$2.00 |
| Total target COGS | <$7.00 |

At $39/cluster/month, gross margin should be 80%+ if support is low-touch.

## Pricing tests

Run these pricing experiments during beta:

1. **Pro $39 vs $49 monthly**
   - Measure conversion resistance.
2. **Business $99 vs $129 monthly**
   - Useful once CVE correlation is available.
3. **Per-cluster vs bundled team pricing**
   - Example: $99/month includes 3 clusters, then $29/additional cluster.
4. **Annual discount**
   - 2 months free for annual plans.
5. **Founder plan**
   - First 25 customers lock $39/cluster/month for 12 months in exchange for feedback and public testimonial.

## Retention levers

Kextant risks becoming a one-time audit unless reports continue to produce value. Retention should come from:

- New release and EOL monitoring.
- Trend reports showing drift growth or remediation progress.
- New CVEs mapped to currently deployed versions.
- Multi-cluster comparisons.
- Report archives for audits and leadership updates.
- Upgrade planning workflows.

## Revenue expansion path

1. **Cluster expansion:** charge per cluster as customers install Kextant across staging/prod/fleet clusters.
2. **Security expansion:** upgrade from Pro to Business for CVE correlation and audit evidence.
3. **Enterprise expansion:** custom components, SSO, audit logs, on-prem/private deploy.
4. **Automation expansion:** GitOps PR generation, admission policies, CI/CD gates.
5. **Future module:** cost optimization once the version intelligence product is converting.

## Business cautions

- The KB creates the moat but also creates maintenance obligations. This is not fully passive until enrichment and QA are automated.
- Avoid enterprise-only features too early; procurement and SOC2 can slow progress.
- Avoid dashboard creep. The paid product is recurring intelligence delivered where users already work.
- Unknown or low-confidence findings should be labeled clearly. Trust is more important than looking comprehensive.
