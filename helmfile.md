# URL Shortener - Helmfile Deployment

## Setup

### 1. Create Kind Cluster
```bash
kind create cluster --name urlshortener --config - <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
EOF

kubectl cluster-info && kubectl get nodes
```

### 2. Configure Local DNS
```bash
# Add to /etc/hosts for l.it domain resolution
echo "127.0.0.1 l.it" | sudo tee -a /etc/hosts
```

### 3. Deploy Services
```bash
# Create secrets
bash helmfile-secrets.sh

# Deploy all
helmfile sync && kubectl -A get pods

# Deploy single service
helmfile -l name=api-gateway apply

# Check status
helmfile list
helmfile -l name=api-gateway status
```

## Operations

### Port Forward
```bash
sudo kubectl -n api-gateway port-forward svc/api-gateway 80:8080
kubectl -n frontend port-forward svc/frontend 8090:8090
# Access: http://localhost:8090
```

### Restart Service
```bash
# Via Helmfile
helmfile -l name=api-gateway destroy && helmfile -l name=api-gateway apply

# Via kubectl
kubectl rollout restart deployment api-gateway -n api-gateway
```

### Database Access
```bash
# Quick query
kubectl run pg-client --rm -it --image=postgres:18-alpine --env PGPASSWORD=postgres -- \
  psql -h pg-postgresql.postgres.svc -U postgres -d urls -c "SELECT * FROM short_links;"

# Interactive session
kubectl run pg-client --rm -it --image=postgres:18-alpine --env PGPASSWORD=postgres -- bash
# Then: psql -h pg-postgresql.postgres.svc -U postgres -d urls
```

### Logs
```bash
kubectl -A get pods,svc
kubectl -n api-gateway logs -l app.kubernetes.io/name=api-gateway -f
kubectl -n frontend logs -l app.kubernetes.io/name=frontend -f
kubectl -n url-gen logs -l app.kubernetes.io/name=url-gen -f
kubectl -n url-read logs -l app.kubernetes.io/name=url-read -f
```

## Cleanup
```bash
helmfile destroy
kind delete cluster --name urlshortener
```