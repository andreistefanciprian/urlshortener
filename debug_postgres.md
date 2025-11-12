```
helmfile sync


k -n postgres logs -l app.kubernetes.io/instance=cnpg-operator
k -n postgres logs -l cnpg.io/cluster=postgresql-cluster
k -n postgres logs -l job-name=postgresql-cluster-1-initdb


kubectl --namespace postgres exec --stdin --tty services/postgresql-cluster-rw -- bash
kubectl run pg-client --namespace default --image=postgres:18-alpine --restart=Never --rm -it --env PGPASSWORD=postgres -- bash

psql -h postgresql-cluster-rw.postgres.svc -p 5432 -U postgres

psql -h pg-postgresql -p 5432 -U postgres


\l
\du
\c urls
\dt

# cleanup
helmfile destroy

```



kubectl -n postgres delete pvc --all
kubectl -n postgres delete configmap --all
kubectl -n postgres delete secret --all

kubectl create namespace postgres --dry-run=client -o yaml | kubectl apply -f -

kubectl -n postgres create secret generic pg-creds \
--from-literal=admin-password='postgres' \
--from-literal=user-password='Auth123' \
--from-literal=replication-password='postgres' \
--dry-run=client -o yaml | kubectl apply -f -

kubectl -n postgres create configmap db-init-sql \
--from-file=init.sql=init.sql \
--dry-run=client -o yaml | kubectl apply -f -

kubectl get pv,pvc,secrets -n postgres
kubectl -n postgres get secret pg-creds -o jsonpath='{.data}' | jq -r 'to_entries[] | "\(.key): \(.value | @base64d)"'
kubectl -n postgres get configmap db-init-sql -o yaml

k -n postgres logs -l app.kubernetes.io/instance=pg -f

