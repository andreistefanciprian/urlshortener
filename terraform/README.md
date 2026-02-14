# URL Shortener - Infrastructure with Terraform

This directory provisions the core GCP infrastructure for the [urlshortener](https://github.com/andreistefanciprian/urlshortener) application using Terraform.

> **Note:** This is a demo/development setup. Each Terraform module contains a README with production recommendations.

## Architecture

The infrastructure is split into independent Terraform modules:

| Module | Description | Status |
|---|---|---|
| `tf_bucket` | GCS bucket for Terraform remote state | Ready |
| `networking` | VPC, subnets, Cloud NAT, firewall rules | Ready |
| `gke` | GKE cluster, node pool, service accounts | Ready |
| `certificate_authority` | Google CAS for TLS certificates via cert-manager | Ready |

Once infrastructure is deployed, Kubernetes applications are deployed via FluxCD from the `flux/` folder in this repo.

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

After bucket creation, get the bucket name from the terraform output:

```bash
docker compose run terraform -chdir=tf_bucket output -raw bucket_name
```

Then update `TF_VAR_tfstate_bucket` in your `.env` file with this value.

### Step 2: Deploy Networking

```bash
make plan TF_TARGET=networking
make deploy-auto-approve TF_TARGET=networking
```

### Step 3: Deploy GKE Cluster

```bash
make plan TF_TARGET=gke
make deploy-auto-approve TF_TARGET=gke

# Configure kubectl
gcloud container clusters get-credentials home --region $GCP_REGION --project $GCP_PROJECT
kubectl cluster-info
```

### Step 4: Deploy Certificate Authority

```bash
make plan TF_TARGET=certificate_authority
make deploy-auto-approve TF_TARGET=certificate_authority
```

## Next Steps: Deploy Applications

Once infrastructure is deployed, use FluxCD to deploy Kubernetes applications from the `flux/` folder in this repo.

## Cleanup

### Destroy All Resources

```bash
# Destroy in reverse order
make destroy-auto-approve TF_TARGET=certificate_authority
make destroy-auto-approve TF_TARGET=gke
make destroy-auto-approve TF_TARGET=networking

# Destroy state bucket
docker compose run terraform -chdir=tf_bucket destroy -auto-approve
```

### Clean Local Terraform Files

```bash
make clean TF_TARGET=tf_bucket
make clean TF_TARGET=networking
make clean TF_TARGET=gke
make clean TF_TARGET=certificate_authority
```
