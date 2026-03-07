# Secrets

Central secrets management layer. Creates Secret Manager secret shells and IAM bindings. Secret values are written by downstream layers.

## What it creates

- Secret Manager API enablement
- Secret shells (values written by downstream layers or manually):
  - `home-cloudflared-tunnel-token` — cloudflared VM tunnel token
  - `home-cloudflare-api-token` — Cloudflare API token for External Secrets Operator
- `home-cloudflared` GCP SA for the cloudflared VM, with `secretAccessor` on its tunnel token secret
- `home-external-secrets` GCP SA for ESO, with:
  - `secretAccessor` on the `cloudflare-api-token` secret
  - Workload Identity binding for k8s SA `external-secrets/external-secrets`

## Adding new secrets

Add entries to the `secrets` local map in `main.tf`. Each entry creates a secret shell named `${project_name}-${key}`.

## Dependencies

- None (standalone layer)

## Required variables

| Variable | Source |
|---|---|
| `gcp_region` | `TF_VAR_gcp_region` |
| `gcp_project` | `TF_VAR_gcp_project` |

See [terraform/README.md](../README.md) for the full deployment order.
