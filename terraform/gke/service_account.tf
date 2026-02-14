resource "google_project_service" "iam" {
  service                    = "iam.googleapis.com"
  disable_dependent_services = true
  disable_on_destroy         = false
}

resource "google_service_account" "cluster" {
  account_id   = "${var.project_name}-cluster"
  display_name = "${var.project_name}-cluster"
  project      = var.gcp_project
  depends_on   = [google_project_service.iam]
}

locals {
  cluster_service_account_roles = [
    "roles/logging.logWriter",
    "roles/monitoring.metricWriter",
    "roles/monitoring.viewer",
    "roles/stackdriver.resourceMetadata.writer",
    "roles/artifactregistry.reader"
  ]
}

resource "google_project_iam_member" "cluster" {
  for_each = toset(local.cluster_service_account_roles)
  project  = var.gcp_project
  role     = each.value
  member   = "serviceAccount:${google_service_account.cluster.email}"
}
