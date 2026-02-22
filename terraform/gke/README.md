# GKE Terraform - Demo Cluster

This Terraform code provisions a GKE cluster intended for **demo and development purposes**. Several configuration choices prioritise cost savings and convenience over production-grade resilience and security. The sections below outline what would need to change for a production deployment.

## What would need to change for production

### Security

| Item | Current (demo) | Production recommendation |
|---|---|---|
| SSH firewall rule (`network.tf`) | Open to `0.0.0.0/0` on port 22 | Remove entirely, or restrict `source_ranges` to a bastion/VPN CIDR |
| Deletion protection (`main.tf`) | `false` | Set to `true` to prevent accidental cluster destruction |

### Availability and resilience

| Item | Current (demo) | Production recommendation |
|---|---|---|
| Node pool VM type (`main.tf`) | `spot = true` (can be preempted at any time) | Use on-demand instances, or a mix of on-demand + spot node pools |
| Node disk size (`variables.tf`) | `10 GB` | Increase to at least `100 GB` — container images alone can fill 10 GB |
| Autoscaling range (`main.tf`) | `1-2` nodes | Increase `max_node_count` based on expected workload, consider multiple node pools across zones |
| Cluster location | Single region | Consider a multi-zonal or regional cluster for higher availability |

### Networking

| Item | Current (demo) | Production recommendation |
|---|---|---|
| Master authorised networks | Not configured | Restrict API server access to known CIDRs using `master_authorized_networks_config` |

## Apply order

```
1. terraform/networking/
2. terraform/secrets/
3. terraform/gke/               ← this layer
4. terraform/certificate_authority/
5. terraform/cloudflare/
6. terraform/cloudflared-vm/
```
