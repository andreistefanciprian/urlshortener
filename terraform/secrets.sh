


PROJECT_NAME="home"
for secret in redis-password db-admin-password db-user-password db-replication-password; do
  echo -n "$(openssl rand -base64 32)" | gcloud secrets versions add ${PROJECT_NAME}-${secret} --data-file=-
done

# Verify that the secrets were created successfully and display their values
for secret in redis-password db-admin-password db-user-password db-replication-password; do
  value=$(gcloud secrets versions access latest --secret=${PROJECT_NAME}-${secret})
  echo "${PROJECT_NAME}-${secret}: ${value}"
done