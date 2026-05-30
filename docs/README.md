# Kextant Agent Documentation

`PRODUCT_REQUIREMENTS.md` is authoritative for this repository and takes priority over every other document.

## Operator and developer docs

| Document | Purpose |
| --- | --- |
| [Product Requirements](PRODUCT_REQUIREMENTS.md) | Authoritative requirements for the agent product |
| [Installation](INSTALLATION.md) | Helm and raw manifest installation guide |
| [Manifest Schema](MANIFEST_SCHEMA.md) | v0.1.0 inventory manifest schema |
| [Privacy](PRIVACY.md) | RBAC, redaction, and data collection model |
| [Development](DEVELOPMENT.md) | Local development, testing, and release workflow |
| [CI/CD](CI_CD.md) | GitHub Actions build, test, and image publishing pipeline |
| [Exemptions](EXEMPTIONS.md) | Health-check exemption configuration |
| [Changelog](CHANGELOG.md) | Release notes and user-visible changes |

## Product strategy docs

These documents capture the broader Kextant product direction. They are useful context, but they do not override `PRODUCT_REQUIREMENTS.md` for implementation decisions in this repo.

| Document | Purpose |
| --- | --- |
| [Product Strategy](product-strategy.md) | Positioning, ICP, differentiation, and success metrics |
| [Monetization Plan](monetization-plan.md) | Packaging, pricing, funnel, and revenue milestones |
| [Market Requirements](requirements.md) | Cross-product MVP and paid-product requirements |
| [Technical Plan](technical-plan.md) | Target architecture and repo boundaries |
| [Roadmap](roadmap.md) | Phased path to a marketable paid product |

## Current implementation focus

The v0.1.0 prototype focuses on the agent repository responsibilities:

- read-only Kubernetes health scanning
- deterministic inventory manifest generation
- local manifest export
- optional cloud upload
- privacy/redaction controls
- Helm installation
- containerized build and GitHub Actions CI/CD
