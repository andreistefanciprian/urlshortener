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