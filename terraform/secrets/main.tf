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
