
## Prerequisites

- Gateway API CRDs installed (v1). Many controllers install these automatically; if not, install explicitly:
	```bash
	kubectl apply -k "github.com/kubernetes-sigs/gateway-api/config/crd?ref=v1.1.0"
	```
- A Gateway API controller (ingress) installed, such as Envoy Gateway or Istio:
	- Envoy Gateway (example):
		```bash
		helm upgrade eg oci://docker.io/envoyproxy/gateway-helm --version v1.2.4 -n envoy-gateway-system
		```
	- Istio (example): install Istio and its gateway component; ensure the Gateway controller is running.
- cert-manager installed in the cluster (provides TLS certificates consumed by the Gateway listeners).

## Quick start

1) Install via Helm:
```
helm install gw ./k8s/gateway-api/helm/gateway-api -n default
```

2) Verify external address:
```
# List gateways and find your resource name
kubectl get gateway -n default

# Or query a specific Gateway (replace the name if different)
kubectl get gateway urlshortener-gateway -n default -o yaml | yq '.status.addresses'
```

3) Optional: use a reserved static IP (GKE):
```
helm upgrade gw ./k8s/gateway-api/helm/gateway-api \
	--set gateway.addresses.enabled=true \
	--set gateway.addresses.value=YOUR_IP \
	-n default
```

4) Optional: update Route 53 DNS:
```
# Example: replace 1.2.3.4 with your Gateway's public IP
NEW_IP=<1.2.3.4> bash k8s/gateway-api/update-route53.sh
```
- Requires AWS CLI configured; set `HOSTED_ZONE_ID`, `ROOT_DOMAIN`, and `APP_DOMAIN` environment variables.