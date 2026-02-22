# Secrets

Central secrets management layer. Creates Secret Manager secret shells and IAM bindings. Secret values are written by downstream layers.

## What it creates

- Secret Manager API enablement
- A `google_secret_manager_secret` shell for the cloudflared tunnel token
- A dedicated GCP service account (`home-cloudflared`) for the cloudflared VM
- IAM binding granting the service account `roles/secretmanager.secretAccessor` scoped to the cloudflared secret only

## Adding new secrets

Add entries to the `secrets` local map in `main.tf`. Each entry creates a secret shell named `${project_name}-${key}`.

## Dependencies

- None (standalone layer)

## Required variables

| Variable | Source |
|---|---|
| `gcp_region` | `TF_VAR_gcp_region` |
| `gcp_project` | `TF_VAR_gcp_project` |

## Apply order

```
1. terraform/networking/
2. terraform/secrets/            ← this layer
3. terraform/gke/
4. terraform/certificate_authority/
5. terraform/cloudflare/
6. terraform/cloudflared-vm/
```
