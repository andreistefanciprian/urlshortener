# Certificate Authority - Demo Setup

This Terraform code provisions a Google Cloud Certificate Authority Service (CAS) infrastructure intended for **demo and development purposes**. It is used by [cert-manager Google CAS Issuer](https://github.com/andreistefanciprian/flux-demo/tree/main/infra/cert-manager) to issue TLS certificates for Kubernetes services via Workload Identity.

Key usage is scoped to TLS server certificate issuance only (`digital_signature`, `key_encipherment`, `cert_sign`, `crl_sign`, `server_auth`).

## Resources created

- Private CA API enablement
- CA Pool (Enterprise tier)
- Certificate Authority (root, EC P384)
- CAS Issuer service account with Workload Identity binding
- CA Pool IAM binding for certificate requesting

## What would need to change for production

### Security

| Item | Current (demo) | Production recommendation |
|---|---|---|
| Deletion protection (`main.tf`) | `false` | Set to `true` to prevent accidental CA destruction |
| `ignore_active_certificates_on_deletion` | `true` | Set to `false` — ensures a CA cannot be deleted while certificates are still in use |
| `disable_on_destroy` (API service) | `true` (API disabled on destroy) | Set to `false` — prevents disabling the Private CA API when other resources may depend on it |
| CA hierarchy | Single root CA issuing directly | Use a root CA + subordinate CA hierarchy. The root CA stays offline/disabled, subordinate CA issues day-to-day certificates |
| `max_issuer_path_length` | `10` | Reduce to `1` (or `0` if no subordinate CAs) to limit chain depth |
| IAM binding type | `iam_binding` (authoritative) | Consider `iam_member` (additive) to avoid overwriting manually added members in shared environments |
| `client_auth` | `false` | Set to `true` if mTLS between services is required |

### Availability and lifecycle

| Item | Current (demo) | Production recommendation |
|---|---|---|
| CA lifetime | `7776000s` (90 days) | Increase to 5-10 years for a root CA (e.g. `315360000s` for 10 years) |
| CA Pool tier | `ENTERPRISE` | Keep `ENTERPRISE` for audit logging and finer controls, or use `DEVOPS` if cost is a concern and audit logs are not required |
| Region | Single region | Consider the CA's availability requirements — CAS is a global service but the CA Pool is regional |

## Apply order

```
1. terraform/networking/
2. terraform/secrets/
3. terraform/gke/
4. terraform/certificate_authority/ ← this layer
5. terraform/cloudflare/
6. terraform/cloudflared-vm/
```
