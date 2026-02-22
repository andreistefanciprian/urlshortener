# Cloudflare

Creates a Cloudflare Tunnel and writes the tunnel token into GCP Secret Manager.

## What it creates

- A Cloudflare Tunnel (`home-gke-tunnel`)
- A tunnel route for `10.100.0.0/18` (covers nodes, pods, services, and GKE master)
- A Secret Manager secret version containing the tunnel token

## Dependencies

- `terraform/secrets/` — reads `cloudflared_secret_name` from remote state to write the tunnel token

## Required variables

| Variable | Source |
|---|---|
| `gcp_region` | `TF_VAR_gcp_region` |
| `gcp_project` | `TF_VAR_gcp_project` |
| `tfstate_bucket` | `TF_VAR_tfstate_bucket` |
| `cloudflare_account_id` | `TF_VAR_cloudflare_account_id` (sensitive) |
| `cloudflare_api_token` | `TF_VAR_cloudflare_api_token` (sensitive) |

Add `TF_VAR_cloudflare_account_id` and `TF_VAR_cloudflare_api_token` to your `.env` file.

## Apply order

```
1. terraform/networking/
2. terraform/secrets/
3. terraform/gke/
4. terraform/certificate_authority/
5. terraform/cloudflare/         ← this layer
6. terraform/cloudflared-vm/
```
