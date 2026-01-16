# GitHub Actions Workflows

This directory contains CI/CD workflows for building, scanning, signing, and pushing container images to GitHub Container Registry (GHCR).

## Available Workflows

- `ghcr-api-gateway.yaml` - API Gateway service
- `ghcr-frontend.yaml` - Frontend service
- `ghcr-url-gen.yaml` - URL Generation service
- `ghcr-url-read.yaml` - URL Read service

## Image Security Features

All images are:
- Scanned with Trivy for vulnerabilities (CRITICAL, HIGH)
- Signed with Cosign (keyless signing via GitHub OIDC)
- Attested with SBOM (Software Bill of Materials)
- Built with provenance attestations

## Verify Image Signatures

```bash
# Get digest for a specific tag
docker pull ghcr.io/andreistefanciprian/api-gateway:latest
docker inspect ghcr.io/andreistefanciprian/api-gateway:latest --format='{{.RepoDigests}}'

# Display the signature and attestation tree
cosign tree \
  ghcr.io/andreistefanciprian/api-gateway@sha256:b530f8c4d287d6fc52aabf940cf86cc1c7f0c637d986d76d354b26e503638397

# Verify the image was signed by GitHub Actions from this repository
cosign verify \
  --certificate-identity-regexp "https://github.com/andreistefanciprian/.+" \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  ghcr.io/andreistefanciprian/api-gateway@sha256:b530f8c4d287d6fc52aabf940cf86cc1c7f0c637d986d76d354b26e503638397

# Verify the SBOM attestation is signed
cosign verify-attestation \
  --type spdxjson \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp "https://github.com/andreistefanciprian/.+" \
  ghcr.io/andreistefanciprian/api-gateway@sha256:b530f8c4d287d6fc52aabf940cf86cc1c7f0c637d986d76d354b26e503638397

# Extract the SBOM to a file
cosign verify-attestation \
  --type spdxjson \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp "https://github.com/andreistefanciprian/.+" \
  ghcr.io/andreistefanciprian/api-gateway@sha256:b530f8c4d287d6fc52aabf940cf86cc1c7f0c637d986d76d354b26e503638397 \
  | jq -r '.payload' | base64 -d | jq '.predicate'
```