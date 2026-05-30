# Kextant Agent Installation

`docs/PRODUCT_REQUIREMENTS.md` is authoritative. This guide explains how to install and operate the v0.1.0 prototype.

## Prerequisites

- Kubernetes cluster with access to create a namespace, ServiceAccount, ClusterRole, and ClusterRoleBinding.
- Helm 3 for the recommended install path.
- Network egress to Kextant Cloud only if cloud upload is enabled.

## Helm install

```bash
helm install kextant charts/kextant-agent \
  --namespace kextant \
  --create-namespace
```

Default behavior:

- Runs as a Deployment.
- Uses read-only Kubernetes RBAC.
- Sends local health reports to logs.
- Does not upload data to Kextant Cloud.
- Does not read Helm release Secrets unless explicitly enabled.

## Enable Kextant Cloud upload

Create a cluster in Kextant Cloud, then install with the generated cluster ID and API key:

```bash
helm install kextant charts/kextant-agent \
  --namespace kextant \
  --create-namespace \
  --set config.clusterName=prod-us-east-1 \
  --set config.clusterID=<cluster-id> \
  --set config.cloud.enabled=true \
  --set secrets.cloudAPIKey=<cluster-api-key>
```

Cloud upload submits only the inventory manifest. It does not upload Secret values, ConfigMap values, environment variable values, pod logs, or application data.

## Enable Helm release metadata

Helm metadata improves chart/app version matching. It requires read access to Helm release Secrets.

```bash
helm upgrade kextant charts/kextant-agent \
  --namespace kextant \
  --reuse-values \
  --set config.inventory.helmMetadataEnabled=true \
  --set rbac.helmSecrets=true
```

Kextant decodes Helm release metadata to extract release name, chart, chart version, app version, status, and update timestamp. It does not serialize arbitrary Secret values.

## CronJob mode

For scan-on-schedule operation without a long-running pod:

```bash
helm install kextant charts/kextant-agent \
  --namespace kextant \
  --create-namespace \
  --set mode=cronjob \
  --set cronJob.schedule="0 9 * * 1"
```

## Local commands inside the pod

```bash
kubectl exec -n kextant deploy/kextant-agent -- /app/kextant-agent version --json
kubectl exec -n kextant deploy/kextant-agent -- /app/kextant-agent scan
kubectl exec -n kextant deploy/kextant-agent -- /app/kextant-agent inventory
kubectl exec -n kextant deploy/kextant-agent -- /app/kextant-agent inventory --upload
```

The deployment name differs when installed through Helm. Use:

```bash
kubectl get deploy -n kextant
```

## Raw manifest install

Raw manifests are provided in `deploy/` for manual usage:

```bash
kubectl apply -f deploy/
```

The Helm chart is preferred because it supports configurable minimal/full RBAC profiles.

## Uninstall

Helm:

```bash
helm uninstall kextant -n kextant
kubectl delete namespace kextant
```

Raw manifests:

```bash
kubectl delete -f deploy/
```
