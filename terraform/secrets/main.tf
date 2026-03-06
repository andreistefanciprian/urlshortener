locals {
  labels = merge({
    managed_by = "terraform"
    env        = var.project_name
  }, var.labels)

  # Add new secrets here as needed
  secrets = {
    cloudflared-tunnel-token = {
      description = "Cloudflare Tunnel token for cloudflared VM"
    }
    cloudflare-api-token = {
      description = "Cloudflare API token for External Secrets Operator"
    }
    redis-password = {
      description = "Redis password for External Secrets Operator"
    }
    db-admin-password = {
      description = "Database admin password for External Secrets Operator"
    }
    db-user-password = {
      description = "Database user password for External Secrets Operator"
    }
    db-replication-password = {
      description = "Database replication password for External Secrets Operator"
    }
  }
}

# Enable Secret Manager API
resource "google_project_service" "secretmanager" {
  service                    = "secretmanager.googleapis.com"
  disable_dependent_services = true
  disable_on_destroy         = false
}

# Wait for API to propagate after enablement
resource "time_sleep" "wait_for_secretmanager_api" {
  create_duration = "30s"
  depends_on      = [google_project_service.secretmanager]
}

# Secret shells — values are written by downstream layers
resource "google_secret_manager_secret" "secrets" {
  for_each  = local.secrets
  secret_id = "${var.project_name}-${each.key}"

  replication {
    auto {}
  }

  labels     = local.labels
  depends_on = [time_sleep.wait_for_secretmanager_api]
}

# Enable IAM API
resource "google_project_service" "iam" {
  service                    = "iam.googleapis.com"
  disable_dependent_services = true
  disable_on_destroy         = false
}

# Dedicated service account for cloudflared VM
resource "google_service_account" "cloudflared" {
  account_id   = "${var.project_name}-cloudflared"
  display_name = "Service account for cloudflared tunnel VM"
  project      = var.gcp_project
  depends_on   = [google_project_service.iam]
}

# Grant cloudflared SA access to its tunnel token secret only
resource "google_secret_manager_secret_iam_member" "cloudflared_accessor" {
  secret_id = google_secret_manager_secret.secrets["cloudflared-tunnel-token"].id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.cloudflared.email}"
}

# Dedicated service account for External Secrets Operator
resource "google_service_account" "external_secrets" {
  account_id   = "${var.project_name}-external-secrets"
  display_name = "Service account for External Secrets Operator"
  project      = var.gcp_project
  depends_on   = [google_project_service.iam]
}

# Grant ESO SA access to the Cloudflare API token secret
resource "google_secret_manager_secret_iam_member" "external_secrets_cloudflare_accessor" {
  secret_id = google_secret_manager_secret.secrets["cloudflare-api-token"].id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.external_secrets.email}"
}

# Grant ESO SA access to the Redis password secret
resource "google_secret_manager_secret_iam_member" "external_secrets_redis_accessor" {
  secret_id = google_secret_manager_secret.secrets["redis-password"].id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.external_secrets.email}"
}

# Grant ESO SA access to the DB secrets
resource "google_secret_manager_secret_iam_member" "external_secrets_db_admin_accessor" {
  secret_id = google_secret_manager_secret.secrets["db-admin-password"].id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.external_secrets.email}"
}

resource "google_secret_manager_secret_iam_member" "external_secrets_db_user_accessor" {
  secret_id = google_secret_manager_secret.secrets["db-user-password"].id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.external_secrets.email}"
}

resource "google_secret_manager_secret_iam_member" "external_secrets_db_replication_accessor" {
  secret_id = google_secret_manager_secret.secrets["db-replication-password"].id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.external_secrets.email}"
}

# Allow the k8s ESO service account to impersonate the GCP SA via Workload Identity
resource "google_service_account_iam_member" "external_secrets_workload_identity" {
  service_account_id = google_service_account.external_secrets.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "serviceAccount:${var.gcp_project}.svc.id.goog[external-secrets/external-secrets]"
}
