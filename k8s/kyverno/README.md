# Kyverno Image Verification Policies

This directory contains Kyverno policy to enforce image signature and SBOM attestation verification for the URL Shortener microservices.

## Policies

### verify-images-policy.yaml
Verifies that all container images from `ghcr.io/andreistefanciprian` are:
- **Signed** with cosign using keyless signing (GitHub Actions OIDC)
- **Attested** with SBOM (SPDX format) using keyless attestation

**Enforcement**: 
- Blocks unsigned images from being deployed

**Namespaces covered**:
- `api-gateway`
- `frontend`
- `url-gen`
- `url-read`

## Apply Policies

```bash
# Install Kyverno
helm install kyverno kyverno/kyverno -n kyverno --create-namespace --version=3.7.0 --set features.logging.verbosity=4

# Provide GHCR credentials to kyverno
kubectl -n kyverno create secret docker-registry ghcr-credentials \
  --docker-server=ghcr.io \
  --docker-username=andreistefanciprian \
  --docker-password=$GITHUB_TOKEN \
  --docker-email='andreistefanciprian@gmail.com'

# Allow kyverno SAs to read secrets in cluster
kubectl apply -f k8s/kyverno/rbac.yaml

# Apply the policy
kubectl apply -f k8s/kyverno/imagevalidatingpolicy.yaml

# Verify policy is active
kubectl get imagevalidatingpolicy
```

## Testing

```bash
# Try deploying unsigned image (should fail)
kubectl run test --image=ghcr.io/andreistefanciprian/api-gateway:<UNSIGNED_SHA256_DIGEST> -n api-gateway

# Deploy signed image (should succeed)
kubectl run test --image=ghcr.io/andreistefanciprian/api-gateway:<SIGNED_SHA256_DIGEST> -n api-gateway

# Check policy reports and events
kubectl get policyreport -A
kubectl describe imagevalidatingpolicy verify-urlshortener-images | grep Events -A 10

# Check Kyverno logs
kubectl logs -n kyverno -l app.kubernetes.io/name=kyverno -f
```

## Check image via cosign cli

```bash
# Verify signature with cosign cli
export COSIGN_REPOSITORY=ghcr.io/andreistefanciprian/cosign-signatures
cosign triangulate ghcr.io/andreistefanciprian/api-gateway@sha256:<SIGNED_SHA256_DIGEST>
cosign verify \
--certificate-identity-regexp "https://github.com/andreistefanciprian/.+" \
--certificate-oidc-issuer https://token.actions.githubusercontent.com \
ghcr.io/andreistefanciprian/api-gateway:latest@sha256:<SIGNED_SHA256_DIGEST>

# Verify SBOM attestation
cosign verify-attestation --type spdxjson \
--certificate-identity-regexp "https://github.com/andreistefanciprian/.+" \
--certificate-oidc-issuer https://token.actions.githubusercontent.com \
ghcr.io/andreistefanciprian/api-gateway@sha256:<SIGNED_SHA256_DIGEST>
```
