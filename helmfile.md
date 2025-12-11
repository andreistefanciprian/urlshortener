# URL Shortener — Helmfile for manual deployment only (short notes)

## Deploy
```bash
# deploy db secrets prior to db creation
kubectl create namespace postgres --dry-run=client -o yaml | kubectl apply -f -

kubectl -n postgres create secret generic pg-creds \
--from-literal=admin-password='postgres' \
--from-literal=user-password='Auth123' \
--from-literal=replication-password='postgres' \
--dry-run=client -o yaml | kubectl apply -f -

kubectl -n postgres create -f- <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: db-init-sql-2
data:
  init.sql: |
    -- Connect to urls database
    \c urls;

    -- Create table
    CREATE TABLE IF NOT EXISTS short_links (
        code          VARCHAR(16) PRIMARY KEY,
        original_url  TEXT        NOT NULL,
        created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
        expires_at    TIMESTAMPTZ NULL
    );
EOF

kubectl create namespace url-gen --dry-run=client -o yaml | kubectl apply -f -
kubectl -n url-gen create secret generic redis-creds \
	--from-literal=redis-password='redispassword' \
	--dry-run=client -o yaml | kubectl apply -f -

kubectl -n url-gen create secret generic pg-creds \
--from-literal=user-password='Auth123' \
--dry-run=client -o yaml | kubectl apply -f -

kubectl create namespace url-read --dry-run=client -o yaml | kubectl apply -f -
kubectl -n url-read create secret generic redis-creds \
	--from-literal=redis-password='redispassword' \
	--dry-run=client -o yaml | kubectl apply -f -

kubectl create namespace url-read --dry-run=client -o yaml | kubectl apply -f -
kubectl -n url-read create secret generic pg-creds \
--from-literal=user-password='Auth123' \
--dry-run=client -o yaml | kubectl apply -f -

kubectl create namespace redis --dry-run=client -o yaml | kubectl apply -f -
kubectl -n redis create secret generic redis-creds \
	--from-literal=redis-password='redispassword' \
	--dry-run=client -o yaml | kubectl apply -f -

# deploy all services
helmfile sync
kubectl -n postgres get pods
kubectl -n default  get pods
```

## Db commands
```bash
# one command
kubectl run pg-client --namespace default --image=postgres:18-alpine --restart=Never --rm -it --env PGPASSWORD=postgres -- psql -h pg-postgresql.postgres.svc -p 5432 -U postgres -d urls -c "SELECT * FROM short_links;"

# or interactive prompt
kubectl run pg-client --namespace default --image=postgres:18-alpine --restart=Never --rm -it --env PGPASSWORD=postgres -- bash
psql -h pg-postgresql -p 5432 -U postgres
# other postgresql commands
\l
\du
\c urls
\dt
\q
exit

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
kubectl -n default logs -l app.kubernetes.io/name=api-gateway -f
kubectl -n default logs -l app.kubernetes.io/name=frontend -f
kubectl -n default logs -l app.kubernetes.io/name=url-read -f
kubectl -n default logs -l app.kubernetes.io/name=url-gen -f
```

## Destroy
```bash
helmfile destroy
```