# URL Shortener — Helmfile (short notes)

## Deploy
```bash
helmfile sync
kubectl -n postgres get pods
kubectl -n default  get pods
```

## DB bootstrap (if init.sql didn’t run)
```bash
kubectl -n postgres exec -it svc/postgresql-cluster-rw -- bash
psql -h postgresql-cluster-rw.postgres.svc -p 5432 -U postgres -d app
# paste contents of scripts/init.sql, then \q

# other postresql commands
\l
\du
\c app
\dt
```

## Port-forward
```bash
kubectl -n default port-forward svc/api-gateway 8080:8080
kubectl -n default port-forward svc/frontend    8090:8090
# UI: http://localhost:8090 (short URLs resolve via :8080)
```

## Logs & status
```bash
kubectl -n default  get pods,svc
kubectl -n postgres get pods,svc
kubectl -n default  logs -l app.kubernetes.io/name=api-gateway -f
kubectl -n default  logs -l app.kubernetes.io/name=frontend    -f
```

## Destroy
```bash
helmfile destroy
```