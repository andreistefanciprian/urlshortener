
set -e

# Set gcloud to use your project and region
gcloud config set project $GCP_PROJECT
gcloud config set account $GCP_EMAIL

# Enable necessary GCP APIs
gcloud services enable --project $GCP_PROJECT \
cloudresourcemanager.googleapis.com \
servicenetworking.googleapis.com \
servicemanagement.googleapis.com \
iamcredentials.googleapis.com \
compute.googleapis.com \
cloudkms.googleapis.com \
secretmanager.googleapis.com \
container.googleapis.com

# Set gcloud to use your region and zone
gcloud config set compute/region $GCP_REGION
gcloud config set compute/zone ${GCP_REGION}-a

# Check gcloud config
gcloud config list

# Create and set up Terraform service account
if ! gcloud iam service-accounts describe terraform@${GCP_PROJECT}.iam.gserviceaccount.com >/dev/null 2>&1; then
  gcloud iam service-accounts create terraform \
  --description="Used by Terraform" \
  --display-name="Terraform"
  echo "Terraform service account created."
else
  echo "Terraform service account already exists."
fi

# Verify service account was created
gcloud iam service-accounts list

# Grant the necessary roles to the service account
gcloud projects add-iam-policy-binding $GCP_PROJECT \
--member="serviceAccount:terraform@${GCP_PROJECT}.iam.gserviceaccount.com" \
--role="roles/container.admin" \
--role="roles/storage.admin" \
--role="roles/compute.admin" \
--role="roles/storage.legacyBucketWriter" \
--role="roles/owner"

# Describe service account
gcloud iam service-accounts describe terraform@${GCP_PROJECT}.iam.gserviceaccount.com

# Generate and download a service account key
gcloud iam service-accounts keys create gcp_sa_key.json \
--iam-account=terraform@${GCP_PROJECT}.iam.gserviceaccount.com

# Verify service account key was created
cat gcp_sa_key.json | grep project_id