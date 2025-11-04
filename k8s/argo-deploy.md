## Argo CD deploy notes (short)

### Setup
- Install Argo CD: https://argo-cd.readthedocs.io/en/stable/getting_started/
- Port-forward UI (optional):
```bash
kubectl -n argocd port-forward svc/argocd-server 8080:80
```
- CLI login (example):
```bash
argocd login localhost:8080 --username admin --password <pwd> --insecure
```

### Deploy
```bash
kubectl apply -n argocd -f k8s/argocd/project-urlshortener.yaml
kubectl apply -n argocd -f k8s/argocd/root-app.yaml
argocd app sync urlshortener-root
argocd app wait urlshortener-root --health --timeout 600
```

### Secrets (before first sync)
```bash
# Postgres
kubectl create ns postgres --dry-run=client -o yaml | kubectl apply -f -
kubectl -n postgres create secret generic pg-creds \
	--from-literal=admin-password='postgres' \
	--from-literal=user-password='Auth123' \
	--from-literal=replication-password='postgres' \
	--dry-run=client -o yaml | kubectl apply -f -

# Redis
kubectl create ns redis --dry-run=client -o yaml | kubectl apply -f -
kubectl -n redis create secret generic redis-creds \
	--from-literal=redis-password='hTzvklFVp7' \
	--dry-run=client -o yaml | kubectl apply -f -

# url-read
kubectl create ns url-read --dry-run=client -o yaml | kubectl apply -f -
kubectl -n url-read create secret generic redis-creds \
	--from-literal=redis-password='hTzvklFVp7' \
	--dry-run=client -o yaml | kubectl apply -f -
kubectl -n url-read create secret generic pg-creds \
	--from-literal=user-password='Auth123' \
	--dry-run=client -o yaml | kubectl apply -f -

# url-gen
kubectl create ns url-gen --dry-run=client -o yaml | kubectl apply -f -
kubectl -n url-gen create secret generic redis-creds \
	--from-literal=redis-password='hTzvklFVp7' \
	--dry-run=client -o yaml | kubectl apply -f -
kubectl -n url-gen create secret generic pg-creds \
	--from-literal=user-password='Auth123' \
	--dry-run=client -o yaml | kubectl apply -f -
```

### Delete all
```bash
argocd app delete urlshortener-root --cascade --yes
kubectl delete ns api-gateway url-gen url-read frontend redis postgres || true
```

### Tips
- Force reconcile:
```bash
kubectl annotate application urlshortener-root -n argocd argocd.argoproj.io/refresh=hard --overwrite
```
- Private charts: add Argo CD repo credentials Secret matching the OCI URL.