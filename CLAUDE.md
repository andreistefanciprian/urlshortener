# Claude Agent Context

## Repo Overview
URL Shortener app + GCP infrastructure. Two repos:
- **This repo** (`urlshortener`): Terraform infra + Helm charts + app source
- **GitOps repo**: https://github.com/andreistefanciprian/urlshortener-gitops (FluxCD manifests)

## Architecture
- Private GKE cluster (no public endpoint) — access via Cloudflare WARP client enrolled in Zero Trust
- cloudflared runs on a GCE VM (not in k8s), inside the GCP VPC, tunnels traffic from Cloudflare to the cluster
- DNS: Cloudflare public CNAME records point to the tunnel (`<tunnel-id>.cfargotunnel.com`); no split horizon or GCP Cloud DNS involved
- External Secrets Operator (ESO) in k8s reads secrets from GCP Secret Manager via Workload Identity
- cert-manager issues internal TLS certs via Google CAS (`certificate_authority` layer)
- GitOps via FluxCD from urlshortener-gitops repo

## Repo Structure
```
terraform/
  tf_bucket/          # GCS state bucket — bootstrapped via docker compose, NOT make
  networking/         # VPC, subnets, Cloud NAT, firewall rules
  gke/                # GKE cluster, node pool, cluster SA
  secrets/            # Secret Manager secrets + IAM for cloudflared & ESO SAs
  certificate_authority/  # Google CAS for internal TLS
  cloudflare/         # Cloudflare Tunnel + Zero Trust access policies
  cloudflared_vm/     # GCE VM running cloudflared connector
  setup.sh            # One-time GCP bootstrap (APIs, terraform-admin SA, roles)
  makefile            # Terraform workflow (plan/deploy/destroy via docker compose)
  docker-compose.yaml # Runs terraform in container with host ADC credentials
  .env                # TF_VAR_* and GOOGLE_IMPERSONATE_SERVICE_ACCOUNT (gitignored)
  .env.example        # Template for .env
  secrets.sh          # CLI script to populate Secret Manager values after tf apply

k8s/
  gateway-api/        # Helm chart: GatewayClass, Gateway, HTTPRoute, ReferenceGrant
```

## Terraform Layer Deploy Order
1. `tf_bucket` — `docker compose run terraform -chdir=tf_bucket apply -auto-approve`
2. `networking`
3. `gke` — must come before secrets; creates the Workload Identity pool (`<project>.svc.id.goog`) that the ESO SA binding depends on
4. `secrets` — creates secret shells + IAM; populate values afterwards with `secrets.sh`
5. `certificate_authority`
6. `cloudflare`
7. `cloudflared_vm`

Standard commands: `make plan TF_TARGET=<layer>` / `make deploy-auto-approve TF_TARGET=<layer>`

## GCP Auth Model
- Personal GCP account (not an org). SA key creation blocked by org policy (`constraints/iam.disableServiceAccountKeyCreation`).
- Keyless auth: `gcloud auth application-default login` on laptop → host `~/.config/gcloud` mounted into Docker container → `GOOGLE_IMPERSONATE_SERVICE_ACCOUNT` env var causes Terraform provider to impersonate `terraform-admin` SA. No SA key ever downloaded.
- `setup.sh` creates `terraform-admin` SA and grants all required roles. Run once per project.

## terraform-admin SA Roles
- `compute.networkAdmin` — VPC, subnets, NAT, routes (does NOT cover firewalls)
- `compute.securityAdmin` — firewall rules
- `compute.instanceAdmin.v1` — VM instances
- `container.admin` — GKE cluster and node pools
- `secretmanager.admin` — Secret Manager
- `privateca.admin` — Certificate Authority
- `storage.admin` — GCS buckets
- `iam.serviceAccountAdmin`, `iam.serviceAccountKeyAdmin`, `iam.serviceAccountUser`
- `resourcemanager.projectIamAdmin` — grant IAM roles to created SAs
- `serviceusage.serviceUsageAdmin` — enable GCP APIs

## Key Secrets
All secrets are prefixed with `${project_name}-` (default: `home`):
- `home-cloudflared-tunnel-token` — cloudflared VM only
- `home-cloudflare-api-token` — ESO
- `home-redis-password`, `home-db-admin-password`, `home-db-user-password`, `home-db-replication-password` — ESO

Secrets are shell-only after `tf apply`. Run `terraform/secrets.sh` to populate with random passwords.

## GKE Cluster
- Private: `enable_private_endpoint=true`, `enable_private_nodes=true`, master CIDR `10.100.48.0/28`
- VPC native, Workload Identity, Calico network policy, Shielded nodes
- Node type: `n1-standard-2`, spot, `pd-standard` 10GB
- Network: nodes `10.100.0.0/24`, pods `10.100.16.0/20`, services `10.100.32.0/24`
- `use_default_node_pool` var (bool, default `false`): when `true` uses default pool (no autoscaling); when `false` uses separately managed pool with autoscaling 1–2 nodes

## Secrets Layer — Ordering
`secrets` must be deployed **after** `gke`. The ESO SA Workload Identity binding references `${project}.svc.id.goog` pool, which is created by the GKE cluster. Destroy order is the reverse: `secrets` before `gke`.

## Cloudflare Provider (v5)
- Using `~> 5.0`. Key v5 changes vs v4:
  - Policies are inline inside `cloudflare_zero_trust_access_application` via `policies = [...]`
  - No `application_id`/`zone_id`/`precedence` on standalone policy resources
  - `warp = {}` rule type removed — use `cloudflare_zero_trust_device_posture_rule` type `"warp"` + `require = [{ device_posture = { integration_uid = ... } }]`
- `cloudflare_tunnel_config` ingress rules not needed — CNAME + private network route handles routing
- Device enrollment cannot be terraformed (no resource in provider)
- `cloudflare_record` `name` is relative to zone: `"echo"` → `echo.9tzy.xyz`
- Access policies scoped by `account_id` only (no `zone_id`)

## Gateway API Helm Chart (`k8s/gateway-api/`)
- GatewayClass creation is optional (cluster may already have one)
- Gateway addresses block is optional (private ingress without static IP)
- cert-manager: uses `cert-manager.io/cluster-issuer: letsencrypt-prod` annotation on Gateway; `letsencrypt-prod` ClusterIssuer is pre-existing in cluster
- External-DNS running in cluster, points at Cloudflare zone `9tzy.xyz`
