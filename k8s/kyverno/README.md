# Kyverno Image Verification Policies

This directory contains Kyverno policies to enforce image signature verification for the URL Shortener microservices.

## Policies

### verify-images-policy.yaml
Verifies that all container images from `ghcr.io/andreistefanciprian` are signed with cosign using keyless signing (GitHub Actions OIDC).

**Enforcement**: 
- Blocks unsigned images from being deployed
- Validates signatures against Sigstore/Rekor transparency log
- Verifies GitHub Actions as the trusted issuer

**Namespaces covered**:
- `api-gateway`
- `frontend`
- `url-gen`
- `url-read`

## Apply Policies

```bash
# Install Kyverno
helm install kyverno kyverno/kyverno -n kyverno --create-namespace

# Provide GHCR credentials to kyverno
kubectl -n kyverno create secret docker-registry ghcr-credentials \
  --docker-server=ghcr.io \
  --docker-username=andreistefanciprian \
  --docker-password=<GITHUB_PAT_WITH_PACKAGES_WRITE_PERMISSION> \
  --docker-email='andreistefanciprian@gmail.com'

# Apply the policy
kubectl apply -f k8s/kyverno/verify-images-policy.yaml

# Verify policy is active
kubectl get clusterpolicy verify-urlshortener-images

# Check policy reports
kubectl get policyreport -A
```

## Testing

```bash
# Try deploying unsigned image (should fail)
kubectl run test --image=ghcr.io/andreistefanciprian/api-gateway:<SHA256_DIGEST> -n api-gateway

# Deploy signed image (should succeed)
kubectl run test --image=ghcr.io/andreistefanciprian/api-gateway:<SHA256_DIGEST> -n api-gateway
# or
helmfile -l name=api-gateway apply
```

## Troubleshooting

```bash
# Check policy status
kubectl describe clusterpolicy verify-urlshortener-images

# View admission events
kubectl get events -A --sort-by='.lastTimestamp' | grep -i kyverno

# Check Kyverno logs
kubectl logs -n kyverno -l app.kubernetes.io/name=kyverno -f

# Verify with cosign cli
export COSIGN_REPOSITORY=ghcr.io/andreistefanciprian/cosign-signatures
cosign triangulate ghcr.io/andreistefanciprian/api-gateway@sha256:<SHA256_DIGEST>
cosign verify \
--certificate-identity-regexp "https://github.com/andreistefanciprian/.+" \
--certificate-oidc-issuer https://token.actions.githubusercontent.com \
ghcr.io/andreistefanciprian/api-gateway:latest@sha256:<SHA256_DIGEST>
```

## Notes

- Uses **keyless signing** with GitHub Actions OIDC tokens
- Validates against **Sigstore Rekor** transparency log
- **No keys required** - verification based on OIDC identity
- Failure policy is set to **Fail** - unsigned images are rejected
