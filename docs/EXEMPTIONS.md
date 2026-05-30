# Health Check Exemptions

`docs/PRODUCT_REQUIREMENTS.md` is authoritative. This document describes the current exemption model for local health checks.

## Global exemptions

Set `EXEMPT_CHECKS` to a comma-separated list of check IDs:

```yaml
EXEMPT_CHECKS: "REP001,IMG001"
```

Use `*` to exempt all checks. This should be rare and temporary.

## Resource-level exemptions

Annotate a Kubernetes resource with `kextant.com/skip-checks`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
  annotations:
    kextant.com/skip-checks: "REP001,IMG001"
spec:
  replicas: 1
```

Pod-based checks inherit exemptions from owning ReplicaSets and Deployments where owner references allow resolution.

## Supported check IDs

### Resources

- `RES001`: Missing CPU requests
- `RES002`: Missing CPU limits
- `RES003`: Missing memory requests
- `RES004`: Missing memory limits
- `RES005`: CPU limit equals request

### Probes

- `PRB001`: Missing readiness probe
- `PRB002`: Missing liveness probe
- `PRB003`: Missing startup probe
- `PRB004`: Identical liveness/readiness probes

### Replicas

- `REP001`: Single replica deployment
- `REP002`: No PodDisruptionBudget
- `REP003`: Mismatched desired vs available replicas

### Images

- `IMG001`: `:latest` tag
- `IMG002`: no image tag
- `IMG003`: ImagePullPolicy not set
- `IMG004`: ImagePullPolicy Always with a specific tag

### Security

- `SEC001`: Running as root
- `SEC002`: Privileged container
- `SEC003`: Writable root filesystem
- `SEC004`: Privilege escalation allowed
- `SEC005`: All capabilities not dropped

### APIs

- `API001`: Deprecated API version
- `API002`: Removed API version

### Namespaces

- `NS001`: No ResourceQuota
- `NS002`: No LimitRange
- `NS003`: No NetworkPolicy
