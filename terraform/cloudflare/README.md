# Cloudflare

Creates a Cloudflare Tunnel and writes the tunnel token into GCP Secret Manager.

## What it creates

- A Cloudflare Tunnel (`home-gke-tunnel`)
- A tunnel route for `10.100.0.0/18` (covers nodes, pods, services, and GKE master)
- A WARP device default profile: WARP mode (traffic + DNS proxy), split tunnel include `10.100.0.0/18`
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

## Device Profile vs Device Enrollment

This layer provisions `cloudflare_zero_trust_device_default_profile` which configures:
- **WARP mode** (proxies both traffic and DNS)
- **Split tunnel include**: `10.100.0.0/18` (GKE nodes, pods, services, master)

**Device enrollment** (the step where a user logs into the WARP client with a team name and receives a one-time code via email) cannot be terraformed — there is no Cloudflare provider resource for it. Users must enroll manually after this layer is applied.

See [terraform/README.md](../README.md) for the full deployment order.
