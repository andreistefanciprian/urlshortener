
set -e

# ── Config ──────────────────────────────────────────────────────────────────
gcloud config set account $GCP_EMAIL
gcloud config set project $GCP_PROJECT
gcloud config set compute/region $GCP_REGION
gcloud config set compute/zone ${GCP_REGION}-a

gcloud config list

ADMIN_SA="terraform-admin@${GCP_PROJECT}.iam.gserviceaccount.com"

# ── Enable GCP APIs ──────────────────────────────────────────────────────────
gcloud services enable --project $GCP_PROJECT \
  cloudresourcemanager.googleapis.com \
  servicenetworking.googleapis.com \
  servicemanagement.googleapis.com \
  iamcredentials.googleapis.com \
  compute.googleapis.com \
  cloudkms.googleapis.com \
  secretmanager.googleapis.com \
  container.googleapis.com

# ── terraform-admin SA ───────────────────────────────────────────────────────
# Full infra permissions. Never used directly — impersonated by your user account.
# No key is ever downloaded for this SA.
if ! gcloud iam service-accounts describe $ADMIN_SA >/dev/null 2>&1; then
  gcloud iam service-accounts create terraform-admin \
    --description="Terraform admin — full infra permissions, impersonated only" \
    --display-name="Terraform Admin"
  echo "terraform-admin SA created."
else
  echo "terraform-admin SA already exists."
fi

# Roles scoped to what each terraform layer actually provisions:
#   networking/            → compute.networkAdmin (VPC, subnet, router, NAT, firewall), dns.admin (private DNS zone)
#   cloudflared_vm/        → compute.instanceAdmin.v1 (VM instance)
#   gke/                   → container.admin
#   secrets/               → secretmanager.admin, iam.serviceAccountAdmin
#   certificate_authority/ → privateca.admin
#   tf_bucket/             → storage.admin
#   all layers             → serviceusage.serviceUsageAdmin (enable APIs)
#                            resourcemanager.projectIamAdmin (grant roles to created SAs)
ADMIN_ROLES=(
  "roles/compute.networkAdmin"
  "roles/compute.instanceAdmin.v1"
  "roles/container.admin"
  "roles/secretmanager.admin"
  "roles/iam.serviceAccountAdmin"
  "roles/iam.serviceAccountKeyAdmin"
  "roles/iam.serviceAccountUser"
  "roles/resourcemanager.projectIamAdmin"
  "roles/privateca.admin"
  "roles/storage.admin"
  "roles/serviceusage.serviceUsageAdmin"
  "roles/compute.securityAdmin"  # grants compute.firewalls.create and related firewall permissions
  "roles/dns.admin"              # networking/ — create/manage Cloud DNS managed zones
)

for role in "${ADMIN_ROLES[@]}"; do
  gcloud projects add-iam-policy-binding $GCP_PROJECT \
    --member="serviceAccount:${ADMIN_SA}" \
    --role="$role"
done

# Allow your personal account to impersonate terraform-admin
gcloud iam service-accounts add-iam-policy-binding $ADMIN_SA \
  --member="user:${GCP_EMAIL}" \
  --role="roles/iam.serviceAccountTokenCreator"

echo ""
echo "Verifying roles assigned to $ADMIN_SA:"
gcloud projects get-iam-policy $GCP_PROJECT \
  --flatten="bindings[].members" \
  --filter="bindings.members:${ADMIN_SA}" \
  --format="table(bindings.role)"

echo ""
echo "Setup complete."
echo "  Admin SA: $ADMIN_SA (no key — impersonated only)"
echo ""
echo "To run Terraform locally:"
echo "  gcloud auth application-default login"
echo "  export GOOGLE_IMPERSONATE_SERVICE_ACCOUNT=$ADMIN_SA"
