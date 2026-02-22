# URL Shortener - Infrastructure with Terraform

This directory provisions the core GCP infrastructure for the [urlshortener](https://github.com/andreistefanciprian/urlshortener) application using Terraform.

> **Note:** This is a demo/development setup. Each Terraform module contains a README with production recommendations.

## Architecture

The infrastructure is split into independent Terraform modules:

| Module | Description | Status |
|---|---|---|
| `tf_bucket` | GCS bucket for Terraform remote state | Ready |
| `networking` | VPC, subnets, Cloud NAT, firewall rules | Ready |
| `secrets` | Secret Manager secrets and IAM for cloudflared | Ready |
| `gke` | GKE cluster, node pool, service accounts | Ready |
| `certificate_authority` | Google CAS for internal TLS certificates issued by cert-manager | Ready |
| `cloudflare` | Cloudflare Tunnel and private network route | Ready |
| `cloudflared-vm` | VM running cloudflared tunnel connector | Ready |

Once infrastructure is deployed, Kubernetes applications are deployed via FluxCD from the [urlshortener-gitops](https://github.com/andreistefanciprian/urlshortener-gitops) repo.

### Network CIDR Allocation

All subnets fit under `10.100.0.0/18` for simple [Cloudflare Tunnel CIDR routing](https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/private-net/cloudflared/connect-cidr/). 

| Resource | CIDR | IPs |
|---|---|---|
| Nodes | `10.100.0.0/24` | 254 |
| Pods | `10.100.16.0/20` | 4,096 |
| Services | `10.100.32.0/24` | 256 |
| GKE Master | `10.100.48.0/28` | 16 |

## Prerequisites

* **[gcloud CLI](https://cloud.google.com/sdk/docs/install)** - For interacting with Google Cloud
* **[GCP Account](https://console.cloud.google.com/)** - Active project with billing enabled
* **[Docker Compose](https://docs.docker.com/compose/install/other/)** - Terraform runs in a container
* **[Make](https://formulae.brew.sh/formula/make)** - Build automation tool

> No local Terraform installation needed - everything runs in Docker containers.

## Getting Started

### 1. Initial GCP Setup

```bash
# Set your GCP project environment variables
export GCP_PROJECT=<yourGcpProjectNameGoesHere>
export GCP_EMAIL=<yourAccountNameGoesHere>@gmail.com
export GCP_REGION=<yourGcpRegionGoesHere>

# Authenticate gcloud CLI
gcloud auth login $GCP_EMAIL

# Run setup script
# Configures gcloud, enables required GCP APIs, creates a Terraform service account with admin roles, and generates a service account key
bash setup.sh
```

### 2. Configure Environment

```bash
cp env.example .env
```

Update `.env` with your GCP project ID and region. The `TF_VAR_tfstate_bucket` value will be updated after creating the state bucket.

### 3. Verify Terraform Version

```bash
make verify_version
```

## Deployment Steps

### Step 1: Create Terraform State Bucket

```bash
docker compose run terraform -chdir=tf_bucket init
docker compose run terraform -chdir=tf_bucket apply -auto-approve
```

After bucket creation, update `TF_VAR_tfstate_bucket` in your `.env` file with the bucket name from outputs above.

### Step 2: Deploy Networking

```bash
make plan TF_TARGET=networking
make deploy-auto-approve TF_TARGET=networking
```

### Step 3: Deploy Secrets

```bash
make plan TF_TARGET=secrets
make deploy-auto-approve TF_TARGET=secrets
```

### Step 4: Deploy GKE Cluster

```bash
make plan TF_TARGET=gke
make deploy-auto-approve TF_TARGET=gke
```

### Step 5: Deploy Certificate Authority

```bash
make plan TF_TARGET=certificate_authority
make deploy-auto-approve TF_TARGET=certificate_authority
```

### Step 6: Deploy Cloudflare Tunnel

Add `TF_VAR_cloudflare_account_id` and `TF_VAR_cloudflare_api_token` to your `.env` file.

```bash
make plan TF_TARGET=cloudflare
make deploy-auto-approve TF_TARGET=cloudflare
```

### Step 7: Deploy Cloudflared VM

Add `TF_VAR_gcp_zone` to your `.env` file.

```bash
make plan TF_TARGET=cloudflared-vm
make deploy-auto-approve TF_TARGET=cloudflared-vm
```

### Step 8: Configure kubectl

The GKE cluster has a private endpoint only. Before running `kubectl`, ensure:
1. **Cloudflare WARP client** installed on your machine
2. **Device enrollment profile** configured in Cloudflare One dashboard (Team & Resources): both Traffic and DNS proxy enable and split tuneling to include internal GCP CIDR

```bash
gcloud container clusters get-credentials home --region $GCP_REGION --project $GCP_PROJECT --internal-ip
kubectl cluster-info
```

## Next Steps: Deploy Applications

Once infrastructure is deployed, use FluxCD to deploy Kubernetes applications from the [urlshortener-gitops](https://github.com/andreistefanciprian/urlshortener-gitops) repo.

## Cleanup

### Destroy All Resources

```bash
# Destroy in reverse order
make destroy-auto-approve TF_TARGET=cloudflared-vm
make destroy-auto-approve TF_TARGET=cloudflare
make destroy-auto-approve TF_TARGET=certificate_authority
make destroy-auto-approve TF_TARGET=gke
make destroy-auto-approve TF_TARGET=secrets
make destroy-auto-approve TF_TARGET=networking

# Destroy state bucket
docker compose run terraform -chdir=tf_bucket destroy -auto-approve
```

### Clean Local Terraform Files

```bash
make clean TF_TARGET=networking
make clean TF_TARGET=secrets
make clean TF_TARGET=gke
make clean TF_TARGET=certificate_authority
make clean TF_TARGET=cloudflare
make clean TF_TARGET=cloudflared-vm
```
