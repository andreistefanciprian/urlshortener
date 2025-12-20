
## Expose frontend and API gateway

- Install via Helm:
```
helm install gw ./k8s/gateway-api/helm/gateway-api -n default
```

- Optional: use a reserved static IP (GKE):
```
helm upgrade gw ./k8s/gateway-api/helm/gateway-api \
	--set gateway.addresses.enabled=true \
	--set gateway.addresses.value=YOUR_IP \
	-n default
```

- Verify external address:
```
kubectl get gateway urlshortener-gateway -n default -o yaml | yq '.status.addresses'
```

- Optional: update Route 53 DNS:
```
# Get public IP address of the Gateway
kubectl get gateway

# Example: replace 1.2.3.4 with your Gateway's public IP
NEW_IP=<1.2.3.4> bash k8s/gateway-api/update-route53.sh
```
- Requires AWS CLI configured; set `HOSTED_ZONE_ID`, `ROOT_DOMAIN`, and `APP_DOMAIN` environment variables.

## Prerequisites

- cert-manager installed in the cluster (provides TLS certificates consumed by the Gateway listeners).
- A Gateway API controller (ingress) installed, such as Envoy Gateway or Istio:
	- Envoy Gateway (example):
		```bash
		helm upgrade eg oci://docker.io/envoyproxy/gateway-helm --version v1.2.4 -n envoy-gateway-system
		```
	- Istio (example): install Istio and its gateway component; ensure the Gateway controller is running.