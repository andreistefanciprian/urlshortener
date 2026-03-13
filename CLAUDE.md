# Claude Agent Context

## Repo Overview
URL Shortener app + GitOps fluxcd manifests. Two repos:
- **This repo** (`urlshortener`): Terraform GCP infra + Helm charts + app source
- **GitOps repo**: https://github.com/andreistefanciprian/urlshortener-gitops (FluxCD manifests)

## Architecture
- Private GKE cluster (no public endpoint) — access via Cloudflare WARP client enrolled in Zero Trust
- cloudflared runs on a GCE VM (not in k8s), inside the GCP VPC, tunnels traffic from Cloudflare to the cluster
- DNS: Cloudflare public CNAME records point to the tunnel (`<tunnel-id>.cfargotunnel.com`); no split horizon or GCP Cloud DNS involved
- External Secrets Operator (ESO) in k8s reads secrets from GCP Secret Manager via Workload Identity
- cert-manager issues internal TLS certs via Google CAS (`certificate_authority` layer) or LetsEncrypt
- GitOps via FluxCD from urlshortener-gitops [repo](https://github.com/andreistefanciprian/urlshortener-gitops)

## Repo Structure
```
api-gateway/            # REST API server — routes HTTP to gRPC services
frontend/               # Web UI — Bootstrap admin interface (create/delete URLs)
url-gen/                # gRPC service — creates and deletes short URLs (PostgreSQL + Redis)
url-read/               # gRPC service — resolves short codes to long URLs (Redis cache-first)
  <component>/infra/    # Each component has infra/ with Dockerfile and Helm chart

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
  kyverno/            # Kyverno policy for enforcing signed images at runtime
  argocd/              # argocd manifest to deploy urlshorneter
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
