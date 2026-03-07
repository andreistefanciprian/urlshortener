# Cloudflared VM

Deploys a GCP VM running cloudflared as a tunnel connector in the same VPC as the GKE cluster.

## What it creates

- A GCP Compute instance (`home-cloudflared`) running Ubuntu 22.04 LTS
  - e2-small, internal only (no public IP), egress via Cloud NAT
  - Startup script installs cloudflared, fetches tunnel token from Secret Manager, and runs as a systemd service
- A VPC egress firewall rule allowing TCP 443 and UDP 7844 for `cloudflared-tunnel` tagged instances

## Dependencies

- `terraform/networking/` — reads VPC and subnet from remote state
- `terraform/secrets/` — reads cloudflared service account email and secret ID from remote state
- `terraform/cloudflare/` — must be applied first so the tunnel token exists in Secret Manager

## Required variables

| Variable | Source |
|---|---|
| `gcp_region` | `TF_VAR_gcp_region` |
| `gcp_project` | `TF_VAR_gcp_project` |
| `gcp_zone` | `TF_VAR_gcp_zone` |
| `tfstate_bucket` | `TF_VAR_tfstate_bucket` |

Add `TF_VAR_gcp_zone` to your `.env` file (e.g., `TF_VAR_gcp_zone=us-central1-a`).

See [terraform/README.md](../README.md) for the full deployment order.
