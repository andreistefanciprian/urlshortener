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

# Service account for external-dns managing the private Cloud DNS zone
resource "google_service_account" "external_dns" {
  account_id   = "${var.project_name}-external-dns"
  display_name = "Service account for external-dns (Cloud DNS private zone)"
  project      = var.gcp_project
  depends_on   = [google_project_service.iam]
}

resource "google_project_iam_member" "external_dns_admin" {
  project = var.gcp_project
  role    = "roles/dns.admin"
  member  = "serviceAccount:${google_service_account.external_dns.email}"
}

resource "google_service_account_iam_member" "external_dns_workload_identity" {
  service_account_id = google_service_account.external_dns.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "serviceAccount:${var.gcp_project}.svc.id.goog[external-dns/external-dns-clouddns]"
}
