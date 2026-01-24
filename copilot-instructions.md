# Copilot Instructions

This repository implements a microservices architecture in Go, with the following services:

## Services

### Backend Services
- **api-gateway**: REST API and gRPC client
  - Location: `./api-gateway`
  - Dockerfile and Helm chart: `./api-gateway/infra`

- **url-gen**: gRPC server (handles short URL generation logic)
  - Location: `./url-gen`
  - Dockerfile and Helm chart: `./url-gen/infra`

- **url-read**: gRPC server (handles short URL reader logic)
  - Location: `./url-read`
  - Dockerfile and Helm chart: `./url-read/infra`

### Frontend Services
- **frontend**: Web UI for URL shortener
  - Location: `./frontend`
  - Dockerfile and Helm chart: `./frontend/infra`

### Infrastructure Services
- **redis**: Cache layer for fast URL lookups and reduced database load

- **postgresql**: Primary database for persistent URL storage and metadata

## Kubernetes Manifests
- ArgoCD manifests: `./k8s/argocd`
- Gateway API manifests: `./k8s/gateway-api`

## General Guidelines
- Do not write complete functions or snippets unless I specifically ask for them.
- When writing code for me, help me understand it rather than just giving me the answer.
- My purpose is to get better at coding not just to get the code and follow your instructions blindly.
- Follow idiomatic Go practices and conventions.
- All services use context propagation.

## Architecture Decisions

- LongURL expiration time validation is performed upstream (at the API Gateway level) rather than at the database or cache layer.
